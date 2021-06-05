package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	mcv "github.com/thegodmouse/url-shortener/converter/mock"
	"github.com/thegodmouse/url-shortener/db"
	"github.com/thegodmouse/url-shortener/dto"
	mr "github.com/thegodmouse/url-shortener/services/redirect/mock"
	ms "github.com/thegodmouse/url-shortener/services/shortener/mock"
)

func TestAPI(t *testing.T) {
	suite.Run(t, new(APITestSuite))
}

type APITestSuite struct {
	suite.Suite

	ctrl *gomock.Controller

	mockShortener *ms.MockService
	mockRedirect  *mr.MockService
	mockConv      *mcv.MockConverter
	hostname      string
}

func (s *APITestSuite) SetupSuite() {
	s.hostname = "localhost:5566"
	s.ctrl = gomock.NewController(s.T())
}

func (s *APITestSuite) SetupTest() {
	s.mockShortener = ms.NewMockService(s.ctrl)
	s.mockRedirect = mr.NewMockService(s.ctrl)
	s.mockConv = mcv.NewMockConverter(s.ctrl)
}

func (s *APITestSuite) TestCreateURL() {
	server := NewServer(s.hostname, s.mockShortener, s.mockRedirect, s.mockConv)

	url := "http://localhost:7788"
	expireAt := time.Now().Add(time.Minute).Round(time.Second)
	id := int64(12345)
	urlID := "12345"
	expectShortURL := s.hostname + "/" + urlID
	s.mockShortener.
		EXPECT().
		Shorten(gomock.Any(), gomock.Eq(url), gomock.Eq(expireAt)).
		Return(id, nil)
	s.mockConv.
		EXPECT().
		ConvertToShortURL(gomock.Eq(id)).
		Return(urlID, nil)

	// create test context
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(
		"POST", ShortenerPathV1, s.makeTestCreateURLRequestBody(url, expireAt))
	// SUT
	server.createURL(ctx)

	response := &dto.CreateURLResponse{}
	json.NewDecoder(w.Body).Decode(response)
	s.Equal(expectShortURL, response.ShortURL)
	s.Equal(http.StatusOK, w.Code)
}

func (s *APITestSuite) TestCreateURL_withBadRequest() {
	server := NewServer(s.hostname, s.mockShortener, s.mockRedirect, s.mockConv)

	testCases := []struct {
		body io.Reader
	}{
		{
			body: nil,
		},
		{
			body: s.makeTestCreateURLRequestBody(
				"http://localhost:5556",
				time.Now().Add(-time.Minute),
			),
		},
		{
			body: s.makeTestCreateURLRequestBody(
				"http://localhost:5566",
				time.Now().Add(-time.Minute),
			),
		},
	}
	for _, testCase := range testCases {
		// create test context
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("POST", ShortenerPathV1, testCase.body)
		// SUT
		server.createURL(ctx)

		s.Equal(http.StatusBadRequest, w.Code)
	}
}

func (s *APITestSuite) TestCreateURL_withShortenerError() {
	server := NewServer(s.hostname, s.mockShortener, s.mockRedirect, s.mockConv)

	url := "http://localhost:7788"
	expireAt := time.Now().Add(time.Minute).Round(time.Second)

	s.mockShortener.
		EXPECT().
		Shorten(gomock.Any(), gomock.Eq(url), gomock.Eq(expireAt)).
		Return(int64(0), errors.New("unknown shortener error"))

	// create test context
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(
		"POST", ShortenerPathV1, s.makeTestCreateURLRequestBody(url, expireAt))
	// SUT
	server.createURL(ctx)

	s.Equal(http.StatusInternalServerError, w.Code)
}

func (s *APITestSuite) makeTestCreateURLRequestBody(url string, expireAt time.Time) io.Reader {
	request := &dto.CreateURLRequest{
		URL:      url,
		ExpireAt: expireAt.Format(time.RFC3339),
	}
	buf := new(bytes.Buffer)
	json.NewEncoder(buf).Encode(request)
	return buf
}

func (s *APITestSuite) TestDeleteURL() {
	server := NewServer(s.hostname, s.mockShortener, s.mockRedirect, s.mockConv)

	id := int64(12345)
	urlID := "12345"

	s.mockConv.
		EXPECT().
		ConvertToID(gomock.Eq(urlID)).
		Return(id, nil)
	s.mockShortener.
		EXPECT().
		Delete(gomock.Any(), gomock.Eq(id)).
		Return(nil)

	// create test context
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest("DELETE", ShortenerPathV1, nil)
	ctx.Params = append(ctx.Params, gin.Param{Key: "url_id", Value: urlID})
	// SUT
	server.deleteURL(ctx)

	s.Equal(http.StatusNoContent, w.Code)
}

func (s *APITestSuite) TestDeleteURL_withShortenerError() {
	server := NewServer(s.hostname, s.mockShortener, s.mockRedirect, s.mockConv)

	testCases := []struct {
		id           int64
		urlID        string
		shortenerErr error
		expCode      int
	}{
		{
			id:           int64(123),
			urlID:        "123",
			shortenerErr: db.ErrNoRows,
			expCode:      http.StatusNotFound,
		},
		{
			id:           int64(789),
			urlID:        "789",
			shortenerErr: errors.New("unexpected error"),
			expCode:      http.StatusInternalServerError,
		},
	}
	for _, testCase := range testCases {
		s.mockConv.
			EXPECT().
			ConvertToID(gomock.Eq(testCase.urlID)).
			Return(testCase.id, nil)
		s.mockShortener.
			EXPECT().
			Delete(gomock.Any(), gomock.Eq(testCase.id)).
			Return(testCase.shortenerErr)

		// create test context
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("DELETE", ShortenerPathV1, nil)
		ctx.Params = append(ctx.Params, gin.Param{Key: "url_id", Value: testCase.urlID})
		// SUT
		server.deleteURL(ctx)

		s.Equal(testCase.expCode, w.Code)
	}
}

func (s *APITestSuite) TestRedirectURL() {
	server := NewServer(s.hostname, s.mockShortener, s.mockRedirect, s.mockConv)

	id := int64(12345)
	urlID := "12345"
	redirectURL := "http://localhost:7788"
	s.mockConv.
		EXPECT().
		ConvertToID(gomock.Eq(urlID)).
		Return(id, nil)
	s.mockRedirect.
		EXPECT().
		RedirectTo(gomock.Any(), gomock.Eq(id)).
		Return(redirectURL, nil)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest("GET", "/", nil)
	ctx.Params = append(ctx.Params, gin.Param{Key: "url_id", Value: urlID})
	// SUT
	server.redirectURL(ctx)

	s.Equal(redirectURL, w.Header().Get("location"))
	s.Equal(http.StatusSeeOther, w.Code)
}

func (s *APITestSuite) TestRedirectURL_withRedirectError() {
	server := NewServer(s.hostname, s.mockShortener, s.mockRedirect, s.mockConv)

	testCases := []struct {
		id          int64
		urlID       string
		redirectErr error
		expCode     int
	}{
		{
			id:          int64(123),
			urlID:       "123",
			redirectErr: db.ErrNoRows,
			expCode:     http.StatusNotFound,
		},
		{
			id:          int64(789),
			urlID:       "789",
			redirectErr: errors.New("unexpected error"),
			expCode:     http.StatusInternalServerError,
		},
	}
	for _, testCase := range testCases {
		s.mockConv.
			EXPECT().
			ConvertToID(gomock.Eq(testCase.urlID)).
			Return(testCase.id, nil)
		s.mockRedirect.
			EXPECT().
			RedirectTo(gomock.Any(), gomock.Eq(testCase.id)).
			Return("", testCase.redirectErr)

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("GET", "/", nil)
		ctx.Params = append(ctx.Params, gin.Param{Key: "url_id", Value: testCase.urlID})
		// SUT
		server.redirectURL(ctx)

		s.Equal(testCase.expCode, w.Code)
	}
}

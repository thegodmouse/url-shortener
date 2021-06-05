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
	"github.com/thegodmouse/url-shortener/converter"
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
	hostname      string
}

func (s *APITestSuite) SetupSuite() {
	s.hostname = "localhost:5566"
	s.ctrl = gomock.NewController(s.T())
}

func (s *APITestSuite) SetupTest() {
	s.mockShortener = ms.NewMockService(s.ctrl)
	s.mockRedirect = mr.NewMockService(s.ctrl)
}

func (s *APITestSuite) TestCreateURL() {
	server := NewServer(s.hostname, s.mockShortener, s.mockRedirect)

	url := "http://localhost:7788"
	expireAt := time.Now().Add(time.Minute).Round(time.Second)

	urlID := "12345"
	expectShortURL := s.hostname + "/" + urlID
	s.mockShortener.
		EXPECT().
		Shorten(gomock.Any(), gomock.Eq(url), gomock.Eq(expireAt)).
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
	server := NewServer(s.hostname, s.mockShortener, s.mockRedirect)

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
	server := NewServer(s.hostname, s.mockShortener, s.mockRedirect)

	url := "http://localhost:7788"
	expireAt := time.Now().Add(time.Minute).Round(time.Second)

	expErr := errors.New("unexpected error")
	s.mockShortener.
		EXPECT().
		Shorten(gomock.Any(), gomock.Eq(url), gomock.Eq(expireAt)).
		Return("", expErr)

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
	server := NewServer(s.hostname, s.mockShortener, s.mockRedirect)

	urlID := "12345"
	s.mockShortener.
		EXPECT().
		Delete(gomock.Any(), gomock.Eq(urlID)).
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
	server := NewServer(s.hostname, s.mockShortener, s.mockRedirect)

	testCases := []struct {
		urlID        string
		shortenerErr error
		expCode      int
	}{
		{
			urlID:        "123",
			shortenerErr: db.ErrNoRows,
			expCode:      http.StatusNotFound,
		},
		{
			urlID:        "456",
			shortenerErr: converter.ErrURLFormat,
			expCode:      http.StatusBadRequest,
		},
		{
			urlID:        "789",
			shortenerErr: errors.New("unexpected error"),
			expCode:      http.StatusInternalServerError,
		},
	}
	for _, testCase := range testCases {
		s.mockShortener.
			EXPECT().
			Delete(gomock.Any(), gomock.Eq(testCase.urlID)).
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
	server := NewServer(s.hostname, s.mockShortener, s.mockRedirect)

	urlID := "12345"
	redirectURL := "http://localhost:7788"
	s.mockRedirect.
		EXPECT().
		RedirectTo(gomock.Any(), gomock.Eq(urlID)).
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
	server := NewServer(s.hostname, s.mockShortener, s.mockRedirect)

	testCases := []struct {
		urlID       string
		redirectErr error
		expCode     int
	}{
		{
			urlID:       "123",
			redirectErr: db.ErrNoRows,
			expCode:     http.StatusNotFound,
		},
		{
			urlID:       "456",
			redirectErr: converter.ErrURLFormat,
			expCode:     http.StatusBadRequest,
		},
		{
			urlID:       "789",
			redirectErr: errors.New("unexpected error"),
			expCode:     http.StatusInternalServerError,
		},
	}
	for _, testCase := range testCases {
		s.mockRedirect.
			EXPECT().
			RedirectTo(gomock.Any(), gomock.Eq(testCase.urlID)).
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

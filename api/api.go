package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/thegodmouse/url-shortener/converter"
	"github.com/thegodmouse/url-shortener/db"
	"github.com/thegodmouse/url-shortener/dto"
	"github.com/thegodmouse/url-shortener/services/redirect"
	"github.com/thegodmouse/url-shortener/services/shortener"
)

const (
	ShortenerPathV1 = "/api/v1/urls"
)

func NewServer(hostname string, shortenSrv shortener.Service, redirectSrv redirect.Service) *Server {
	router := gin.Default()
	server := &Server{
		hostname:    hostname,
		router:      router,
		shortenSrv:  shortenSrv,
		redirectSrv: redirectSrv,
	}
	shortenerGroupV1 := router.Group(ShortenerPathV1)
	shortenerGroupV1.POST("/", server.createURL)
	shortenerGroupV1.DELETE("/:url_id", server.deleteURL)
	router.GET("/:url_id", server.redirectURL)
	return server
}

type Server struct {
	hostname    string
	shortenSrv  shortener.Service
	redirectSrv redirect.Service
	router      *gin.Engine
}

func (s *Server) Serve(addr string) error {
	log.Infof("url-shortener server is running at addr: %v", addr)
	return s.router.Run(addr)
}

func (s *Server) createURL(ctx *gin.Context) {
	var createURLRequest dto.CreateURLRequest
	if err := ctx.ShouldBindJSON(&createURLRequest); err != nil {
		log.Errorf("createURL: bad request format, err: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "bad request"})
		return
	}
	expireAt, err := time.Parse(time.RFC3339, createURLRequest.ExpireAt)
	if err != nil {
		log.Errorf("createURL: invalid time format for expireAt: %v, err: %v", expireAt, err)
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "invalid time format"})
		return
	}
	if expireAt.Before(time.Now()) {
		log.Errorf("createURL: expireAt was expired, expireAt: %v", expireAt)
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "expireAt is in the past"})
		return
	}
	var urlID string
	urlID, err = s.shortenSrv.Shorten(ctx, createURLRequest.URL, expireAt)
	if err != nil {
		log.Errorf("createURL: shorten url for request %+v, err: %v", createURLRequest, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
		return
	}
	log.Infof("createURL: generated short url: %v, request: %+v", urlID, createURLRequest)
	ctx.JSON(http.StatusOK, &dto.CreateURLResponse{
		ID:       urlID,
		ShortURL: fmt.Sprintf("%v/%v", s.hostname, urlID),
	})
}

// deleteURL deletes a short url in the db.
func (s *Server) deleteURL(ctx *gin.Context) {
	urlID := ctx.Param("url_id")
	if err := s.shortenSrv.Delete(ctx, urlID); err != nil {
		switch err {
		case converter.ErrURLFormat:
			log.Errorf("deleteURL: wrong format for url_id: %v", urlID)
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "url_id is in wrong format"})
		case db.ErrNoRows:
			log.Errorf("deleteURL: cannot find url_id: %v", urlID)
			ctx.JSON(http.StatusNotFound, gin.H{"message": "requested url_id not found"})
		default:
			log.Errorf("deleteURL: shorten url for url_id: %v, err: %v", urlID, err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
		}
		return
	}
	log.Infof("deleteURL: short url with id: %v has been successfully deleted", urlID)
	ctx.JSON(http.StatusNoContent, nil)
}

// redirectURL redirects a short url to its original url.
func (s *Server) redirectURL(ctx *gin.Context) {
	urlID := ctx.Param("url_id")
	location, err := s.redirectSrv.RedirectTo(ctx, urlID)
	if err != nil {
		switch err {
		case converter.ErrURLFormat:
			log.Errorf("redirectURL: wrong format for url_id: %v", urlID)
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "url_id is in wrong format"})
		case db.ErrNoRows:
			log.Errorf("redirectURL: cannot find url_id: %v", urlID)
			ctx.JSON(http.StatusNotFound, gin.H{"message": "requested url_id not found"})
		default:
			log.Errorf("redirectURL: shorten url for url_id: %v, err: %v", urlID, err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
		}
		return
	}
	log.Infof("redirectURL: short url with id: %v has been successfully redirected to %v", urlID, location)
	ctx.Redirect(http.StatusSeeOther, location)
}

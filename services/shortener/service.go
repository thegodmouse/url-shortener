package shortener

import "time"

type Service interface {
	Shorten(url string, expireAt time.Time) (string, error)
	Delete(urlID string) error
}

func NewService() Service {
	return &serviceImpl{}
}

type serviceImpl struct {

}

func (s *serviceImpl) Shorten(url string, expireAt time.Time) (string, error) {
	panic("implement me")
}

func (s *serviceImpl) Delete(urlID string) error {
	panic("implement me")
}




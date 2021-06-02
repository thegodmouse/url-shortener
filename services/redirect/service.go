package redirect

type Service interface {
	RedirectTo(urlID string) (string, error)
}

func NewService() Service {
	return &serviceImpl{}
}

type serviceImpl struct {
}

func (r *serviceImpl) RedirectTo(urlID string) (string, error) {
	return "https://www.google.com", nil
}
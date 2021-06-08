package dto

// CreateURLResponse defines the response format for creating shorten url.
type CreateURLResponse struct {
	ID       string `json:"id"`
	ShortURL string `json:"shortUrl"`
}

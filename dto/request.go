package dto

// CreateURLRequest defines the request format for creating shorten url.
type CreateURLRequest struct {
	URL      string `json:"url"`
	ExpireAt string `json:"expireAt"`
}

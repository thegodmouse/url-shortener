package dto

type CreateURLRequest struct {
	URL      string `json:"url"`
	ExpireAt string `json:"expireAt"`
}

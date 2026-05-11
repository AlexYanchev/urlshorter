package model

type URLItem struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type BatchURLItem struct {
	ShortURL    string
	OriginalURL string
}

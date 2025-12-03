package service

type ShortenerService interface {
	ShortenURL(originalURL string) (string, error)
	GetOriginalURL(shortURL string) (string, error)
	SaveBatch(shortURLs, originalURLs []string) error
	GenerateShortURL(originalURL string) (string, error)
}

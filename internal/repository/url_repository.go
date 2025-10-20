package repository

type URLRepository interface {
	Save(shortURL, originalURL string) error
	Get(shortURL string) (string, error)
}

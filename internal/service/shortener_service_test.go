package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"link-shortener/internal/repository"
)

func newTestService() *ShortenerServiceImpl {
	return NewShortenerService(repository.NewLocalRepository(), zap.NewNop().Sugar())
}

func TestShortenURL(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	t.Run("valid URL", func(t *testing.T) {
		short, err := svc.ShortenURL(ctx, "https://practicum.yandex.ru/", "user1")
		require.NoError(t, err)
		assert.Len(t, short, 8)
	})

	t.Run("same URL returns same short", func(t *testing.T) {
		s1, err := svc.ShortenURL(ctx, "https://practicum.yandex.ru/same", "user1")
		require.NoError(t, err)
		s2, err := svc.ShortenURL(ctx, "https://practicum.yandex.ru/same", "user1")
		require.NoError(t, err)
		assert.Equal(t, s1, s2)
	})

	t.Run("invalid URL", func(t *testing.T) {
		_, err := svc.ShortenURL(ctx, "not-a-url", "user1")
		assert.Error(t, err)
	})

	t.Run("empty URL", func(t *testing.T) {
		_, err := svc.ShortenURL(ctx, "", "user1")
		assert.Error(t, err)
	})
}

func TestGetOriginalURL(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	t.Run("found", func(t *testing.T) {
		short, err := svc.ShortenURL(ctx, "https://example.com/", "user1")
		require.NoError(t, err)

		original, err := svc.GetOriginalURL(ctx, short)
		require.NoError(t, err)
		assert.Equal(t, "https://example.com/", original)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := svc.GetOriginalURL(ctx, "notexist")
		assert.Error(t, err)
	})
}

func TestGenerateShortURL(t *testing.T) {
	svc := newTestService()

	t.Run("valid URL", func(t *testing.T) {
		short, err := svc.GenerateShortURL("https://example.com/")
		require.NoError(t, err)
		assert.Len(t, short, 8)
	})

	t.Run("invalid URL", func(t *testing.T) {
		_, err := svc.GenerateShortURL("not-a-url")
		assert.Error(t, err)
	})

	t.Run("deterministic", func(t *testing.T) {
		s1, _ := svc.GenerateShortURL("https://example.com/det")
		s2, _ := svc.GenerateShortURL("https://example.com/det")
		assert.Equal(t, s1, s2)
	})
}

func TestSaveBatch(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	shortURLs := []string{"abc12345", "def67890"}
	originalURLs := []string{"https://a.com/", "https://b.com/"}

	err := svc.SaveBatch(ctx, shortURLs, originalURLs, "user1")
	require.NoError(t, err)

	orig, err := svc.GetOriginalURL(ctx, "abc12345")
	require.NoError(t, err)
	assert.Equal(t, "https://a.com/", orig)
}

func TestGetURLsByUserID(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	_, err := svc.ShortenURL(ctx, "https://first.com/", "userA")
	require.NoError(t, err)
	_, err = svc.ShortenURL(ctx, "https://second.com/", "userA")
	require.NoError(t, err)
	_, err = svc.ShortenURL(ctx, "https://other.com/", "userB")
	require.NoError(t, err)

	urls, err := svc.GetURLsByUserID(ctx, "userA")
	require.NoError(t, err)
	assert.Len(t, urls, 2)
}

func TestDeleteURLsAsync(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	short, err := svc.ShortenURL(ctx, "https://todelete.com/", "userD")
	require.NoError(t, err)

	svc.DeleteURLsAsync(ctx, []string{short}, "userD")
}

func TestInternalGenerateShortURL(t *testing.T) {
	svc := newTestService()

	result := svc.generateShortURL("https://practicum.yandex.ru/")
	assert.Len(t, result, 8)

	result2 := svc.generateShortURL("https://practicum.yandex.ru/")
	assert.Equal(t, result, result2)

	result3 := svc.generateShortURL("https://other.com/")
	assert.NotEqual(t, result, result3)
}

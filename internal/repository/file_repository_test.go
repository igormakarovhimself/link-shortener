package repository

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileRepository_SaveAndGet(t *testing.T) {
	f, err := os.CreateTemp("", "repo-*.json")
	require.NoError(t, err)
	_ = f.Close()
	defer func() { _ = os.Remove(f.Name()) }()

	repo, err := NewFileRepository(f.Name())
	require.NoError(t, err)
	defer func() { _ = repo.Close() }()

	ctx := context.Background()
	err = repo.Save(ctx, "abc", "https://example.com", "user1")
	require.NoError(t, err)

	orig, err := repo.Get(ctx, "abc")
	require.NoError(t, err)
	assert.Equal(t, "https://example.com", orig)
}

func TestFileRepository_GetNotFound(t *testing.T) {
	f, err := os.CreateTemp("", "repo-*.json")
	require.NoError(t, err)
	_ = f.Close()
	defer func() { _ = os.Remove(f.Name()) }()

	repo, err := NewFileRepository(f.Name())
	require.NoError(t, err)
	defer func() { _ = repo.Close() }()

	_, err = repo.Get(context.Background(), "missing")
	assert.Error(t, err)
}

func TestFileRepository_GetDeleted(t *testing.T) {
	f, err := os.CreateTemp("", "repo-*.json")
	require.NoError(t, err)
	_ = f.Close()
	defer func() { _ = os.Remove(f.Name()) }()

	repo, err := NewFileRepository(f.Name())
	require.NoError(t, err)
	defer func() { _ = repo.Close() }()

	ctx := context.Background()
	require.NoError(t, repo.Save(ctx, "s1", "https://example.com", "u1"))
	require.NoError(t, repo.DeleteURLs(ctx, []string{"s1"}, "u1"))

	_, err = repo.Get(ctx, "s1")
	assert.ErrorIs(t, err, ErrURLDeleted)
}

func TestFileRepository_GetByOriginalURL(t *testing.T) {
	f, err := os.CreateTemp("", "repo-*.json")
	require.NoError(t, err)
	_ = f.Close()
	defer func() { _ = os.Remove(f.Name()) }()

	repo, err := NewFileRepository(f.Name())
	require.NoError(t, err)
	defer func() { _ = repo.Close() }()

	ctx := context.Background()
	require.NoError(t, repo.Save(ctx, "xyz", "https://orig.com", "u1"))

	short, err := repo.GetByOriginalURL(ctx, "https://orig.com")
	require.NoError(t, err)
	assert.Equal(t, "xyz", short)

	_, err = repo.GetByOriginalURL(ctx, "https://notfound.com")
	assert.Error(t, err)
}

func TestFileRepository_SaveBatch(t *testing.T) {
	f, err := os.CreateTemp("", "repo-*.json")
	require.NoError(t, err)
	_ = f.Close()
	defer func() { _ = os.Remove(f.Name()) }()

	repo, err := NewFileRepository(f.Name())
	require.NoError(t, err)
	defer func() { _ = repo.Close() }()

	ctx := context.Background()
	err = repo.SaveBatch(ctx, []string{"s1", "s2"}, []string{"https://a.com", "https://b.com"}, "u1")
	require.NoError(t, err)

	orig, err := repo.Get(ctx, "s1")
	require.NoError(t, err)
	assert.Equal(t, "https://a.com", orig)
}

func TestFileRepository_GetURLsByUserID(t *testing.T) {
	f, err := os.CreateTemp("", "repo-*.json")
	require.NoError(t, err)
	_ = f.Close()
	defer func() { _ = os.Remove(f.Name()) }()

	repo, err := NewFileRepository(f.Name())
	require.NoError(t, err)
	defer func() { _ = repo.Close() }()

	ctx := context.Background()
	require.NoError(t, repo.Save(ctx, "x1", "https://a.com", "userA"))
	require.NoError(t, repo.Save(ctx, "x2", "https://b.com", "userA"))

	urls, err := repo.GetURLsByUserID(ctx, "userA")
	require.NoError(t, err)
	assert.Len(t, urls, 2)

	urls, err = repo.GetURLsByUserID(ctx, "unknown")
	require.NoError(t, err)
	assert.Empty(t, urls)
}

func TestFileRepository_Ping(t *testing.T) {
	f, err := os.CreateTemp("", "repo-*.json")
	require.NoError(t, err)
	_ = f.Close()
	defer func() { _ = os.Remove(f.Name()) }()

	repo, err := NewFileRepository(f.Name())
	require.NoError(t, err)
	defer func() { _ = repo.Close() }()

	assert.NoError(t, repo.Ping(context.Background()))
}

func TestFileRepository_LoadExisting(t *testing.T) {
	f, err := os.CreateTemp("", "repo-*.json")
	require.NoError(t, err)
	_ = f.Close()
	defer func() { _ = os.Remove(f.Name()) }()

	repo, err := NewFileRepository(f.Name())
	require.NoError(t, err)

	ctx := context.Background()
	require.NoError(t, repo.Save(ctx, "persist1", "https://persist.com", "u1"))
	_ = repo.Close()

	repo2, err := NewFileRepository(f.Name())
	require.NoError(t, err)
	defer func() { _ = repo2.Close() }()

	orig, err := repo2.Get(ctx, "persist1")
	require.NoError(t, err)
	assert.Equal(t, "https://persist.com", orig)
}

package repository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalRepository_SaveAndGet(t *testing.T) {
	repo := NewLocalRepository()
	ctx := context.Background()

	err := repo.Save(ctx, "short1", "https://example.com/", "user1")
	require.NoError(t, err)

	orig, err := repo.Get(ctx, "short1")
	require.NoError(t, err)
	assert.Equal(t, "https://example.com/", orig)
}

func TestLocalRepository_GetNotFound(t *testing.T) {
	repo := NewLocalRepository()
	ctx := context.Background()

	_, err := repo.Get(ctx, "missing")
	assert.Error(t, err)
}

func TestLocalRepository_GetDeleted(t *testing.T) {
	repo := NewLocalRepository()
	ctx := context.Background()

	require.NoError(t, repo.Save(ctx, "s1", "https://example.com/", "u1"))
	require.NoError(t, repo.DeleteURLs(ctx, []string{"s1"}, "u1"))

	_, err := repo.Get(ctx, "s1")
	assert.ErrorIs(t, err, ErrURLDeleted)
}

func TestLocalRepository_GetByOriginalURL(t *testing.T) {
	repo := NewLocalRepository()
	ctx := context.Background()

	require.NoError(t, repo.Save(ctx, "abc", "https://orig.com/", "u1"))

	short, err := repo.GetByOriginalURL(ctx, "https://orig.com/")
	require.NoError(t, err)
	assert.Equal(t, "abc", short)

	_, err = repo.GetByOriginalURL(ctx, "https://notfound.com/")
	assert.Error(t, err)
}

func TestLocalRepository_SaveBatch(t *testing.T) {
	repo := NewLocalRepository()
	ctx := context.Background()

	err := repo.SaveBatch(ctx, []string{"s1", "s2"}, []string{"https://a.com/", "https://b.com/"}, "u1")
	require.NoError(t, err)

	orig, err := repo.Get(ctx, "s1")
	require.NoError(t, err)
	assert.Equal(t, "https://a.com/", orig)

	orig, err = repo.Get(ctx, "s2")
	require.NoError(t, err)
	assert.Equal(t, "https://b.com/", orig)
}

func TestLocalRepository_GetURLsByUserID(t *testing.T) {
	repo := NewLocalRepository()
	ctx := context.Background()

	require.NoError(t, repo.Save(ctx, "x1", "https://a.com/", "userA"))
	require.NoError(t, repo.Save(ctx, "x2", "https://b.com/", "userA"))
	require.NoError(t, repo.Save(ctx, "x3", "https://c.com/", "userB"))

	urls, err := repo.GetURLsByUserID(ctx, "userA")
	require.NoError(t, err)
	assert.Len(t, urls, 2)

	urls, err = repo.GetURLsByUserID(ctx, "userB")
	require.NoError(t, err)
	assert.Len(t, urls, 1)

	urls, err = repo.GetURLsByUserID(ctx, "unknown")
	require.NoError(t, err)
	assert.Empty(t, urls)
}

func TestLocalRepository_DeleteURLs_WrongUser(t *testing.T) {
	repo := NewLocalRepository()
	ctx := context.Background()

	require.NoError(t, repo.Save(ctx, "d1", "https://del.com/", "owner"))
	require.NoError(t, repo.DeleteURLs(ctx, []string{"d1"}, "notowner"))

	orig, err := repo.Get(ctx, "d1")
	require.NoError(t, err)
	assert.Equal(t, "https://del.com/", orig)
}

func TestLocalRepository_Ping(t *testing.T) {
	repo := NewLocalRepository()
	assert.NoError(t, repo.Ping(context.Background()))
}

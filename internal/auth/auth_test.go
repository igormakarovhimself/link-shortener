package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateUserID(t *testing.T) {
	id1, err := GenerateUserID()
	require.NoError(t, err)
	assert.Len(t, id1, 32)

	id2, err := GenerateUserID()
	require.NoError(t, err)
	assert.NotEqual(t, id1, id2)
}

func TestSign(t *testing.T) {
	sig1 := Sign("user1")
	sig2 := Sign("user1")
	assert.Equal(t, sig1, sig2)

	sig3 := Sign("user2")
	assert.NotEqual(t, sig1, sig3)
}

func TestVerify(t *testing.T) {
	userID := "test-user"
	sig := Sign(userID)

	assert.True(t, Verify(userID, sig))
	assert.False(t, Verify(userID, "invalidsig"))
	assert.False(t, Verify("other-user", sig))
}

func TestBuildAndParseCookieValue(t *testing.T) {
	userID := "cookie-user"
	cookie := BuildCookieValue(userID)

	parsed, valid := ParseCookieValue(cookie)
	assert.True(t, valid)
	assert.Equal(t, userID, parsed)
}

func TestParseCookieValue_Invalid(t *testing.T) {
	t.Run("no separator", func(t *testing.T) {
		_, valid := ParseCookieValue("noseparator")
		assert.False(t, valid)
	})

	t.Run("wrong signature", func(t *testing.T) {
		_, valid := ParseCookieValue("user:badsig")
		assert.False(t, valid)
	})

	t.Run("empty string", func(t *testing.T) {
		_, valid := ParseCookieValue("")
		assert.False(t, valid)
	})
}

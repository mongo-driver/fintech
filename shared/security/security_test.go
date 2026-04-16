package security

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPasswordHashing(t *testing.T) {
	hash, err := HashPassword("SecurePass123")
	require.NoError(t, err)
	require.NotEqual(t, "SecurePass123", hash)
	require.NoError(t, CheckPassword(hash, "SecurePass123"))
	require.Error(t, CheckPassword(hash, "wrong"))
}

func TestJWTGenerateAndParse(t *testing.T) {
	j := NewJWTManager("secret", time.Hour)
	token, err := j.Generate("u1", "user@example.com")
	require.NoError(t, err)
	require.NotEmpty(t, token)

	claims, err := j.Parse(token)
	require.NoError(t, err)
	require.Equal(t, "u1", claims.UserID)
	require.Equal(t, "user@example.com", claims.Email)
}

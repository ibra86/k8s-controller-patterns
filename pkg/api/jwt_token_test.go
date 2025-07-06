package api

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
)

func TestTokenHandler_IssuesValidJWT(t *testing.T) {
	JWTSecret = "test-secret"
	ctx := &fasthttp.RequestCtx{}
	TokenHandler(ctx)
	resp := string(ctx.Response.Body())
	require.Contains(t, resp, "token")

	var parsed struct {
		Token string `json:"token"`
	}
	err := json.Unmarshal([]byte(resp), &parsed)
	require.NoError(t, err, "Failed to parse JSON from response")
	parsedToken, err := jwt.Parse(parsed.Token, func(token *jwt.Token) (interface{}, error) {
		return []byte(JWTSecret), nil
	})

	require.NoError(t, err)
	require.True(t, parsedToken.Valid)
}

func TestJWTMiddleware_ValidAndInvalidToken(t *testing.T) {
	JWTSecret = "test-secret"
	claims := jwt.MapClaims{
		"sub": "test-user",
		"exp": time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(JWTSecret))
	require.NoError(t, err)

	ctx := &fasthttp.RequestCtx{}
	ctx.Request.Header.Set("Authorization", "Bearer "+tokenStr)
	called := false
	mw := JWTMiddleware(func(ctx *fasthttp.RequestCtx) { called = true })
	mw(ctx)
	require.True(t, called, "middleware should call next handler with valid token")

	//invalid token
	ctx2 := &fasthttp.RequestCtx{}
	ctx2.Request.Header.Set("Authorization", "Bearer invalid-token")
	called = false
	mw(ctx2)
	require.False(t, called, "middleware should not call next handler with invalid token")
	require.Equal(t, fasthttp.StatusUnauthorized, ctx2.Response.StatusCode(), "should return 401 Unauthorized for invalid token")

	expiredClaims := jwt.MapClaims{
		"sub": "test-user",
		"exp": time.Now().Add(-time.Hour).Unix(), // expired token
	}
	expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims)
	expiredStr, err := expiredToken.SignedString([]byte(JWTSecret))
	require.NoError(t, err)
	ctx3 := &fasthttp.RequestCtx{}
	ctx3.Request.Header.Set("Authorization", "Bearer "+expiredStr)
	called = false
	mw(ctx3)
	require.False(t, called, "middleware should not call next handler with expired token")
	require.Equal(t, fasthttp.StatusUnauthorized, ctx3.Response.StatusCode(), "should return 401 Unauthorized for expired token")

}

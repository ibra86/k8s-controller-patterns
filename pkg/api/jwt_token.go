package api

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/valyala/fasthttp"
)

func TokenHandler(ctx *fasthttp.RequestCtx) {
	claims := jwt.MapClaims{
		"sub": "testuser",
		"exp": time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(JWTSecret))
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetBodyString(`{"error": "Failed to generate token"}`)
		return
	}
	ctx.SetContentType("application/json")
	ctx.SetBodyString(`{"token": "` + tokenString + `"}`)
}

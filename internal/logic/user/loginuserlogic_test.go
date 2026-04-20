package user

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

func TestGetJwtTokenStoresUserIDAsString(t *testing.T) {
	t.Parallel()

	const (
		secret = "jwt-secret"
		userID = int64(1921565896585154562)
	)

	logic := &LoginUserLogic{}
	tokenString, err := logic.getJwtToken(secret, time.Now().Unix(), 3600, userID)
	if err != nil {
		t.Fatalf("getJwtToken() error = %v", err)
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		t.Fatalf("jwt.Parse() error = %v", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatalf("token.Claims type = %T, want jwt.MapClaims", token.Claims)
	}

	got, ok := claims["userId"].(string)
	if !ok {
		t.Fatalf("claims[userId] type = %T, want string", claims["userId"])
	}
	if got != "1921565896585154562" {
		t.Fatalf("claims[userId] = %q, want %q", got, "1921565896585154562")
	}
}

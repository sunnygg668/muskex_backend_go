package metamask

import (
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"muskex/mmlogin/domain"
)

type token struct {
	*jwt.Token
}

func newToken(address domain.Address, d time.Duration) *token {
	return &token{jwt.NewWithClaims(
		jwt.SigningMethodHS256, newClaims(address, d),
	)}
}

func (token *token) signedString(secret string) (string, error) {
	return token.Token.SignedString([]byte(secret))
}

func (token *token) signedBytes(secret string) ([]byte, error) {
	str, err := token.signedString(secret)
	if err != nil {
		return nil, err
	}

	return []byte(str), nil
}

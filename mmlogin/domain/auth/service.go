package auth

import "muskex/mmlogin/domain"

type Service interface {
	SetUpChallenge(u *domain.User) error
	VerifyResponse(u *domain.User, responseBytes []byte) error
	IssueToken(u *domain.User) ([]byte, error)
}

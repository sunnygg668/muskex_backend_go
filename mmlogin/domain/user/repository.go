package user

import (
	"context"

	"muskex/mmlogin/domain"
)

type Repository interface {
	Add(ctx context.Context, u *domain.User) error
	Get(ctx context.Context, address domain.Address) (*domain.User, error)
	Update(ctx context.Context, u *domain.User) error
	Delete(ctx context.Context, u *domain.User) error
}

package application

import (
	"muskex/mmlogin/domain/auth"
	"muskex/mmlogin/domain/user"
)

type Core struct {
	Config       *Config
	Services     *Services
	Repositories *Repositories
}

type Services struct {
	Auth auth.Service
}

type Repositories struct {
	User user.Repository
}

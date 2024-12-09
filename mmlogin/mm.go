package mmlogin

import (
	"muskex/mmlogin/application"
	appAuth "muskex/mmlogin/application/auth"
	appUser "muskex/mmlogin/application/user"
	"muskex/mmlogin/infrastructure/auth/metamask"
	cacheUser "muskex/mmlogin/infrastructure/cache/user"
)

func newAppCore(conf *application.Config) *application.Core {
	return &application.Core{
		Services: &application.Services{
			Auth: metamask.NewServiceWithOutToken(),
		},
		Repositories: &application.Repositories{
			User: cacheUser.NewRepository(),
		},
	}
}

type apps struct {
	Auth appAuth.Application
	User appUser.Application
}

var Apps *apps

func InitMMLogin() {
	appCore := newAppCore(nil)
	Apps = &apps{
		appAuth.NewApplication(appCore),
		appUser.NewApplication(appCore),
	}
}

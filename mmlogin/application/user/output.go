package user

import "muskex/mmlogin/domain"

type GetUserOutput struct {
	Name       string `json:"name"`
	AddressHex string `json:"address"`
}

func NewGetUserOutput(u *domain.User) *GetUserOutput {
	return &GetUserOutput{
		Name:       u.Name,
		AddressHex: u.Address.Hex(),
	}
}

type UpdateUserOutput struct{}

func NewUpdateUserOutput() *UpdateUserOutput {
	return &UpdateUserOutput{}
}

type DeleteUserOutput struct{}

func NewDeleteUserOutput() *DeleteUserOutput {
	return &DeleteUserOutput{}
}

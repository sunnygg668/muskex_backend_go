package rserver

import (
	"muskex/utils"
)

type Helper struct {
}

func (h *Helper) Numeric(number int) string {
	str, _ := utils.Build("numeric", number)
	return str
}

func (h *Helper) Alnum(number int) string {
	str, _ := utils.Build("alnum", number)
	return str
}

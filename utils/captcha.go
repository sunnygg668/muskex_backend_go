package utils

import (
	"github.com/mojocn/base64Captcha"
	"log"
)

type captchTool struct {
	captchaDrive    *base64Captcha.DriverDigit
	captchaInstance *base64Captcha.Captcha
	captchaStore    base64Captcha.Store
}

var MobileCaptchTool *captchTool

//var PwdCaptchTool *captchTool

func InitCapcha() {
	MobileCaptchTool = NewCaptchTool()
	//PwdCaptchTool = NewCaptchTool()
}
func NewCaptchTool() *captchTool {
	ct := &captchTool{}
	ct.captchaDrive = &base64Captcha.DriverDigit{Length: 6,
		Height:   80,
		Width:    240,
		MaxSkew:  0.7,
		DotCount: 100}
	//ct.captchaStore = base64Captcha.DefaultMemStore
	ct.captchaStore = base64Captcha.NewMemoryStore(2000, base64Captcha.Expiration)
	ct.captchaInstance = base64Captcha.NewCaptcha(ct.captchaDrive, ct.captchaStore)
	return ct
}
func (ct *captchTool) GenCaptcha(onlyAnwser bool, bindId string) (id, answer, b64sImg string) {
	//id, b64s, err := captchaInstance.Generate()
	var content string
	id, content, answer = ct.captchaDrive.GenerateIdQuestionAnswer()
	if bindId != "" {
		id = bindId
	}
	err := ct.captchaStore.Set(id, answer)
	if err != nil {
		log.Fatalln(err)
	}
	if !onlyAnwser {

		item, err := ct.captchaDrive.DrawCaptcha(content)
		if err != nil {
			log.Fatalln(err)
		}
		b64sImg = item.EncodeB64string()
	}
	return
}
func (ct *captchTool) VerifyCaptcha(id, value string) bool {
	if id == "" {
		return false
	}
	return ct.captchaInstance.Verify(id, value, true)
}

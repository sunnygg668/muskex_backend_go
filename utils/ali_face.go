package utils

import (
	cloudauth20190307 "github.com/alibabacloud-go/cloudauth-20190307/client"
	openapi "github.com/alibabacloud-go/tea-rpc/client"
	"github.com/alibabacloud-go/tea/tea"

	teautil "github.com/alibabacloud-go/tea-utils/service"
	"sync"
)

// docs https://api.aliyun.com/document/Cloudauth/
//
//	var faceConfig=map[string]string{
//		"access_key_id":     "LTAI5tBgsBvrYXEfbABNFcRC",
//		"access_key_secret": "EizVAB7qlvMyK5SpeHbdb7n7lGcVvO",
//		"scene_id":          "1000011016",
//		"outer_order_no":    "4f9be124a50b4b9b9780e550938ba06b",
//		"callback_url":      "https://muskex2.cbnftcoin.vip/index.php/api/certification/callback",
//	}
var fcfg = faceConfig{
	AccessKeyId:     tea.String("LTAI5tBgsBvrYXEfbABNFcRC"),
	AccessKeySecret: tea.String("EizVAB7qlvMyK5SpeHbdb7n7lGcVvO"),
	SceneId:         tea.Int64(1000011016),
	OuterOrderNo:    tea.String("4f9be124a50b4b9b9780e550938ba06b"),
	CallbackUrl:     tea.String("https://muskex2.cbnftcoin.vip/index.php/api/certification/callback"),
}

type faceConfig struct {
	AccessKeyId     *string `json:"accessKeyId"`
	AccessKeySecret *string `json:"accessKeySecret"`
	RegionId        *string `json:"regionId"`
	SceneId         *int64  `json:"sceneId"`
	OuterOrderNo    *string `json:"outerOrderNo"`
	CallbackUrl     *string `json:"callbackUrl"`
}

var client *cloudauth20190307.Client
var tmpOnce sync.Once

func InitFaceVerify(name, no, returnUrl, metaInfo string) (res *cloudauth20190307.InitFaceVerifyResponse, _err error) {
	// 可用区域Id （请自行配置）
	regionId := tea.String("")
	tmpOnce.Do(func() {
		config := &openapi.Config{}
		config.AccessKeyId = fcfg.AccessKeyId
		// 您的AccessKey ID
		config.AccessKeySecret = fcfg.AccessKeySecret
		// 您的AccessKey Secret
		config.RegionId = regionId
		// 您的可用区ID
		client, _err = cloudauth20190307.NewClient(config)
		if _err != nil {
			panic(_err)
		}
	})
	// 认证场景ID。您必须先在智能核身控制台创建认证场景，才能获得认证场景ID。
	//sceneId := tea.Int(10000000000000)
	sceneId := fcfg.SceneId
	// 客户服务端自定义的业务唯一标识，用于后续定位排查问题使用。值最长为32位长度的字母数字组合，请确保唯一。
	//outerOrderNo := tea.String("e0c34a77f5ac40a5aa5e6ed20c35xxxx")
	outerOrderNo := fcfg.OuterOrderNo
	// 认证方案。唯一取值：LR_FR。
	productCode := tea.String("LR_FR")
	// 活体检测类型。取值：LIVENESS（默认）：动作活体检测 | PHOTINUS_LIVENESS：动作活体+炫彩活体双重检测
	model := tea.String("LIVENESS")
	// 证件类型。取值：IDENTITY_CARD，表示身份证。
	certType := tea.String("IDENTITY_CARD")
	// 用户的真实姓名。
	certName := tea.String(name)
	// 用户的证件号码。
	certNo := tea.String(no)
	// MetaInfo环境参数，需要通过客户端SDK获取。
	//metaInfo := tea.String("{"zimVer":"3.0.0","appVersion":"1","bioMetaInfo":"4.1.0:11501568,0","appName":"com.aliyun.antcloudauth","deviceType":"ios","osVersion":"iOS10.3.2","apdidToken":"","deviceModel":"iPhone9,1"}")

	requestInitFaceVerify := &cloudauth20190307.InitFaceVerifyRequest{}
	requestInitFaceVerify.SceneId = sceneId
	requestInitFaceVerify.OuterOrderNo = outerOrderNo
	requestInitFaceVerify.ProductCode = productCode
	requestInitFaceVerify.Model = model
	requestInitFaceVerify.CertType = certType
	requestInitFaceVerify.CertName = certName
	requestInitFaceVerify.CertNo = certNo
	//requestInitFaceVerify.CallbackUrl = CallbackUrl
	requestInitFaceVerify.MetaInfo = tea.String(metaInfo)
	requestInitFaceVerify.UserId = tea.String("1234")
	requestInitFaceVerify.ReturnUrl = tea.String(returnUrl)
	responseInitFaceVerify, _err := client.InitFaceVerify(requestInitFaceVerify, new(teautil.RuntimeOptions))
	//console.Log(util.ToJSONString(util.ToMap(responseInitFaceVerify)))
	return responseInitFaceVerify, _err
}
func GetAliFaceVerifyRes(certifyId string) (*cloudauth20190307.DescribeFaceVerifyResponse, error) {
	requestDescribeFaceVerify := &cloudauth20190307.DescribeFaceVerifyRequest{}
	requestDescribeFaceVerify.SceneId = fcfg.SceneId
	requestDescribeFaceVerify.CertifyId = tea.String(certifyId)
	responseDescribeFaceVerify, _err := client.DescribeFaceVerify(requestDescribeFaceVerify, new(teautil.RuntimeOptions))
	if _err != nil {
		return nil, _err
	}
	//console.Log(util.ToJSONString(util.ToMap(responseDescribeFaceVerify)))
	return responseDescribeFaceVerify, nil
}

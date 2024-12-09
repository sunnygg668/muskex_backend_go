package rserver

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"io/ioutil"
	"log"
	"muskex/gen/mproto"
	"muskex/gen/mproto/model"
	"muskex/utils"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// sendSmsConfig holds the static configuration for the request
var sendSmsConfig struct {
	smsAccount string
	smsPswd    string
	url        string
}

// sendSmsRequestBody represents the body of the send Sms request
type sendSmsMessage struct {
	template string            `json:"template"`
	content  string            `json:"content"`
	data     map[string]string `json:"data"`
}

// SendSmsResponse represents the response from the send Sms API
type SendSmsResponse struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}

// InitializeSendSmsReqConfig initializes the static configuration
func InitializeSendSmsReqConfig(smsAccount string, smsPswd string, url string) {
	sendSmsConfig.smsAccount = smsAccount
	sendSmsConfig.smsPswd = smsPswd
	sendSmsConfig.url = url
}

// GenerateSmsMD5Signature generates the MD5 signature for the request
func GenerateSmsMD5Signature(smsAccount, smsPswd, timestamp string) string {
	signatureString := smsAccount + smsPswd + timestamp
	hash := md5.Sum([]byte(signatureString))
	return hex.EncodeToString(hash[:])
}

// SendSms sends a request to send a new Sms
func SendSms(mobile string, message *sendSmsMessage) (*SendSmsResponse, error) {
	timestamp := time.Unix(time.Now().Unix(), 0).Format("20060102150405")
	sign := GenerateSmsMD5Signature(sendSmsConfig.smsAccount, sendSmsConfig.smsPswd, timestamp)

	params := make(url.Values)
	params.Set("account", sendSmsConfig.smsAccount)
	params.Set("ts", timestamp)
	params.Set("pswd", sign)
	params.Set("mobile", mobile)
	params.Set("msg", message.content)
	params.Set("needstatus", "true")

	/*非文档上的POST请求
	formDataStr := []byte(params.Encode())
	formDataByte := bytes.NewBuffer(formDataStr)

	requst, err := http.NewRequest(http.MethodPost, sendSmsConfig.url, formDataByte)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(requst)
	if err != nil {
		return nil, err
	}*/
	parseURL, _ := url.Parse(sendSmsConfig.url)
	parseURL.RawQuery = params.Encode()
	urlPathWithParams := parseURL.String()
	resp, err := http.Get(urlPathWithParams)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to send Sms")
	}

	responseBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	responseStr := string(responseBytes)
	responseAr := strings.Split(responseStr, "\n")
	responseArr := strings.Split(responseAr[0], ",")
	if len(responseArr) != 2 {
		return nil, errors.New("请求异常，请重试")
	}

	// Error codes and their corresponding messages
	errorMessages := map[string]string{
		"0":   "提交成功",
		"101": "无此用户",
		"102": "密码错",
		"103": "提交过快（提交速度超过流速限制）",
		"104": "系统忙（因平台侧原因，暂时无法处理提交的短信）",
		"105": "敏感短信（短信内容包含敏感词）",
		"106": "消息长度错（>700或<=0）",
		"107": "包含错误的手机号码",
		"108": "手机号码个数错（群发>50000或<=0;单发>200或<=0）",
		"109": "无发送额度（该用户可用短信数已使用完）",
		"110": "不在发送时间内",
		"111": "超出该账户当月发送额度限制",
		"112": "无此产品，用户没有订购该产品",
		"113": "extno格式错（非数字或者长度不对）",
		"115": "自动审核驳回",
		"116": "签名不合法，未带签名（用户必须带签名的前提下）",
		"117": "IP地址认证错,请求调用的IP地址不是系统登记的IP地址",
		"118": "用户没有相应的发送权限",
		"119": "用户已过期",
		"120": "内容不在白名单模板中",
	}

	// Check the response code and return an error if it's not 200
	code := responseArr[1]
	if code != "0" {
		if msg, exists := errorMessages[code]; exists {
			return nil, errors.New(msg)
		}
		return nil, fmt.Errorf("unknown error code: %s %s", responseArr[1], responseArr[0])
	} else {
		var sendSmsResponse SendSmsResponse
		sendSmsResponse.Code = code
		sendSmsResponse.Message = "msgid:" + responseAr[1]
		return &sendSmsResponse, nil
	}
}

type Variable struct {
	Id    int64
	Name  string
	Type  string
	Value string
}

func AnalysisVariable(content string, variableIds string, keyId string, tplVar map[string]string) map[string]interface{} {
	if content == "" && variableIds == "" {
		return map[string]interface{}{
			"content":   "",
			"variables": map[string]string{},
		}
	}

	// 读取数据库中的模板变量
	allVariable := make(map[string]Variable) // 全部变量
	useVariable := make(map[string]string)   // 使用到的变量
	variableTmp := []*mproto.SmsVariable{}
	err := utils.Orm.Find(&variableTmp, "status=1").Error
	if err != nil {
		return map[string]interface{}{}
	}
	variables := strings.Split(variableIds, ",") // 要分析的变量数组

	for _, value := range variableTmp {
		varName := value.Name
		// 是需要分析的变量
		if Contains(variables, fmt.Sprint(value.Id)) {
			value.Type = "definition_now" // 标记为现在定义
			if val, ok := tplVar[varName]; ok {
				value.Value = val
			} else {
				value.Value = CalcVar(value.Id, keyId)
			}
			useVariable[varName] = value.Value
			allVariable[varName] = Variable{
				Id:    value.Id,
				Name:  value.Name,
				Type:  value.Type,
				Value: value.Value,
			}
			continue
		}

		value.Type = "predefined" // 标记为预定义
		allVariable[varName] = Variable{
			Id:    value.Id,
			Name:  value.Name,
			Type:  value.Type,
			Value: value.Value,
		}
	}

	// 准备传递过来的模板变量
	for key, value := range tplVar {
		allVariable[key] = Variable{
			Type:  "definition_now", // 标记为现在定义
			Value: value,
		}
	}

	// 正则匹配模板中的变量
	re := regexp.MustCompile(`\${(.*?)}`)
	matches := re.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) > 1 {
			varName := match[1]
			if variable, ok := allVariable[varName]; ok {
				if variable.Type == "definition_now" {
					content = strings.Replace(content, match[0], variable.Value, 1)
					useVariable[varName] = variable.Value
				} else if variable.Type == "predefined" {
					variableValue := CalcVar(int64(variable.Id), keyId)
					content = strings.Replace(content, match[0], variableValue, 1)
					useVariable[varName] = variableValue
				}
			}
		}
	}

	return map[string]interface{}{
		"content":   content,
		"variables": useVariable,
	}
}

// 判断字符串切片中是否包含某个元素
func Contains(slice []string, item string) bool {
	for _, a := range slice {
		if a == item {
			return true
		}
	}
	return false
}

func CalcVar(id int64, keyId string) string {

	_, code, _ := utils.MobileCaptchTool.GenCaptcha(true, keyId)
	return code

	// 获取变量数据
	varData := &model.SmsVariable{}
	err := utils.Orm.First(varData, "id=?", id).Error
	if err != nil && err == gorm.ErrRecordNotFound {
		return ""
	}

	// 根据 value_source 的不同处理逻辑
	switch varData.ValueSource {
	case "literal":
		// 如果是字面量，直接返回值
		return varData.Value
	case "func":
		// 动态调用类的方法 go中没有类
		className := fmt.Sprintf("%s.%s", varData.Namespace, varData.Class)
		if className != "modules\\sms\\library.Helper" {
			return ""
		}
		h := &Helper{}

		// 使用反射调用方法  注意：方法必须首字母大写才能映射成功   建议所有方法都首字母大写
		method := reflect.ValueOf(h).MethodByName(FirstUpper(varData.Func))
		if !method.IsValid() {
			return ""
		}
		// 调用方法，传递参数
		intParam, _ := strconv.Atoi(varData.Param)
		results := method.Call([]reflect.Value{reflect.ValueOf(intParam)})
		if len(results) > 0 {
			return results[0].String()
		}

	case "sql":
		// 如果是 SQL，执行查询
		sqlQuery := strings.Replace(varData.Sql, "__PREFIX__", "ba_", -1) // 替换前缀

		res := []struct{}{}
		err := utils.Orm.Raw(sqlQuery).Scan(res).Error
		if err != nil {
			log.Println(err)
			return ""
		}
		if len(res) > 0 {
			for _, value := range res {
				return fmt.Sprintf("%v", value)
			}
		}
	}

	return ""
}

// 生成保存验证数据
func TemplateAnalysisAfter(template map[string]interface{}, msg map[string]string) {
	if variables, ok := template["variables"].(map[string]string); ok {
		//遍历 variables 中的键值对
		if code, exists := variables["code"]; exists {
			CaptchaCreate(msg["mobile"]+msg["template_code"], code)
		}
		if alnum, exists := variables["alnum"]; exists {
			CaptchaCreate(msg["mobile"]+msg["template_code"], alnum)
		}
	}
}

// 保存到数据库，用于验证
func CaptchaCreate(id string, variable string) {

}

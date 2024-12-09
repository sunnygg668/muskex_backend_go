package rserver

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	"log"
	"muskex/gen/mproto"
	"muskex/gen/mproto/model"
	"muskex/utils"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func LoadHomeIndexFromCache() (*mproto.IndexResonse, error) {
	ckey := "dbconfig"
	res, err := utils.CacheFromLru(1, ckey, 2000,
		func() (interface{}, error) {
			return loadHomeIndex()
		})

	if err != nil {
		log.Println("loadHomeIndexFromCache err", err)
	}
	return res.(*mproto.IndexResonse), nil
}
func loadHomeIndex() (*mproto.IndexResonse, error) {
	coinItems := []*model.Coin{}
	if err := utils.Orm.Order("weigh desc").Find(&coinItems, model.Coin{Status: 1, HomeRecommend: 1}).Error; err != nil {
		return nil, err
	}
	notices := []*model.MarketNotice{}
	if err := utils.Orm.Order("is_top desc, id desc").Limit(5).Find(&notices, "release_time<=?", time.Now().Unix()).Error; err != nil {
		return nil, err
	}
	news := []*model.MarketNews{}
	if err := utils.Orm.Order("id desc").Limit(3).Find(&news, "release_time<=?", time.Now().Unix()).Error; err != nil {
		return nil, err
	}
	for _, marketNews := range news {
		marketNews.Content = marketNews.Content[:60]
	}
	cfgs := LoadDbConfig().Names
	res := &mproto.IndexResonse{
		CoinList:            coinItems,
		NoticeList:          notices,
		NewsList:            news,
		AppVersion:          cfgs["app_version"].Value,
		AndroidDownloadUrl:  cfgs["android_download_url"].Value,
		IosDownloadUrl:      cfgs["ios_download_url"].Value,
		WgtDownloadUrl:      cfgs["wgt_download_url"].Value,
		AppVersionDesc:      cfgs["app_version_desc"].Value,
		OpenFaceRecognition: cfgs["open_face_recognition"].Value,
		CustomerServiceLink: cfgs["customer_service_link"].Value,
		InviteRegisterRule:  cfgs["invite_register_rule"].Value,
	}
	return res, nil
}

type dbConfig struct {
	Names      map[string]*model.Config
	NameValues map[string]string
	Groups     map[string][]*model.Config
}

func LoadDbConfig() *dbConfig {
	ckey := "dbconfig"
	res, _ := utils.CacheFromLru(1, ckey, 3600, func() (interface{}, error) {
		cfs := []*model.Config{}
		err := utils.Orm.Find(&cfs).Error
		if err != nil {
			log.Println("load loadDbConfig err", err)
			panic(err)
		}
		var namesCfg = map[string]*model.Config{}
		var nameValues = map[string]string{}
		var groupsCfg = map[string][]*model.Config{}
		for _, cf := range cfs {
			namesCfg[cf.Name] = cf
			nameValues[cf.Name] = cf.Value
			groupsCfg[cf.Group] = append(groupsCfg[cf.Group], cf)
		}
		return &dbConfig{Names: namesCfg, NameValues: nameValues, Groups: groupsCfg}, nil
	})
	return res.(*dbConfig)
}
func GetCfgValueInt64(name string) int64 {
	taskGiveCoinType, _ := strconv.Atoi(GetCfgValue(name))
	return int64(taskGiveCoinType)
}
func GetCfgValueF64(name string) float64 {
	taskGiveCoinType, _ := strconv.ParseFloat(GetCfgValue(name), 64)
	return float64(taskGiveCoinType)
}
func GetCfgValue(name string) string {
	return LoadDbConfig().NameValues[name]
}
func IsUserActive(userId int64) bool {
	cfg := LoadDbConfig().NameValues
	interval, _ := strconv.Atoi(cfg["user_activation_calc_interval"])
	beginTime := time.Now().Unix() - int64(interval*3600)
	sqlstr := `
select COALESCE(max(1), 0) exist
from dual
where exists(select id from ba_trade_contract_order where user_id = @userId and buy_time > @beginTime)
   or exists(select id from ba_trade_management_order where user_id = @userId and create_time > @beginTime)
   or exists(select id from ba_miners_order where user_id = @userId and create_time > @beginTime)
;`
	exits := 0
	utils.Orm.Raw(sqlstr, map[string]interface{}{"userId": userId, "beginTime": beginTime}).Scan(&exits)
	return exits == 1
}

func GetCtxUserId(ctx context.Context) int64 {
	uid, ok := ctx.Value("userId").(int64)
	if !ok {
		panic("token err")
	} else {
		return uid
	}
}
func GetCtxUserIdStr(ctx context.Context) string {
	return strconv.Itoa(int(GetCtxUserId(ctx)))
}

//	func GetCtxPid(ctx context.Context) int {
//		uid, ok := ctx.Value("pid").(int)
//		if !ok {
//			return 1
//		} else {
//			return uid
//		}
//	}
func Convert2UserInfo(user *model.User) *mproto.UserInfo {
	idcard := ""
	if len(user.Idcard) > 15 {
		idcard = user.Idcard[0:5] + "**********" + user.Idcard[15:]
	} else {
		idcard = user.Idcard
	}
	uinfo := mproto.UserInfo{
		Id:           int64(user.Id),
		Username:     user.Username,
		Nickname:     user.Nickname,
		Email:        user.Email,
		Mobile:       user.Mobile,
		Avatar:       user.Avatar,
		RefereeNums:  user.RefereeNums,
		TeamNums:     user.TeamNums,
		Token:        "",
		RefreshToken: "",
		Name:         user.Name,
		IdCard:       idcard,
		IsCertified:  user.IsCertified,
	}
	return &uinfo
}
func GetCtxUser(ctx context.Context) *model.User {
	uid := GetCtxUserId(ctx)
	user := GetUserById(uid)
	return user
}
func GetUserById(userId int64) *model.User {
	user := new(model.User)
	err := utils.Orm.First(user, userId).Error
	if err != nil {
		log.Println("GetUserById db err", err)
	}
	return user
}
func checkfundPwd(user *model.User, fundPwd string) bool {
	return pwdSum(fundPwd, user.Salt) == user.FundPassword
}
func pwdSum(pwd, salt string) string {
	return utils.GetMd5String(utils.GetMd5String(pwd) + salt)
}

// IsValidPhoneNumber 验证手机号是否有效
func IsValidPhoneNumber(phoneNumber string) bool {
	// 中国大陆的手机号码规则：1开头，第二位为3-9，后跟9位数字
	regex := `^1[3-9]\d{9}$`
	re := regexp.MustCompile(regex)
	return re.MatchString(phoneNumber)
}

func In_array(needle interface{}, hystack interface{}) bool {
	switch key := needle.(type) {
	case string:
		for _, item := range hystack.([]string) {
			if key == item {
				return true
			}
		}
	case int:
		for _, item := range hystack.([]int) {
			if key == item {
				return true
			}
		}
	case int64:
		for _, item := range hystack.([]int64) {
			if key == item {
				return true
			}
		}
	default:
		return false
	}
	return false
}

func FirstUpper(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func GetBankByCode(code string, length int) (*model.FinancialBank, error) {
	cardPre := code[0:length]
	bankName := mproto.CardNoPreMap[cardPre]
	if len(bankName) != 0 {
		bank := &model.FinancialBank{
			Status: 1,
			Name:   bankName,
		}
		err := utils.Orm.Order("id desc").First(bank, bank).Error
		if err != nil && err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("record not found")
		} else if err != nil {
			return nil, err
		}
		return bank, nil
	} else {
		return nil, fmt.Errorf("bank not found")
	}
}

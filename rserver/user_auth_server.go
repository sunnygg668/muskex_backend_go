package rserver

import (
	"connectrpc.com/connect"
	"context"
	"errors"
	"github.com/lithammer/shortuuid/v4"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"gorm.io/gorm"
	"log"
	"muskex/gen/mproto"
	"muskex/gen/mproto/model"
	"muskex/utils"
	"strconv"
	"strings"
	"time"
)

type UserAuthServer struct{}

func (UserAuthServer) VerifyCode(ctx context.Context, c *connect.Request[mproto.VerifyCodeRequest]) (*connect.Response[mproto.VerifyCodeResponse], error) {
	id, an, _ := utils.MobileCaptchTool.GenCaptcha(true, c.Msg.Mobile)
	//todo send sms
	log.Println("mobile:", c.Msg.Mobile, "verify code:", an)
	return connect.NewResponse(&mproto.VerifyCodeResponse{VerifyCodeId: id}), nil
}

func (UserAuthServer) SignUp(ctx context.Context, c *connect.Request[mproto.SignUpRequest]) (*connect.Response[mproto.SignUpResponse], error) {
	//check vcode
	if !utils.MobileCaptchTool.VerifyCaptcha(c.Msg.Mobile+"user_register", c.Msg.Vcode) {
		return nil, errors.New("验证码错误")
	}

	tuser := &model.User{}
	err := utils.Orm.First(tuser, "mobile=?", c.Msg.Mobile).Error
	if err == nil {
		return nil, errors.New("该手机号已注册，请使用其它手机号")
	}

	//check RegistCode
	if c.Msg.RegistCode != "" {
		err = utils.Orm.First(tuser, "invitationcode=?", c.Msg.RegistCode).Error
		if err != nil && err == gorm.ErrRecordNotFound {
			return nil, errors.New("邀请码错误")
		}
	}
	in := c.Msg
	icode := genInviteCode()
	uinfo := &model.User{
		Mobile:         in.Mobile,
		Salt:           icode,
		Password:       pwdSum(in.Password, icode),
		Invitationcode: icode,
		Refereeid:      tuser.Id,
		Status:         "1",
	}
	err = utils.Orm.Create(uinfo).Error
	if err != nil {
		return nil, err
	}

	//todo delete test
	res, err := CreateAddress("test " + uinfo.Mobile)
	if err != nil {
		log.Println("CreateAddress err", err)
	}
	//createAssets
	assets := []*model.UserAssets{}
	for id, _ := range IdCoinMap(ctx) {
		address := ""
		if id == 1 { // maincoin  usdt
			address = res.Data.Address
		}
		assets = append(assets, &model.UserAssets{
			UserId:  uinfo.Id,
			CoinId:  id,
			Address: address,
		})
	}
	utils.Orm.Create(&assets)

	//register_reward
	cfg := LoadDbConfig().NameValues
	giveCoinType, _ := strconv.Atoi(cfg["give_coin_type"])
	giveCoinNum, _ := strconv.ParseFloat(cfg["give_coin_num"], 64)
	if giveCoinType != 0 && giveCoinNum > 0 {
		err = UpdateCoinAssetsBalance(utils.Orm, uinfo.Id, int64(giveCoinType), giveCoinNum, "register_reward", uinfo.Id, 0, "")
		if err != nil {
			log.Println("Failed to update main coin assets balance for give amount:", err)
		}
	}
	//create UserTeamLevel
	lis := []int64{}
	utils.Orm.Model(model.UserLevel{}).Select("level").Pluck("level", &lis)
	uls := []*model.UserTeamLevel{}
	for _, li := range lis {
		uls = append(uls, &model.UserTeamLevel{
			UserId:      uinfo.Id,
			UserLevelId: li,
		})
	}
	utils.Orm.Create(&uls)
	if uinfo.Refereeid > 0 {
		sql1 := `
insert IGNORE into ba_team_user (id, pid, team_level)
select id, pid, team_level
from (WITH recursive parents as (select 1 as rlevel, ? as id, ? as pid
                                 union all
                                 select parents.rlevel + 1 as rlevel, parents.id, ifnull(ba_user.refereeid, 0) pid
                                 from parents
                                          inner join ba_user on parents.pid = ba_user.id)
      select id, pid, rlevel team_level
      from parents
      where pid > 0) abc
ON DUPLICATE KEY UPDATE team_level=abc.team_level;`
		utils.Orm.Exec(sql1, uinfo.Id, uinfo.Refereeid)
	}
	token, err := uinfo.GenToken(getIp(ctx), in.DeviceId, in.Os, in.OsVersion)
	return connect.NewResponse(&mproto.SignUpResponse{Token: token}), nil
}
func genInviteCode() string {
	icode := shortuuid.New()
	icode = icode[len(icode)-6:]
	return icode
}

func (UserAuthServer) SignIn(ctx context.Context, c *connect.Request[mproto.SignInRequest]) (*connect.Response[mproto.SignInResponse], error) {
	if c.Msg.UserName == "" || c.Msg.Password == "" {
		return nil, errors.New("Username and password cannot be empty")
	}
	uinfo := new(model.User)
	err := utils.Orm.First(uinfo, "username=? or mobile=?", c.Msg.UserName, c.Msg.UserName).Error
	if err != nil {
		return nil, err
	}
	var loginErr error
	if uinfo.Status == "0" {
		loginErr = errors.New("Account disabled")
	}
	if uinfo.LoginFailure >= 10 && time.Now().Unix()-uinfo.LastLoginTime < 86400 {
		loginErr = errors.New("Please try again after 1 day")
	}
	if uinfo.Password != pwdSum(c.Msg.Password, uinfo.Salt) {
		loginErr = errors.New("Password is incorrect")
	}
	if loginErr != nil {
		uinfo.LoginFailure++
		utils.RawOrm.Updates(&model.User{Id: uinfo.Id,
			LoginFailure:  uinfo.LoginFailure,
			LastLoginTime: time.Now().Unix(),
		})
		return nil, loginErr
	} else {
		uinfo.LastLoginIp = getIp(ctx)
		uinfo.LastLoginTime = time.Now().Unix()
		uinfo.LoginFailure = 0
		utils.RawOrm.Save(uinfo)
	}
	token, err := uinfo.GenToken(uinfo.LastLoginIp, c.Msg.DeviceId, c.Msg.Os, c.Msg.OsVersion)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&mproto.SignInResponse{Token: token, IsCertified: uinfo.IsCertified == 2}), nil
}

func getIp(ctx context.Context) string {
	ipAddress := ""
	m, ok := metadata.FromIncomingContext(ctx)
	if ok {
		ips := m.Get("X-Forwarded-For")
		if len(ips) > 0 {
			ipAddress = ips[0]
			//return ipAddress
		}
	}
	if len(ipAddress) == 0 {
		p, ok := peer.FromContext(ctx)
		if ok {
			ipAddress = p.Addr.String()
			//trim port
			items := strings.Split(ipAddress, ":")
			if len(items) == 2 {
				ipAddress = items[0]
			}
		}
	}
	//X-Forwarded-For may have multi ip
	iplist := strings.SplitN(ipAddress, ", ", 2)
	if len(iplist) > 0 {
		ipAddress = iplist[0]
	}
	return ipAddress
}

func (UserAuthServer) ForgetPwd(ctx context.Context, c *connect.Request[mproto.ForgetPwdRequest]) (*connect.Response[mproto.MsgResponse], error) {
	in := c.Msg
	if !utils.MobileCaptchTool.VerifyCaptcha(in.Mobile+"user_change_pwd", in.Vcode) {
		return nil, errors.New("验证码错误")
	}
	tuser := &model.User{}
	err := utils.Orm.First(tuser, "mobile=?", in.Mobile).Error
	if err != nil && err == gorm.ErrRecordNotFound {
		return nil, errors.New("该手机号用户未找到")
	} else if err != nil {
		return nil, err
	}
	//tuser.PasswordHash=getPwdHash(in.Password)
	err = utils.Orm.Updates(&model.User{Id: tuser.Id, Password: pwdSum(in.Password, tuser.Salt)}).Error
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&mproto.MsgResponse{Message: "密码修改成功"}), nil
}

func (UserAuthServer) UpdateFundPassword(ctx context.Context, c *connect.Request[mproto.ForgetPwdRequest]) (*connect.Response[mproto.MsgResponse], error) {
	in := c.Msg
	if !utils.MobileCaptchTool.VerifyCaptcha(in.Mobile+"user_retrieve_fund_pwd", in.Vcode) {
		return nil, errors.New("验证码错误")
	}
	tuser := &model.User{}
	err := utils.Orm.First(tuser, "mobile=?", in.Mobile).Error
	if err != nil && err == gorm.ErrRecordNotFound {
		return nil, errors.New("该手机号用户未找到")
	} else if err != nil {
		return nil, err
	}
	//tuser.PasswordHash=getPwdHash(in.Password)
	err = utils.Orm.Updates(&model.User{Id: tuser.Id, FundPassword: pwdSum(in.Password, tuser.Salt)}).Error
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&mproto.MsgResponse{Message: "密码修改成功"}), nil
}

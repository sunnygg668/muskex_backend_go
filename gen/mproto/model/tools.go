package model

import (
	"gorm.io/gorm"
	"muskex/utils"
	"strconv"
	"time"
)
import shortuuid "github.com/lithammer/shortuuid/v4"

type StatusConfig struct {
	Withdraw          map[string]string
	ManagementOrder   map[string]string
	FinancialRecharge map[string]string
	FinancialCard     map[string]string
	ContractOrder     map[string]string
}

var StatusTxtConfig = StatusConfig{
	Withdraw: map[string]string{
		"0": "处理中",
		"1": "已完成",
		"2": "已拒绝",
		"3": "已完成",
		"4": "审核中",
	},
	FinancialRecharge: map[string]string{
		"0": "等待支付",
		"1": "已完成",
		"2": "已取消",
		"3": "已完成",
	},
	FinancialCard: map[string]string{
		"0": "审核中",
		"1": "审核成功",
		"2": "审核失败",
	},
	ContractOrder: map[string]string{
		"0": "进行中",
		"1": "已完成",
	},
	ManagementOrder: map[string]string{
		"0": "未开始",
		"1": "进行中",
		"2": "已完成",
	},
}

var ManChangeType = map[string]string{
	"register_reward":               "注册奖励",
	"invite_register_reward":        "邀请注册奖励",
	"invite_first_recharge":         "邀请用户首充赠送",
	"recharge_coin":                 "充币",
	"u_recharge":                    "U充值赠送",
	"contract_buy":                  "合约保证金扣除",
	"contract_buy_fee":              "合约手续费",
	"contract_sell":                 "合约保证金退回",
	"miners_reward":                 "矿机代理返点",
	"auth_give":                     "完成认证赠送",
	"first_recharge_reached_give":   "首次充值满赠",
	"today_recharge_reached_give":   "当日充值满赠",
	"invite_num_reached_give":       "直推人数满赠",
	"team_num_reached_give":         "团队人数满赠",
	"today_contract_num_reached":    "当日交易笔数满赠",
	"today_contract_amount_reached": "当日交易量满赠",
	"month_contract_amount_reached": "当月交易量满赠",
	"today_invite_reached_give":     "当日直推满赠",
	"week_invite_reached_give":      "当周直推满赠",
	"month_invite_reached_give":     "当月直推满赠",
	"first_contract_amount_reached": "首次交易量满赠",
	"margin_reward":                 "保证金返点",
	"financial_recharge":            "快捷买币",
	"commission_pool_collect":       "佣金池领取",
	"system_recharge":               "充币",
	"system_deduction":              "系统扣除",
	"system_freeze":                 "系统冻结",
	"system_unfreeze":               "系统解冻",
	"coin_withdraw":                 "提币",
	"contract_payment":              "交易赔付",
	"check_in_reward":               "签到奖励",
	"transfer_in_money":             "转入理财钱包",
	"transfer_out_money":            "理财钱包转出",
	"management_buy":                "购买理财",
	"lease_miners":                  "租赁矿机",
	"management_income":             "理财收益",
	"miners_income":                 "矿机产出",
	"contract_income":               "合约盈亏",
	"management_total_price_return": "理财本金退回",
	"miners_real_pay_return":        "矿机到期退回",
	"coin_withdraw_fee":             "提币手续费",
	"bonus_award":                   "分红奖励",
	"coin_withdraw_return":          "提币退回",
	"coin_withdraw_fee_return":      "提币手续费退回",
	"lottery_deduction":             "抽奖扣除",
	"lottery_gain":                  "抽奖获得",
	"lottery_give":                  "抽奖赠送",
	"money_income":                  "余额宝收益",
	"rebate_income":                 "理财返点",
}

func (card *TradeContractOrder) AfterFind(tx *gorm.DB) (err error) {
	card.StatusText = StatusTxtConfig.ContractOrder[strconv.Itoa(int(card.Status))]
	if card.StatusText == "" {
		card.StatusText = "未知状态"
	}
	return
}
func (card *ManagementOrder) AfterFind(tx *gorm.DB) (err error) {
	card.StatusText = StatusTxtConfig.ManagementOrder[card.Status]
	if card.StatusText == "" {
		card.StatusText = "未知状态"
	}
	return
}
func (card *ManChange) AfterFind(tx *gorm.DB) (err error) {
	card.TypeName = ManChangeType[card.Type]
	if card.TypeName == "" {
		card.TypeName = "未知状态"
	}
	return
}
func (card *UserCommissionChange) AfterFind(tx *gorm.DB) (err error) {
	card.TypeName = ManChangeType[card.Type]
	if card.TypeName == "" {
		card.TypeName = "未知状态"
	}
	return
}
func (card *FinancialRecharge) AfterFind(tx *gorm.DB) (err error) {
	card.StatusText = StatusTxtConfig.FinancialRecharge[card.Status]
	if card.StatusText == "" {
		card.StatusText = "未知状态"
	}
	return
}
func (card *FinancialWithdraw) AfterFind(tx *gorm.DB) (err error) {
	card.StatusText = StatusTxtConfig.Withdraw[card.Status]
	if card.StatusText == "" {
		card.StatusText = "未知状态"
	}
	return
}
func (card *FinancialCard) AfterFind(tx *gorm.DB) (err error) {
	card.StatusText = StatusTxtConfig.FinancialCard[card.Status]
	if card.StatusText == "" {
		card.StatusText = "未知状态"
	}
	return
}

func (u *User) GenToken(ip, deviceId, os, osVersion string) (string, error) {
	tokenId := shortuuid.New()
	tokenObj := &TokenV2{
		Token:     tokenId,
		UserId:    u.Id,
		Ip:        ip,
		DeviceId:  deviceId,
		OsVersion: osVersion,
		Os:        os,
		//TokenType: TokenType,
	}
	utils.Orm.Exec("update ba_token_v2 set disabled =1,updated_at=now() where user_id=? and disabled =0 ", u.Id)

	err := utils.Orm.Create(tokenObj).Error
	if err != nil {
		return "", err
	}
	return tokenId, nil
}

type TokenV2 struct {
	Id        int64
	Token     string
	UserId    int64
	UpdatedAt time.Time
	CreatedAt time.Time
	Disabled  bool
	Ip        string
	DeviceId  string
	Os        string
	OsVersion string
	//TokenType string `gorm:"column:type"` //admin user
}

func (*ManagementOrder) TableName() string {
	return "ba_trade_management_order"
}
func (*ManChange) TableName() string {
	return "ba_user_management_change"
}

func (*Lecturer) TableName() string {
	return "ba_market_lecturer"
}

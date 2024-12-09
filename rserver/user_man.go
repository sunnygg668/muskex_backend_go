package rserver

import (
	"connectrpc.com/connect"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"google.golang.org/genproto/googleapis/api/httpbody"
	"gorm.io/gorm"
	"io"
	"log"
	"math"
	"math/rand"
	"muskex/gen/mproto"
	"muskex/gen/mproto/model"
	"muskex/utils"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

type UserManServer struct{}

func userIdFilter(c context.Context) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("user_id =?", GetCtxUserId(c))
	}
}
func getUserDb(c context.Context) *gorm.DB {
	return utils.Orm.Scopes(userIdFilter(c))
}
func getUserOp(c context.Context) *gorm.DB {
	return utils.Orm.Model(&model.User{Id: GetCtxUserId(c)})
}
func balanceCheck(c context.Context, coinId int64, amount float64) bool {
	asset := new(model.UserAssets)
	getUserDb(c).Select("balance").First(asset, " coin_id=?", coinId)
	return asset.Balance >= amount
}

/*php updateManagement*/
func UpdateCmBalance(tx *gorm.DB, user model.User, amount float64, changeType string, refId int64, fromUserId int64, remark string) error {
	if amount == 0 {
		return nil
	}
	if amount < 0 && user.Money < -amount {
		return errors.New("理财钱包余额不足")
	}
	before := user.Money
	user.Money += amount
	tx.Updates(&model.User{Id: user.Id, Money: user.Money})
	manChange := &model.ManChange{
		UserId:     user.Id,
		Amount:     amount,
		Before:     before,
		After:      user.Money,
		Type:       changeType,
		FromUserId: fromUserId,
		Remark:     remark,
		ReferrerId: refId,
	}
	tx.Create(manChange)
	return nil
}
func UpdateCoinAssetsBalance(tx *gorm.DB, userId, coinId int64, amount float64, changeType string, refId int64, fromUserId int64, remark string) error {
	asset := new(model.UserAssets)
	err := tx.First(asset, "user_id=? and coin_id=?", userId, coinId).Error
	if err != nil && err == gorm.ErrRecordNotFound {
		asset = &model.UserAssets{
			UserId: userId,
			CoinId: coinId,
		}
		utils.Orm.Create(asset)
	} else if err != nil {
		return err
	}
	if amount == 0 {
		return nil
	}
	if amount < 0 && asset.Balance < -amount {
		return errors.New("余额不足")
	}
	before := asset.Balance
	asset.Balance += amount
	tx.Save(asset)
	coinChange := &model.UserCoinChange{
		UserId:     userId,
		CoinId:     coinId,
		Amount:     amount,
		Before:     before,
		After:      asset.Balance,
		Type:       changeType,
		FromUserId: fromUserId,
		Remark:     remark,
		ReferrerId: refId,
	}
	tx.Create(coinChange)
	return nil
}

// rpc  SubmitRecharge(SubmitRechargeRequest) returns (SubmitRechargeResponse) {}
func (s UserManServer) SubmitRecharge(ctx context.Context, req *connect.Request[mproto.SubmitRechargeRequest]) (*connect.Response[mproto.SubmitRechargeResponse], error) {
	userId := GetCtxUserId(ctx)
	user := GetCtxUser(ctx)
	methodId := req.Msg.MethodId
	name := req.Msg.Name
	amount := req.Msg.Amount

	if user.IsCertified == 0 || user.Idcard == "" || user.IsCertified != 2 {
		return nil, errors.New("请先完成实名认证")
	}

	var method *model.FinancialPaymentMethod
	if methodId == -1 {
		var quickMethodList []model.FinancialPaymentMethod
		if err := utils.Orm.Where("status = ? AND type = ?", 1, "0").Find(&quickMethodList).Error; err != nil {
			return nil, err
		}
		for _, quickMethod := range quickMethodList {
			if amount >= quickMethod.MinAmount && amount <= quickMethod.MaxAmount {
				method = &quickMethod
				break
			}
		}
	} else {
		if err := utils.Orm.First(&method, methodId).Error; err != nil {
			return nil, err
		}
	}

	if method == nil || method.Status == 0 {
		return nil, errors.New("该支付方式暂不可用，请刷新重试")
	}

	minAmount := method.MinAmount
	maxAmount := method.MaxAmount
	if amount < minAmount || amount > maxAmount {
		return nil, errors.New(fmt.Sprintf("充值金额的范围为：%.2f - %.2f", minAmount, maxAmount))
	}

	mainCoinPrice := float32(getUPrice())
	feeRatio := method.FeeRatio
	fee := amount * (feeRatio / 100)
	actualMoney := amount - fee
	mainCoinNum := actualMoney / mainCoinPrice
	mainCoinFee := fee / mainCoinPrice

	orderNo := fmt.Sprintf("FR%s%05d", time.Now().Format("20060102"), rand.Intn(99999))
	rechargeData := model.FinancialRecharge{
		UserId:                   userId,
		OrderNo:                  orderNo,
		Name:                     name,
		FinancialPaymentMethodId: method.Id,
		Amount:                   amount,
		FeeRatio:                 feeRatio,
		Fee:                      fee,
		ActualMoney:              actualMoney,
		MainCoinNum:              mainCoinNum,
		MainCoinFee:              mainCoinFee,
		CoinPrice:                mainCoinPrice,
	}

	if err := utils.Orm.Create(&rechargeData).Error; err != nil {
		return nil, err
	}

	var url string
	if method.Type == "0" {
		url1, err := s.quickPay2(method, amount, orderNo)
		if err != nil {
			return nil, err
		}
		url = url1
	}
	if url != "" {
		return connect.NewResponse(&mproto.SubmitRechargeResponse{RedirectUrl: url}), nil
	} else {
		return connect.NewResponse(&mproto.SubmitRechargeResponse{Message: "提交成功，等待商家确认"}), nil
	}
}
func (s UserManServer) quickPay2(method *model.FinancialPaymentMethod, amount float32, orderNo string) (string, error) {
	cfg := LoadDbConfig().NameValues
	//todo config notifyUrl in db
	notifyUrl := fmt.Sprintf("%s/notify2", cfg["app.pay_notifyurl"])
	params := map[string]string{
		"pay_memberid":    method.MerchantNum,
		"pay_orderid":     orderNo,
		"pay_amount":      fmt.Sprintf("%.2f", amount),
		"pay_applydate":   time.Now().Format("2006-01-02 15:04:05"),
		"pay_bankcode":    method.PaymentChannels,
		"pay_notifyurl":   notifyUrl,
		"pay_callbackurl": "",
	}

	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var paramsArray []string
	for _, k := range keys {
		if v := params[k]; v != "" {
			paramsArray = append(paramsArray, fmt.Sprintf("%s=%s", k, v))
		}
	}

	paramsStr := strings.Join(paramsArray, "&")
	params["pay_md5sign"] = strings.ToUpper(utils.GetMd5String(fmt.Sprintf("%s&key=%s", paramsStr, method.EncryptionKey)))
	params["pay_productname"] = "VIP基础服务"

	client := &http.Client{}
	form := url.Values{}
	for k, v := range params {
		form.Add(k, v)
	}

	req, err := http.NewRequest("POST", method.Url, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var resData map[string]interface{}
	if err := json.Unmarshal(body, &resData); err != nil {
		return "", err
	}

	if code, ok := resData["code"].(string); ok && code == "200" {
		if data, ok := resData["data"].(string); ok {
			return data, nil
		}
	}
	return "", errors.New("充值通道繁忙，请稍后再试！")
}
func (UserManServer) RechargeList(c context.Context, req *connect.Request[mproto.NullMsg]) (*connect.Response[mproto.RechargeListResponse], error) {

	items := []*model.FinancialRecharge{}
	err := getUserDb(c).Order("id desc").Find(&items).Error
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&mproto.RechargeListResponse{List: items}), nil
}
func (UserManServer) QuickPayMethodList(c context.Context, req *connect.Request[mproto.NullMsg]) (*connect.Response[mproto.QuickPayMethodListResponse], error) {
	cfg := LoadDbConfig().NameValues
	mainCoinPrice := getUPrice()
	rechargeMoneyTip := cfg["recharge_money_tip"]
	methodList := []*model.FinancialPaymentMethod{}

	var quickMethodList []*model.FinancialPaymentMethod
	if err := utils.Orm.Where("status = ? AND type = ?", 1, "0").Find(&quickMethodList).Error; err != nil {
		return nil, err
	}

	var otherMethodList []*model.FinancialPaymentMethod
	if err := utils.Orm.Where("status = ?", 1).Where("type IN (?)", []string{"1", "2", "3"}).Order("weigh desc").Find(&otherMethodList).Error; err != nil {
		return nil, err
	}

	if len(quickMethodList) > 0 {
		minAmountValues := make([]float32, len(quickMethodList))
		maxAmountValues := make([]float32, len(quickMethodList))
		for i, method := range quickMethodList {
			minAmountValues[i] = method.MinAmount
			maxAmountValues[i] = method.MaxAmount
		}
		minAmount := min(minAmountValues)
		maxAmount := max(maxAmountValues)
		methodList = append(methodList, &model.FinancialPaymentMethod{
			Id:        -1,
			Name:      "快捷支付",
			ShortName: "快捷支付",
			Type:      "0",
			MinAmount: minAmount,
			MaxAmount: maxAmount,
		})
	}

	for _, method := range otherMethodList {
		methodList = append(methodList, method)
	}

	result := &mproto.QuickPayMethodListResponse{
		MainCoinPrice:    float32(mainCoinPrice),
		RechargeMoneyTip: rechargeMoneyTip,
		MethodList:       methodList,
	}
	return connect.NewResponse(result), nil
}

func min(arr []float32) float32 {
	min := arr[0]
	for _, v := range arr {
		if v < min {
			min = v
		}
	}
	return min
}

func max(arr []float32) float32 {
	max := arr[0]
	for _, v := range arr {
		if v > max {
			max = v
		}
	}
	return max
}

func (UserManServer) BuyMan(c context.Context, req *connect.Request[mproto.BuyManRequest]) (*connect.Response[mproto.MsgResponse], error) {
	userId := GetCtxUserId(c)
	user := GetCtxUser(c)
	id := req.Msg.Id
	num := req.Msg.Num
	fundPassword := req.Msg.FundPassword

	if num <= 0 {
		return nil, errors.New("请输入正确的数量")
	}
	if !checkfundPwd(user, fundPassword) {
		return nil, errors.New("资金密码错误")
	}

	management := new(model.CoinManagement)
	if err := utils.Orm.First(management, id).Error; err != nil {
		return nil, errors.New("理财产品不存在")
	}
	if management.Status == 0 {
		return nil, errors.New("该理财产品已停售")
	}
	if num < management.MinBuyNum || num > management.MaxBuyNum {
		return nil, fmt.Errorf("可申购数量范围为：%d - %d", management.MinBuyNum, management.MaxBuyNum)
	}
	leftNum := management.IssuesNum - management.SoldNum
	if num > leftNum {
		return nil, fmt.Errorf("可申购数量不足，目前可申购数：%d", leftNum)
	}

	orderNo := fmt.Sprintf("TMO%s%05d", time.Now().Format("20060102"), rand.Intn(99999))
	closedDays := float64(management.ClosedDays)
	incomeType := management.IncomeType
	incomeRatio := management.IncomeRatio
	totalPrice := float64(num) * management.Price
	unitTotalIncome := totalPrice * (incomeRatio / 100)
	totalIncome := 0.0

	switch incomeType {
	case "hour":
		totalIncome = closedDays * 24 * unitTotalIncome
	case "day":
		totalIncome = closedDays * unitTotalIncome
	case "month":
		totalIncome = (closedDays / 30) * unitTotalIncome
	case "year":
		totalIncome = (closedDays / 365) * unitTotalIncome
	}

	cfg := LoadDbConfig().NameValues
	rebateRatio := 0.0
	if closedDays < 7 {
		rebateRatio = 0
	} else if closedDays == 7 {
		rebateRatio, _ = strconv.ParseFloat(cfg["7_day_rebate_rate"], 64)
	} else if closedDays > 7 && closedDays <= 30 {
		rebateRatio, _ = strconv.ParseFloat(cfg["30_day_rebate_rate"], 64)
	} else {
		rebateRatio, _ = strconv.ParseFloat(cfg["30_over_day_rebate_rate"], 64)
	}
	rebateIncome := (totalPrice / 2) * (rebateRatio / 100)

	tx := utils.Orm.Begin()
	defer tx.Rollback()
	order := model.ManagementOrder{
		OrderNo:          orderNo,
		UserId:           userId,
		Refereeid:        user.Refereeid,
		TeamLeaderId:     user.TeamLeaderId,
		CoinManagementId: id,
		SettlementCoinId: management.SettlementCoinId,
		IncomeCoinId:     management.IncomeCoinId,
		Price:            management.Price,
		BuyNum:           num,
		TotalPrice:       totalPrice,
		IncomeType:       incomeType,
		IncomeRatio:      incomeRatio,
		TotalIncome:      totalIncome,
		RebateIncome:     rebateIncome,
		ClosedDays:       int64(closedDays),
		ExpireTime:       time.Now().AddDate(0, 0, int(closedDays)).Unix(),
	}

	if err := tx.Create(&order).Error; err != nil {
		return nil, err
	}
	if err := UpdateCoinAssetsBalance(tx, userId, management.SettlementCoinId, -totalPrice, "management_buy", 0, 0, ""); err != nil {
		return nil, err
	}

	if err := tx.Model(management).Update("sold_num", gorm.Expr("sold_num + ?", num)).Error; err != nil {
		return nil, err
	}
	tx.Commit()
	updateCommission(user, rebateIncome, "rebate", "Rebate income", order.Id)

	return connect.NewResponse(&mproto.MsgResponse{Message: "购买成功"}), nil
}

func (UserManServer) ManagementOrderList(c context.Context, req *connect.Request[mproto.NullMsg]) (*connect.Response[mproto.ManagementOrderListResponse], error) {
	items := []*model.ManagementOrder{}
	err := getUserDb(c).Preload("CoinManagement").Order("id desc").Find(&items).Error
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&mproto.ManagementOrderListResponse{List: items}), nil
}
func (UserManServer) CommissionPoolIndex(c context.Context, req *connect.Request[mproto.NullMsg]) (*connect.Response[mproto.CommissionPoolIndexResponse], error) {
	user := GetCtxUser(c)
	level := new(model.UserLevel)
	if err := utils.Orm.Where("level = ?", user.Level).First(level).Error; err != nil {
		return nil, err
	}
	cfg := LoadDbConfig().NameValues
	commissionPoolTip := cfg["commission_pool_tip"]
	totalAmount := 0.0
	utils.Orm.Model(&model.UserCommissionChange{}).Where("user_id = ? AND type = ?", user.Id, "commission_pool_collect").Select("SUM(amount)").Scan(&totalAmount)
	totalAmount = math.Abs(totalAmount)

	teamNums := user.TeamNums
	teamNumsGrade := int64(0)
	collectNum := 0
	canCollect := false

	grades := collectGrades{}
	grades.load("collect_grade", false)

	for _, cGrade := range grades {
		count := int64(0)
		utils.Orm.Model(&model.UserCommissionChange{}).Where("user_id = ? AND type = ? AND remark = ?", user.Id, "commission_pool_collect", cGrade.Key).Count(&count)
		canCollect = (count == 0)
		gkey := int64(cGrade.Key)
		if teamNums < int64(gkey) || canCollect {
			teamNumsGrade = int64(gkey)
			collectNum = int(cGrade.Value)
			break
		}
	}

	canCollect = canCollect && (teamNums >= teamNumsGrade) && (user.CommissionPool >= float64(collectNum))

	result := &mproto.CommissionPoolIndexResponse{
		User:              user,
		Level:             level,
		CommissionPoolTip: commissionPoolTip,
		TotalAmount:       totalAmount,
		TeamNums:          teamNums,
		TeamNumsGrade:     teamNumsGrade,
		CollectNum:        int64(collectNum),
		CanCollect:        canCollect,
	}

	return connect.NewResponse(result), nil
}

type rcollectGrade struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
type collectGrade struct {
	Key   float64 `json:"key"`
	Value float64 `json:"value"`
}
type collectGrades []*collectGrade

func (cg collectGrades) Process(cfgKey string, kfilter float64, proc func(key, value float64)) {
	cg.load(cfgKey, true)
	for _, give := range cg {
		if kfilter >= give.Key && give.Value > 0 {
			proc(give.Key, give.Value)
			break
		}
	}
}
func (cg collectGrades) load(key string, isDesc bool) {
	items := []*rcollectGrade{}
	json.Unmarshal([]byte(GetCfgValue(key)), items)
	haveValue := false
	for _, item := range items {
		if item.Key != "0" {
			haveValue = true
		}
	}
	if !haveValue {
		return
	}
	for _, item := range items {
		kvalue, _ := strconv.ParseFloat(item.Key, 64)
		value, _ := strconv.ParseFloat(item.Value, 64)
		cg = append(cg, &collectGrade{Key: kvalue, Value: value})
	}
	sort.Slice(cg, func(i, j int) bool {
		if isDesc {
			return cg[i].Key > cg[j].Key
		} else {
			return cg[i].Key < cg[j].Key
		}
	})
}

func (UserManServer) BuyContract(c context.Context, req *connect.Request[mproto.BuyContractRequest]) (*connect.Response[mproto.MsgResponse], error) {
	userId := GetCtxUserId(c)
	user := GetCtxUser(c)
	klineType := req.Msg.ContractName
	num := float32(req.Msg.Num)

	if user.IsCertified == 0 || user.Idcard == "" || user.IsCertified != 2 {
		return nil, errors.New("请先完成实名认证")
	}

	tx := utils.Orm.Begin()
	defer tx.Rollback()

	coin := new(model.Coin)
	if err := tx.Where("name = ?", klineType).First(coin).Error; err != nil {
		return nil, errors.New("当前合约未开放")
	}

	contract := new(model.CoinContract)
	if err := tx.Where("coin_id = ? AND status = 1", coin.Id).First(contract).Error; err != nil {
		return nil, errors.New("当前合约未开放")
	}

	existsOrder := new(model.TradeContractOrder)
	if err := tx.Where("user_id = ? AND contract_id = ? AND status = 0", userId, contract.Id).First(existsOrder).Error; err == nil {
		return nil, errors.New("当前有未完成的合约订单，禁止交易")
	}

	purchaseUp := contract.PurchaseUp
	purchaseDown := contract.PurchaseDown
	if purchaseUp > 0 && num > purchaseUp {
		return nil, fmt.Errorf("购买上限为 %d", purchaseUp)
	}
	if purchaseDown > 0 && num < purchaseDown {
		return nil, fmt.Errorf("购买下限为 %d", purchaseDown)
	}

	margin := float32(coin.Margin)
	feeRatio := contract.FeeRatio
	investedCoinNum := float64(margin * num)
	fee := float64(investedCoinNum * (float64(feeRatio) / 100))

	orderNo := fmt.Sprintf("CT%s%05d", time.Now().Format("20060102"), rand.Intn(99999))
	order := model.TradeContractOrder{
		UserId:          userId,
		Refereeid:       user.Refereeid,
		TeamLeaderId:    user.TeamLeaderId,
		ContractId:      contract.Id,
		CoinId:          coin.Id,
		OrderNo:         orderNo,
		Title:           coin.Name,
		Num:             num,
		BuyPrice:        margin,
		InvestedCoinNum: investedCoinNum,
		Fee:             fee,
		FeeRatio:        feeRatio,
		BuyTime:         time.Now().Unix(),
	}

	if err := tx.Create(&order).Error; err != nil {
		return nil, err
	}
	if err := UpdateCoinAssetsBalance(tx, userId, coin.Id, -investedCoinNum, "contract_buy", order.Id, 0, ""); err != nil {
		return nil, err
	}
	if err := UpdateCoinAssetsBalance(tx, userId, coin.Id, -fee, "contract_buy_fee", order.Id, 0, ""); err != nil {
		return nil, err
	}

	tx.Commit()
	rew := Reward{}
	rew.userActive(user)
	rew.contractBuyForParent(user, investedCoinNum, order.Id)
	rew1 := reward1{}
	rew1.todayContractNumReached(user.Id, order.Id)
	rew1.firstContractAmountReached(user.Id, order.Id)
	rew1.todayContractAmountReached(user.Id, order.Id)
	rew1.monthContractAmountReached(user.Id, order.Id)
	// Queue jobs (assuming these functions are defined elsewhere)
	//QueuePush("userActivation", map[string]interface{}{"user_id": userId})
	//QueuePush("contractBuy", map[string]interface{}{"user_id": userId, "margin": investedCoinNum})
	//QueuePush("todayContractNumReached", map[string]interface{}{"user_id": userId})
	//QueuePush("todayContractAmountReached", map[string]interface{}{"user_id": userId})
	//QueuePush("monthContractAmountReached", map[string]interface{}{"user_id": userId})
	//QueuePush("firstContractAmountReached", map[string]interface{}{"user_id": userId})

	return connect.NewResponse(&mproto.MsgResponse{Message: "购买成功"}), nil
}
func (UserManServer) TradeContractOrderList(c context.Context, req *connect.Request[mproto.IdParam]) (*connect.Response[mproto.TradeContractOrderListResponse], error) {
	items := []*model.TradeContractOrder{}
	err := utils.Orm.Where("status=?", req.Msg.Id).Order("id desc").Find(&items).Error
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&mproto.TradeContractOrderListResponse{List: items}), nil
}
func (UserManServer) LecturerList(c context.Context, req *connect.Request[mproto.NullMsg]) (*connect.Response[mproto.LecturerListResponse], error) {
	items := []*model.Lecturer{}
	err := utils.Orm.Where("status=1").Order("weigh desc").Find(&items).Error
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&mproto.LecturerListResponse{List: items}), nil
}
func (UserManServer) CommissionCollect(c context.Context, req *connect.Request[mproto.NullMsg]) (*connect.Response[mproto.MsgResponse], error) {
	user := GetCtxUser(c)
	err := Reward{}.collect(user)
	if err != nil {
		return nil, err
	}
	return returnMsg("领取成功")
}
func (UserManServer) CommissionChangeList(c context.Context, req *connect.Request[mproto.NullMsg]) (*connect.Response[mproto.CommissionChangeListResponse], error) {
	items := []*model.UserCommissionChange{}
	err := getUserDb(c).Order("id desc").Find(&items).Error
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&mproto.CommissionChangeListResponse{List: items}), nil
}
func (UserManServer) ManChangeList(c context.Context, req *connect.Request[mproto.NullMsg]) (*connect.Response[mproto.ManChangeListResponse], error) {
	items := []*model.ManChange{}
	err := getUserDb(c).Where("type != 'system_deduction'").Order("id desc").Find(&items).Error
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&mproto.ManChangeListResponse{List: items}), nil
}
func (UserManServer) CMTransferIn(c context.Context, req *connect.Request[mproto.TransferRequest]) (*connect.Response[mproto.MsgResponse], error) {
	cfg := LoadDbConfig().NameValues
	minAmt, _ := strconv.ParseFloat(cfg["management_min_amount"], 64)
	if req.Msg.Amount < minAmt {
		return nil, errors.New(fmt.Sprintf("最小转入金额为：%0.2f", minAmt))
	}
	user := GetCtxUser(c)
	if !checkfundPwd(user, req.Msg.FundPassword) {
		return nil, errors.New("资金密码错误")
	}
	tx := utils.Orm.Begin()
	defer tx.Rollback()
	err := UpdateCoinAssetsBalance(tx, user.Id, 1, -req.Msg.Amount, "transfer_in_money", 0, 0, "")
	if err != nil {
		return nil, err
	}
	err = UpdateCmBalance(tx, *user, req.Msg.Amount, "transfer_in_money", 0, 0, "")
	if err != nil {
		return nil, err
	}
	tx.Commit()
	return returnMsg("转入成功")
}
func (UserManServer) CMTransferOut(c context.Context, req *connect.Request[mproto.TransferRequest]) (*connect.Response[mproto.MsgResponse], error) {
	user := GetCtxUser(c)
	if !checkfundPwd(user, req.Msg.FundPassword) {
		return nil, errors.New("资金密码错误")
	}
	tx := utils.Orm.Begin()
	defer tx.Rollback()
	err := UpdateCmBalance(tx, *user, -req.Msg.Amount, "transfer_out_money", 0, 0, "")
	if err != nil {
		return nil, err
	}
	err = UpdateCoinAssetsBalance(tx, user.Id, 1, req.Msg.Amount, "transfer_out_money", 0, 0, "")
	if err != nil {
		return nil, err
	}

	tx.Commit()
	return returnMsg("转出成功")
}

func (UserManServer) CMWalletInfo(c context.Context, req *connect.Request[mproto.NullMsg]) (*connect.Response[mproto.WalletBalance], error) {
	asset := new(model.UserAssets)
	getUserDb(c).First(asset, "coin_id=1")
	user := GetCtxUser(c)
	cfg := LoadDbConfig().NameValues
	return connect.NewResponse(&mproto.WalletBalance{
		Usdt:                 asset.Balance,
		Money:                user.Money,
		MoneyHourIncomeRatio: cfg["money_hour_income_ratio"],
	}), nil
}

func (UserManServer) ListWithdraw(c context.Context, req *connect.Request[mproto.StringParam]) (*connect.Response[mproto.ListWithdrawResponse], error) {
	items := []*model.FinancialWithdraw{}
	err := getUserDb(c).Preload("FinancialCard").Order("id desc").Find(&items, "type=?", req.Msg.Str).Error
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&mproto.ListWithdrawResponse{List: items}), nil
}
func (UserManServer) WithdrawApply(c context.Context, req *connect.Request[mproto.WithdrawApplyRquest]) (*connect.Response[mproto.MsgResponse], error) {
	user := GetCtxUser(c)
	cfg := LoadDbConfig().Names

	isCertified := user.IsCertified == 2
	fundPassword := req.Msg.FundPassword
	amount := req.Msg.Amount
	addressId := req.Msg.AddressId
	cardId := req.Msg.CardId
	withdrawType := req.Msg.Type

	if fundPassword == "" {
		return returnMsg("非法操作！")
	}

	if !isCertified {
		return returnMsg("未通过实名认证，无法提现")

	}
	if !checkfundPwd(user, fundPassword) {
		return returnMsg("资金密码错误")
	}

	if user.IsCanWithdraw == 0 {
		return returnMsg("当前账号暂时无法提现")
	}

	if user.IsActivation == 0 {
		return returnMsg("活跃度不足，无法提现")
	}

	if user.Idcard == "" {
		return returnMsg("请先完成实名认证")
	}

	if user.LimitWithdrawTime > 0 && user.LimitWithdrawTime > time.Now().Unix() {
		return returnMsg(fmt.Sprintf("%s 之前不可提现（安全保护期中）", time.Unix(user.LimitWithdrawTime, 0).Format("2006-01-02 15:04:05")))
	}

	dayWithdrawNum, _ := strconv.Atoi(cfg["day_withdraw_num"].Value)
	dayWithdrawCount := getUserOp(c).Where("create_time > ?", getDayStart().Unix()).Association("Withdraws").Count()
	if dayWithdrawCount >= int64(dayWithdrawNum) {
		return returnMsg(fmt.Sprintf("当日提现次数已超过 %d 次，暂时无法提现", dayWithdrawNum))
	}

	feeRatio, _ := strconv.ParseFloat(cfg["withdraw_fee_ratio"].Value, 32)
	dayWithdrawFreeCount, _ := strconv.Atoi(cfg["day_withdraw_free_count"].Value)
	if dayWithdrawCount < int64(dayWithdrawFreeCount) {
		feeRatio = 0
	}

	orderNo := fmt.Sprintf("W%s%08d", time.Now().Format("20060102"), rand.Intn(99999999))

	tx := utils.Orm.Begin()
	defer func() {
		tx.Rollback()
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	err := withdraw(c, tx, user.Id, user, withdrawType, amount, orderNo, int64(addressId), int64(cardId), feeRatio)
	if err != nil {
		return nil, err
	}
	tx.Commit()

	return returnMsg("提现申请已提交")
}
func withdraw(c context.Context, tx *gorm.DB, userId int64, user *model.User, withdrawType string, amount float64, orderNo string, addressId, cardId int64, feeRatio float64) error {
	cfg := LoadDbConfig().Names

	coinPrice := getUPrice()
	coinAmount := 0.0
	money := 0.0
	if withdrawType == "1" {
		//coin
		coinAmount = amount
		money = coinAmount * coinPrice
	} else {
		money = amount
		coinAmount = money / coinPrice
	}
	feeMoney := money * (feeRatio / 100)
	feeCoin := coinAmount * (feeRatio / 100)

	if withdrawType == "1" {
		withdrawMinCoinNum, _ := strconv.ParseFloat(cfg["withdraw_min_coin_num"].Value, 64)
		withdrawMaxCoinNum, _ := strconv.ParseFloat(cfg["withdraw_max_coin_num"].Value, 64)
		if coinAmount < withdrawMinCoinNum || coinAmount > withdrawMaxCoinNum {
			return errors.New(fmt.Sprintf("可提币数量为：%0.2f - %0.2f", withdrawMinCoinNum, withdrawMaxCoinNum))
		}
	} else {
		withdrawMinNum, _ := strconv.ParseFloat(cfg["withdraw_min_num"].Value, 32)
		withdrawMaxNum, _ := strconv.ParseFloat(cfg["withdraw_max_num"].Value, 32)
		if money < withdrawMinNum || money > withdrawMaxNum {
			return errors.New(fmt.Sprintf("可提现金额为：%0.2f - %0.2f", withdrawMinNum, withdrawMaxNum))
		}
	}

	withdrawRechargeUsdtNum, _ := strconv.ParseFloat(cfg["withdraw_recharge_usdt_num"].Value, 64)
	totalRecharge := 0.0
	userdb := getUserDb(c)
	getUserDb(c).Model(model.FinancialRecharge{}).Select("sum(amount)").Where("status=1 or status=3").Scan(&totalRecharge)
	if totalRecharge < withdrawRechargeUsdtNum {
		mainCoin := "USDT"
		return errors.New(fmt.Sprintf("充值满 %0.2f %s 可提现", withdrawRechargeUsdtNum, mainCoin))
	}

	address := ""
	if withdrawType == "1" {
		addressModel := &model.FinancialAddress{}
		userdb.First(addressModel, addressId)
		if addressModel.Address == "" {
			return errors.New("钱包地址不存在")
		}
		if !strings.HasPrefix(addressModel.Address, "T") {
			return errors.New("钱包地址格式错误")
		}
		address = addressModel.Address
	}

	if !balanceCheck(c, 1, coinAmount+feeCoin) {
		return errors.New("可提现余额不足")
	}
	withdrawItem := model.FinancialWithdraw{
		UserId:          userId,
		Refereeid:       user.Refereeid,
		TeamLeaderId:    user.TeamLeaderId,
		Type:            withdrawType,
		CoinId:          1,
		OrderNo:         orderNo,
		Money:           float32(money),
		CoinNum:         float32(coinAmount),
		Price:           float32(coinPrice),
		WalletType:      "TRC20",
		WalletAddress:   address,
		AddressId:       addressId,
		FinancialCardId: cardId,
		FeeRatio:        float32(feeRatio),
		FeeMoney:        float32(feeMoney),
		FeeCoin:         float32(feeCoin),
		ActualMoney:     float32(money),
		ActualCoin:      float32(coinAmount),
		Status:          "0",
	}

	if err := tx.Create(&withdrawItem).Error; err != nil {
		return err
	}

	if err := UpdateCoinAssetsBalance(tx, userId, 1, -coinAmount, "coin_withdraw", withdrawItem.Id, 0, ""); err != nil {
		return err
	}
	if err := UpdateCoinAssetsBalance(tx, userId, 1, -feeCoin, "coin_withdraw_fee", withdrawItem.Id, 0, ""); err != nil {
		return err
	}
	//SetRedisValue(orderNo, address)
	return nil
}
func getUPrice() float64 {
	cfg := LoadDbConfig().NameValues
	uprice := cfg["main_coin_price"]
	if uprice == "" {
		uprice = "7.3"
	}
	price, _ := strconv.ParseFloat(uprice, 64)
	return price
}
func (UserManServer) WithdrawInfo(c context.Context, req *connect.Request[mproto.NullMsg]) (*connect.Response[mproto.WithdrawInfoResponse], error) {
	cfg := LoadDbConfig().Names
	price := getUPrice()
	res := mproto.WithdrawInfoResponse{
		Balance:            getUserUsdtBlance(c),
		UsdtPrice:          price,
		OpenWithdrawUsdt:   cfg["open_withdraw_usdt"].Value == "1",
		OpenWithdrawMoney:  cfg["open_withdraw_money"].Value == "1",
		WithdrawMinNum:     cfg["withdraw_min_num"].Value,
		WithdrawMaxNum:     cfg["withdraw_max_num"].Value,
		WithdrawMinCoinNum: cfg["withdraw_min_coin_num"].Value,
		WithdrawMaxCoinNum: cfg["withdraw_max_coin_num"].Value,
		WithdrawRuleTip:    cfg["withdraw_rule_tip"].Value,
		FeeRatio:           cfg["withdraw_fee_ratio"].Value,
	}
	return connect.NewResponse(&res), nil
}
func (UserManServer) AddWithdrawAddress(c context.Context, req *connect.Request[mproto.AddWithdrawAddressRequest]) (*connect.Response[mproto.MsgResponse], error) {
	count := getUserOp(c).Where("address=?", req.Msg.Address).Association("WithdrawAddresses").Count()
	if count > 0 {
		return returnMsg("该地址已存在")
	}
	err := getUserOp(c).Association("WithdrawAddresses").Append(&model.FinancialAddress{
		Address: req.Msg.Address,
		Name:    req.Msg.Name,
	})
	if err != nil {
		return nil, err
	}
	return returnMsg("添加成功")
}
func (UserManServer) ListCard(c context.Context, req *connect.Request[mproto.NullMsg]) (*connect.Response[mproto.ListCardtResponse], error) {
	items := []*model.FinancialCard{}
	err := getUserDb(c).Find(&items, "status=?", "1").Error
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&mproto.ListCardtResponse{List: items}), nil
}
func (UserManServer) ListWithdrawAddress(c context.Context, req *connect.Request[mproto.NullMsg]) (*connect.Response[mproto.ListWithdrawAddressResponse], error) {
	items := []*model.FinancialAddress{}
	err := getUserDb(c).Find(&items, "status=?", 1).Error
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&mproto.ListWithdrawAddressResponse{List: items}), nil
}
func getUserUsdtBlance(c context.Context) float64 {
	asset := new(model.UserAssets)
	getUserDb(c).First(asset, "coin_id=1")
	return asset.Balance
}
func (UserManServer) AssetBalanceWithTip(c context.Context, req *connect.Request[mproto.IdParam]) (*connect.Response[mproto.AssetBalanceWithTipResponse], error) {
	asset := new(model.UserAssets)
	getUserDb(c).First(asset, "coin_id=1")
	tip := LoadDbConfig().Names["u_recharge_tip"].Value
	asset.CoinName = IdCoinMap(c)[asset.CoinId]
	return connect.NewResponse(&mproto.AssetBalanceWithTipResponse{Asset: asset, Tip: tip}), nil
}
func (UserManServer) MinerOrderList(c context.Context, req *connect.Request[mproto.MinerOrderListRequest]) (*connect.Response[mproto.MinerOrderListResponse], error) {
	items := []*model.MinersOrder{}
	if req.Msg.Status == 0 {
		req.Msg.Status = 1
	}
	utils.Orm.Preload("Miners").Order("id desc").Find(&items, "user_id=? and status=?", GetCtxUserId(c), req.Msg.Status)

	//foreach ($orderList as &$order) {
	//	$estimatedIncome = $order->estimated_income;
	//	$runRatio = bcdiv(time() - $order->create_time, $order->run_days * 86400, 2);
	//	$gainedIncome = bcmul($estimatedIncome, $runRatio, 2);
	//	$order->gained_income = min($gainedIncome, $estimatedIncome);
	//}
	for _, item := range items {
		estimatedIncome := item.EstimatedIncome
		runRatio := float64(time.Now().Unix()-item.CreateTime) / float64(item.RunDays*86400)
		gainedIncome := estimatedIncome * runRatio
		item.GainedIncome = math.Min(gainedIncome, estimatedIncome)
	}
	return connect.NewResponse(&mproto.MinerOrderListResponse{List: items}), nil
}

func returnMsg(msg string) (*connect.Response[mproto.MsgResponse], error) {
	return connect.NewResponse(&mproto.MsgResponse{Message: msg}), nil
}
func (UserManServer) LeaseMiner(ctx context.Context, c *connect.Request[mproto.LeaseMinerRequest]) (*connect.Response[mproto.MsgResponse], error) {
	user := GetCtxUser(ctx)
	minersId := int64(c.Msg.MinersId)
	num := int64(c.Msg.Num)
	fundPasswordcd := c.Msg.FundPassword
	exchangeCode := c.Msg.ExchangeCode

	if num <= 0 {
		return nil, errors.New("请输入正确的矿机租赁数量")
	}
	if !checkfundPwd(user, fundPasswordcd) {
		return nil, errors.New("资金密码错误")
	}
	miners := new(model.Miners)
	err := utils.Orm.First(miners, minersId).Error
	if err != nil {
		return nil, errors.New("矿机id不存在")
	}
	if miners.Status == 0 {
		return nil, errors.New("该矿机已暂停租赁")
	}
	purchasedNums := int64(0)
	utils.Orm.Model(model.MinersOrder{}).Where("user_id = ? and miners_id = ?", user.Id, minersId).Count(&purchasedNums)

	if num > miners.BuyLimit || (purchasedNums+num) > miners.BuyLimit {
		return nil, errors.New("该矿机限购 " + strconv.Itoa(int(miners.BuyLimit)) + " 台")
	}
	leftNum := miners.IssuesNum - miners.SalesNum
	if num > leftNum {
		return nil, errors.New("矿机数量不足，目前可租赁数： " + strconv.Itoa(int(leftNum)))
	}
	orderNo := "MO" + time.Now().Format("20060102") + fmt.Sprintf("%05d", rand.Intn(99999))
	totalPrice := miners.Price * float64(num)
	realPay := totalPrice
	discountInfo := new(model.MinersOrder)

	tx := utils.Orm.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if exchangeCode != "" {
		exchange := new(model.MinersExchange)
		err := utils.Orm.First(exchange, "code=? and miners_id=? and status=1 ", exchangeCode, minersId).Error
		if err != nil {
			return nil, errors.New("兑换码无效")
		}
		if exchange.UsedNum >= exchange.TotalNum {
			return nil, errors.New("该兑换码已使用")
		}
		if num > exchange.TotalNum {
			return nil, errors.New("该兑换码最多可以兑换 " + strconv.Itoa(int(exchange.TotalNum)) + " 台矿机")
		}
		discountRatio := exchange.DiscountRatio
		if discountRatio > 0 {
			discountAmount := totalPrice * (discountRatio / 100)
			realPay = totalPrice - discountAmount
			discountInfo.ExchangeCode = exchangeCode
			discountInfo.DiscountRatio = discountRatio
			discountInfo.DiscountAmount = discountAmount
			exchange.UsedNum += num
			exchange.UserId = user.Id
			exchange.OrderNo = orderNo
			if err := tx.Updates(exchange).Error; err != nil {
				tx.Rollback()
				return nil, err
			}
		}
	}
	order := model.MinersOrder{
		MinersId:         minersId,
		UserId:           user.Id,
		Refereeid:        user.Refereeid,
		TeamLeaderId:     user.TeamLeaderId,
		SettlementCoinId: miners.SettlementCoinId,
		ProduceCoinId:    miners.ProduceCoinId,
		OrderNo:          orderNo,
		Price:            miners.Price,
		Num:              num,
		TotalPrice:       totalPrice,
		RealPay:          realPay,
		EstimatedIncome:  float64(miners.GenIncome) * float64(num),
		PendingIncome:    float64(miners.GenIncome) * float64(num),
		RunMinutes:       miners.RunDays * 24 * 60,
		RunDays:          miners.RunDays,
		ExpireTime:       time.Now().AddDate(0, 0, int(miners.RunDays)).Unix(),
		Status:           "1",
	}
	order.ExchangeCode = discountInfo.ExchangeCode
	order.DiscountRatio = discountInfo.DiscountRatio
	order.DiscountAmount = discountInfo.DiscountAmount
	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := UpdateCoinAssetsBalance(tx, user.Id, miners.SettlementCoinId, -realPay, "lease_miners", order.Id, 0, ""); err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := tx.Updates(&model.Miners{Id: miners.Id, SalesNum: miners.SalesNum + 1}).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()
	return connect.NewResponse(&mproto.MsgResponse{Message: "租赁成功"}), nil
}

func (UserManServer) UpdatePwd(ctx context.Context, c *connect.Request[mproto.UpdatePwdRequest]) (*connect.Response[mproto.UpdatePwdResponse], error) {
	//TODO implement me
	panic("implement me")
}

func (UserManServer) InitVerifyFace(ctx context.Context, c *connect.Request[mproto.InitVerifyFaceRequest]) (*connect.Response[httpbody.HttpBody], error) {
	res, err := utils.InitFaceVerify(c.Msg.CertName, c.Msg.CertNo, c.Msg.ReturnUrl, c.Msg.MetaInfo)
	if err != nil {
		return nil, err
	}
	data, _ := json.Marshal(res)
	uid := GetCtxUserId(ctx)
	if *res.Code == "200" {
		utils.Orm.Updates(&model.User{Id: uid, Idcard: c.Msg.CertNo, Name: c.Msg.CertName, CertifyId: *res.ResultObject.CertifyId})
	}
	body := &httpbody.HttpBody{
		Data: data,
	}
	return connect.NewResponse(body), nil
}

func (UserManServer) SaveIdCardInfo(ctx context.Context, c *connect.Request[mproto.IdCardInfo]) (*connect.Response[mproto.MsgResponse], error) {
	if c.Msg.Idcard == "" || c.Msg.Name == "" {
		return returnMsg("身份证号和姓名不能为空")
	}
	uid := GetCtxUserId(ctx)
	err := utils.Orm.Updates(&model.User{Id: uid, Idcard: c.Msg.Idcard, Name: c.Msg.Name}).Error
	if err != nil {
		return nil, err
	}

	//php authGive
	cfg := LoadDbConfig().NameValues
	//目前只送USDT
	//taskGiveCoinType := cfg["task_give_coin_type"]
	taskGiveCoinType := int64(1)
	authGiveCoinNum, _ := strconv.ParseFloat(cfg["auth_give_coin_num"], 64)
	if authGiveCoinNum > 0 {
		coinChange := new(model.UserCoinChange)
		if err := utils.Orm.Where("user_id = ? AND type = ?", uid, "auth_give").First(coinChange).Error; err == gorm.ErrRecordNotFound {
			if err := UpdateCoinAssetsBalance(utils.Orm, uid, taskGiveCoinType, authGiveCoinNum, "auth_give", 0, 0, ""); err != nil {
				log.Println("实名认证返现 err:", err)
				//return nil, err
			}
		}
	}

	return returnMsg("身份证信息已保存")
}
func (UserManServer) GetVerifyFaceRes(ctx context.Context, c *connect.Request[mproto.GetVerifyFaceResRequest]) (*connect.Response[httpbody.HttpBody], error) {
	res, err := utils.GetAliFaceVerifyRes(c.Msg.CertifyId)
	if err != nil {
		return nil, err
	}

	verifyed := int64(3)
	if *res.Code == "200" {
		verifyed = 2
	}
	utils.Orm.Updates(&model.User{Id: GetCtxUserId(ctx), CertifyId: c.Msg.CertifyId, IsCertified: verifyed})

	data, _ := json.Marshal(res)
	body := &httpbody.HttpBody{
		Data: data,
	}
	return connect.NewResponse(body), nil
}

// /源api /api/user_assets/assetsInfo
func (as UserManServer) AssetBalanceList(c context.Context, req *connect.Request[mproto.NullMsg]) (*connect.Response[mproto.AssetBalanceListResponse], error) {
	uid := GetCtxUserId(c)
	sql := `
select coin.id,
       coin.name,
       coin.logo_image,
       (case
            when coin.id = 1 then ucoin.balance + (select money from ba_user where id = ?)
            else ucoin.balance end)                                                                            balance,
       (case when coin.id = 1 then (select value from ba_config where name = 'main_coin_price') else initial_price end) as price,
       (case
            when coin.id = 1 then 1000000
            else coin.margin end)                                                                              margin
from ba_coin coin
         left join ba_user_assets ucoin on coin.id = ucoin.coin_id
where ucoin.user_id = ? order by  margin desc`
	items := []*mproto.AssetBalance{}
	err := utils.Orm.Raw(sql, uid, uid).Find(&items).Error
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&mproto.AssetBalanceListResponse{List: items}), nil
}

func (UserManServer) UserInfoLevel(c context.Context, req *connect.Request[mproto.NullMsg]) (*connect.Response[mproto.UserInfoLevelResponse], error) {
	res, _ := utils.CacheFromLru(1, "UserInfoLevel"+GetCtxUserIdStr(c), 1000, func() (interface{}, error) {
		ulr := new(mproto.UserInfoLevelResponse)
		user := GetCtxUser(c)
		sql := `
select count(1) team_nums, 
   sum(case when team_level = 1 then 1 else 0 end) referee_nums
from ba_team_user
where pid = ? and is_active=1`
		tuser := new(model.User)
		utils.Orm.Raw(sql, user.Id).Scan(tuser)
		user.TeamNums = tuser.TeamNums
		user.RefereeNums = tuser.RefereeNums
		uinfo := Convert2UserInfo(user)
		ulr.UserInfo = uinfo

		level := new(model.UserLevel)
		utils.Orm.First(level, user.Level)
		nextlevel := new(model.UserLevel)
		utils.Orm.First(nextlevel, user.Level+1)
		if nextlevel.Id > 0 {
			ulr.NextRefereeNums = nextlevel.RefereeNum
			ulr.NextTeamNums = nextlevel.TeamNum
			ulr.RefereeNumsDiff = (nextlevel.RefereeNum - level.RefereeNum)
			ulr.TeamNumsDiff = (nextlevel.TeamNum - level.TeamNum)
		}
		teamLeader := GetUserById(user.TeamLeaderId)
		ulr.TeamLeaderMobile = teamLeader.Mobile
		return ulr, nil
	})

	//todo add identifierHash
	return connect.NewResponse(res.(*mproto.UserInfoLevelResponse)), nil
}

// 添加addcard 对应旧项目的 /api/financial_card/add
func (UserManServer) AddCard(c context.Context, req *connect.Request[mproto.AddCardRequest]) (*connect.Response[mproto.NullMsg], error) {
	user := GetCtxUser(c)
	if len(user.Idcard) == 0 {
		return nil, connect.NewError(500, errors.New("请先完成实名认证"))
	}
	err := utils.Orm.First(new(model.FinancialCard), model.FinancialCard{UserId: user.Id, FinancialBankId: req.Msg.BankId, BankNum: req.Msg.BankNum}).Error
	if err == nil {
		return nil, errors.New("该银行卡已存在，请勿重复绑定")
	}
	//todo check 资金密码错误
	//todo update card status to 0
	utils.Orm.Create(&model.FinancialCard{
		UserId:          user.Id,
		FinancialBankId: req.Msg.BankId,
		AccountName:     req.Msg.AccountName,
		BankNum:         req.Msg.BankNum,
		Status:          "1",
	})
	utils.Orm.Exec(`update ba_user user
set card_count=(select count(1) from ba_financial_card t where t.user_id = user.id)
where id = ?`, user.Id)
	utils.Orm.Model(user).Updates(model.User{LimitWithdrawTime: time.Now().Add(24 * time.Hour).Unix()})
	return connect.NewResponse(&mproto.NullMsg{}), nil
}

// 源api /api/index/levelInfo
func (as UserManServer) LevelInfo(c context.Context, req *connect.Request[mproto.NullMsg]) (*connect.Response[mproto.LevelInfoResponse], error) {
	list := []*model.UserLevel{}
	utils.Orm.Find(&list, "level>?", 0)
	var dayCount, weekCount, monthCount int64
	now := time.Now()
	userId := GetCtxUserId(c)
	utils.Orm.Model(model.User{}).Where(model.User{Refereeid: userId, IsActivation: 1}).Where("create_time > ?", getDayStart()).Count(&dayCount)
	utils.Orm.Model(model.User{}).Where(model.User{Refereeid: userId, IsActivation: 1}).Where("create_time between ? and ?", getWeekStart(), now).Count(&weekCount)
	utils.Orm.Model(model.User{}).Where(model.User{Refereeid: userId, IsActivation: 1}).Where("create_time between ? and ?", getMonthStart(), now).Count(&monthCount)

	cfg1 := &[]*mproto.InviterReachcfg{}
	cfg2 := &[]*mproto.InviterReachcfg{}
	cfg3 := &[]*mproto.InviterReachcfg{}
	json.Unmarshal([]byte(LoadDbConfig().Names["today_invite_reached_give"].Value), cfg1)
	json.Unmarshal([]byte(LoadDbConfig().Names["week_invite_reached_give"].Value), cfg2)
	json.Unmarshal([]byte(LoadDbConfig().Names["month_invite_reached_give"].Value), cfg3)

	return connect.NewResponse(&mproto.LevelInfoResponse{LevelList: list,
		TodayInviteCount:       uint32(dayCount),
		WeekInviteCount:        uint32(weekCount),
		MonthInviteCount:       uint32(monthCount),
		TodayInviteReachedGive: *cfg1,
		WeekInviteReachedGive:  *cfg2,
		MonthInviteReachedGive: *cfg3,
	}), nil
}

func getWeekStart() time.Time {
	now := time.Now()
	year, week := now.ISOWeek()
	// ISOWeek returns the year and week number according to ISO 8601
	// The first day of the week is Monday
	startOfWeek := time.Date(year, 1, 1, 0, 0, 0, 0, now.Location())
	for startOfWeek.Weekday() != time.Monday {
		startOfWeek = startOfWeek.AddDate(0, 0, 1)
	}
	return startOfWeek.AddDate(0, 0, (week-1)*7)
}
func getMonthStart() time.Time {
	now := time.Now()
	year, month, _ := now.Date()
	location := now.Location()
	return time.Date(year, month, 1, 0, 0, 0, 0, location)
}
func getDayStart() time.Time {
	now := time.Now()
	year, month, day := now.Date()
	location := now.Location()
	return time.Date(year, month, day, 0, 0, 0, 0, location)
}

func Notify2Handler(w http.ResponseWriter, r *http.Request) {
	memberid := r.PostFormValue("memberid")
	orderid := r.PostFormValue("orderid")
	amount := r.PostFormValue("amount")
	transactionID := r.PostFormValue("transaction_id")
	datetime := r.PostFormValue("datetime")
	returncode := r.PostFormValue("returncode")
	sign := r.PostFormValue("sign")

	log.Printf("快捷支付回调数据 ：%v", r.PostForm)

	tx := utils.Orm.Begin()
	defer tx.Rollback()

	var recharge model.FinancialRecharge
	if err := tx.Where("order_no = ? AND status = 0", orderid).First(&recharge).Error; err != nil {
		log.Println("Recharge not found or already processed")
		http.Error(w, "err", http.StatusInternalServerError)
		return
	}

	var method model.FinancialPaymentMethod
	if err := tx.First(&method, recharge.FinancialPaymentMethodId).Error; err != nil {
		log.Println("Payment method not found")
		http.Error(w, "err", http.StatusInternalServerError)
		return
	}

	params := map[string]string{
		"memberid":       memberid,
		"orderid":        orderid,
		"amount":         amount,
		"transaction_id": transactionID,
		"datetime":       datetime,
		"returncode":     returncode,
	}

	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var paramsArray []string
	for _, k := range keys {
		if v := params[k]; v != "" {
			paramsArray = append(paramsArray, fmt.Sprintf("%s=%s", k, v))
		}
	}

	paramsStr := strings.Join(paramsArray, "&")
	checkSign := strings.ToUpper(utils.GetMd5String(fmt.Sprintf("%s&key=%s", paramsStr, method.EncryptionKey)))

	if sign == checkSign {
		if returncode == "00" {
			recharge.Status = "1"
			if err := tx.Save(&recharge).Error; err != nil {
				log.Println("Failed to update recharge status")
				http.Error(w, "err", http.StatusInternalServerError)
				return
			}

			if err := UpdateCoinAssetsBalance(tx, recharge.UserId, 1, float64(recharge.MainCoinNum), "financial_recharge", recharge.Id, 0, ""); err != nil {
				log.Println("Failed to update main coin assets balance")
				http.Error(w, "err", http.StatusInternalServerError)
				return
			}

			newCardWithdrawalInterval := 24
			limitWithdrawTime := time.Now().Add(time.Duration(newCardWithdrawalInterval) * time.Hour).Unix()
			if err := tx.Model(&model.User{}).Where("id = ?", recharge.UserId).Update("limit_withdraw_time", limitWithdrawTime).Error; err != nil {
				log.Println("Failed to update limit withdraw time")
				http.Error(w, "err", http.StatusInternalServerError)
				return
			}
		} else {
			recharge.Status = "2"
			if err := tx.Save(&recharge).Error; err != nil {
				log.Println("Failed to update recharge status")
				http.Error(w, "err", http.StatusInternalServerError)
				return
			}
		}
	}
	tx.Commit()

	if recharge.Status == "1" && sign == checkSign && returncode == "00" {
		//utils.QueuePush("userFirstRecharge", map[string]interface{}{"user_id": recharge.UserId}, "reward")
		//utils.QueuePush("firstRechargeReachedGive", map[string]interface{}{"user_id": recharge.UserId, "amount": recharge.Amount}, "task_reward")
		//utils.QueuePush("todayRechargeReachedGive", map[string]interface{}{"user_id": recharge.UserId}, "task_reward")
		//utils.QueuePush("giveLotteryCount", map[string]interface{}{"user_id": recharge.UserId, "coinAmount": recharge.MainCoinNum}, "reward")
	}

	w.Write([]byte("ok"))
}

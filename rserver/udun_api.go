// File: `address.go`
package rserver

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"muskex/gen/mproto/model"
	"muskex/utils"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// createAddressConfig holds the static configuration for the request
var createAddressConfig struct {
	Key         string
	URL         string
	RequestBody createAddressRequestBody
}

// createAddressRequestBody represents the body of the create address request
type createAddressRequestBody struct {
	MerchantId   string `json:"merchantId"`
	MainCoinType int    `json:"mainCoinType"`
	CallUrl      string `json:"callUrl"`
	Alias        string `json:"alias,omitempty"`
}

// UdunRequest represents the full create address request
type UdunRequest struct {
	Timestamp string `json:"timestamp"`
	Nonce     string `json:"nonce"`
	Sign      string `json:"sign"`
	Body      string `json:"body"`
}

// CreateAddressResponse represents the response from the create address API
type CreateAddressResponse struct {
	Data struct {
		CoinType int    `json:"coinType"`
		Address  string `json:"address"`
	} `json:"data"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}
type UdunResponse struct {
	Data    json.RawMessage `json:"data"`
	Message string          `json:"message"`
	Code    int             `json:"code"`
}

// InitializeCreateReqConfig initializes the static configuration
func InitializeCreateReqConfig(key, url, merchantId string, mainCoinType int, callUrl string) {
	createAddressConfig.Key = key
	createAddressConfig.URL = url
	createAddressConfig.RequestBody = createAddressRequestBody{
		MerchantId:   merchantId,
		MainCoinType: mainCoinType,
		CallUrl:      callUrl,
	}
}

// generateMD5Signature generates the MD5 signature for the request
func generateMD5Signature(body, key, nonce, timestamp string) string {
	signatureString := body + key + nonce + timestamp
	hash := md5.Sum([]byte(signatureString))
	return strings.ToLower(hex.EncodeToString(hash[:]))
}

// CreateAddress sends a request to create a new address
func CreateAddress(alias string) (*CreateAddressResponse, error) {
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	nonce := fmt.Sprintf("%d", time.Now().UnixNano())

	requestBody := createAddressConfig.RequestBody
	requestBody.Alias = alias

	bodyBytes, err := json.Marshal([]createAddressRequestBody{requestBody})
	if err != nil {
		return nil, err
	}
	bodyString := string(bodyBytes)

	sign := generateMD5Signature(bodyString, createAddressConfig.Key, nonce, timestamp)

	request := UdunRequest{
		Timestamp: timestamp,
		Nonce:     nonce,
		Sign:      sign,
		Body:      bodyString,
	}

	requestBytes, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(createAddressConfig.URL, "application/json", bytes.NewBuffer(requestBytes))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to create address")
	}

	responseBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response UdunResponse
	err = json.Unmarshal(responseBytes, &response)
	if err != nil {
		return nil, err
	}
	// Error codes and their corresponding messages
	errorMessages := map[int]string{
		-1:   "生成地址失敗",
		200:  "生成地址成功",
		4001: "商户不存在",
		4005: "非法參數",
		4045: "幣種信息錯誤",
		4162: "簽名異常",
		4163: "簽名錯誤",
		4166: "商戶沒有配置套餐",
		4168: "商戶地址達到上限",
		4169: "商戶已禁用",
		4175: "錢包編號錯誤",
		4017: "商戶沒有創建錢包",
		4176: "錢包未添加支持該幣種",
		4188: "暫不支持",
		4226: "商戶普通賬戶被禁用",
		4261: "商戶管理員賬戶被禁用",
		4262: "賬戶不存在",
	}

	// Check the response code and return an error if it's not 200
	if response.Code != 200 {
		if msg, exists := errorMessages[response.Code]; exists {
			return nil, errors.New(msg)
		}
		return nil, fmt.Errorf("unknown error code: %d %s %s", response.Code, response.Message, string(response.Data))
	} else {
		var createAddressResponse CreateAddressResponse
		err = json.Unmarshal(responseBytes, &createAddressResponse)
		if err != nil {
			return nil, err
		}
		return &createAddressResponse, nil
	}
}

// CallbackRequest represents the structure of the callback request body
// Status回調接口狀態說明
// 狀態	說明
// 0	待審核
// 1	審核成功
// 2	審核駁回
// 3	交易成功
// 4	交易失敗
// TradeType 回調接口交易類型說明
// 狀態	說明
// 1	充幣回調
// 2	提幣回調
type CallbackRequest struct {
	Address      string `json:"address"`
	Amount       string `json:"amount"`
	BlockHigh    string `json:"blockHigh"`
	CoinType     string `json:"coinType"`
	Decimals     string `json:"decimals"`
	Fee          string `json:"fee"`
	MainCoinType string `json:"mainCoinType"`
	Status       int    `json:"status"`
	TradeId      string `json:"tradeId"`
	TradeType    int    `json:"tradeType"`
	TxId         string `json:"txId"`
	BusinessId   string `json:"businessId"`
	Memo         string `json:"memo"`
}

// TradeCallbackHandler handles the callback requests
func TradeCallbackHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	timestamp := r.FormValue("timestamp")
	nonce := r.FormValue("nonce")
	sign := r.FormValue("sign")
	body := r.FormValue("body")

	if timestamp == "" || nonce == "" || sign == "" || body == "" {
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
		return
	}

	var callbackReq = CallbackRequest{}
	err = json.Unmarshal([]byte(body), &callbackReq)
	if err != nil {
		http.Error(w, "Invalid body format", http.StatusBadRequest)
		return
	}
	// Process the callback request
	//fmt.Fprintf(w, "Callback received: %+v", callbackReq)

	if callbackReq.TradeType == 1 {
		//充幣回調
		log.Printf("充幣回調: %+v", callbackReq)
		if callbackReq.Status == 3 && callbackReq.MainCoinType == "195" && callbackReq.CoinType == "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t" {
			rechargeArrives(&callbackReq)
		} else {

		}
	} else if callbackReq.TradeType == 2 {
		//提幣回調
		log.Printf("提幣回調: %+v", callbackReq)
		if callbackReq.Status == 1 {
			withdrawPassed(&callbackReq)
		}
	}
	//0	待審核
	//1	審核成功
	//2	審核駁回
	//3	交易成功
	//4	交易失敗

}

func withdrawPassed(body *CallbackRequest) {
	businessId := body.BusinessId
	utils.Orm.Where(model.FinancialWithdraw{OrderNo: businessId, Status: "4"}).Updates(&model.FinancialWithdraw{OrderNo: businessId, Status: "3"})
}
func rechargeArrives(body *CallbackRequest) {
	amount, err := strconv.ParseFloat(body.Amount, 64)
	if err != nil {
		log.Println("Failed to parse amount:", err)
		return
	}
	decimal, _ := strconv.ParseFloat(body.Decimals, 64)
	amount = amount / math.Pow(10, decimal)

	cfg := LoadDbConfig().NameValues
	uMinRecharge, _ := strconv.ParseFloat(cfg["u_min_recharge"], 64)
	if amount < uMinRecharge {
		log.Println("Recharge amount is less than the minimum:", amount, uMinRecharge)
		return
	}

	tx := utils.Orm.Begin()
	defer tx.Rollback()
	// Find the user's assets by address
	userAssets := &model.UserAssets{}

	err = tx.First(userAssets, "address=?", body.Address).Error
	if err != nil {
		log.Println("Failed to find assets:", body.Address, body.MainCoinType, body.CoinType, err)
		return
	}

	// Create a recharge record
	recharge := &model.CoinRecharge{
		UserId:       userAssets.UserId,
		TradeId:      body.TradeId,
		Amount:       amount,
		Address:      body.Address,
		MainCoinType: body.MainCoinType,
		TxId:         body.TxId,
	}
	err = tx.Create(recharge).Error
	if err != nil {
		log.Println("Failed to create recharge record:", err)
		return
	}

	// Update the user's main coin assets balance
	err = UpdateCoinAssetsBalance(tx, userAssets.UserId, 1, amount, "recharge_coin", recharge.Id, 0, "")
	if err != nil {
		log.Println("Failed to update main coin assets balance:", err)
		return
	}
	// Apply the recharge give ratio if applicable
	uRechargeGiveRatio, _ := strconv.ParseFloat(cfg["u_recharge_give_ratio"], 64)
	if uRechargeGiveRatio > 0 {
		giveAmount := amount * (uRechargeGiveRatio / 100)
		err = UpdateCoinAssetsBalance(tx, userAssets.UserId, 1, giveAmount, "u_recharge", recharge.Id, 0, "")
		if err != nil {
			log.Println("Failed to update main coin assets balance for give amount:", err)
			return
		}
	}
	// Update the user's limit withdraw time
	newCardWithdrawalInterval := 24
	limitWithdrawTime := time.Now().Add(time.Duration(newCardWithdrawalInterval) * time.Hour)
	err = tx.Updates(model.User{Id: userAssets.UserId, LimitWithdrawTime: limitWithdrawTime.Unix()}).Error
	if err != nil {
		log.Println("Failed to update user's limit withdraw time:", err)
		return
	}
	tx.Commit()
}

//userFirstRecharge giveLotteryCount
//curl localhost:8080/create_address_callback -v \
//-d 'timestamp=1535005047&nonce=100000&sign=e1bee3a417b9c606ba6cedda26db761a&body={"address":"DJY781Z8qbuJeuA7C3McYivbX8kmAUXPsW","amount":"12345678","blockHigh":"102419","coinType":"206","decimals":"8","fee":"452000","mainCoinType":"206","status":3,"tradeId":"20181024175416907","tradeType":1,"txId":"31689c332536b56a2246347e206fbed2d04d461a3d668c4c1de32a75a8d436f0"}'

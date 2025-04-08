package ifp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/shopspring/decimal"
	"golang.org/x/text/language"
)

const (
	BuyTimeout = 25 * time.Minute
)

type BuyCoinMode = string

const (
	// 数量模式, 默认模式
	// 当选择此模式时，需要传入 usddAmount 字段，此时会通过所购币种的数量，筛选最优广告并生成订单。
	BuyCoinMode_USDD BuyCoinMode = "UsddAmount"
	// 支付金额模式
	// 当选择此模式时，需要传入 totalPrice 字段，此时会通过支付币种与总金额数量，筛选最优广告并生成订单。
	BuyCoinMode_Fiat BuyCoinMode = "PaymentPrice"
)

// 语言
type LanguageCode = string

const (
	LanguageCode_en    LanguageCode = "en"
	LanguageCode_zh_CN LanguageCode = "zh_CN"
	LanguageCode_zh_TW LanguageCode = "zh_TW"
	LanguageCode_vi    LanguageCode = "vi"
	LanguageCode_id    LanguageCode = "id"
)

const (
	minBuyUSDDAmount = 50
)

var (
	minBuyUSDDAmountD = decimal.NewFromInt(minBuyUSDDAmount)
	ErrInvalidAmount  = errors.New("invalid amount")
)

var (
	// 顺序必须保持对应
	languageCodes = []string{LanguageCode_en, LanguageCode_zh_CN, LanguageCode_zh_TW, LanguageCode_vi, LanguageCode_id}
	langMatcher   = language.NewMatcher([]language.Tag{
		language.English,
		language.SimplifiedChinese,
		language.TraditionalChinese,
		language.Vietnamese,
		language.Indonesian,
	})
)

func getLanguageCode(lang language.Tag) LanguageCode {
	_, i, _ := langMatcher.Match(lang)
	return languageCodes[i]
}

type IFPStatusCode = string

const (
	IFPStatusCode_Success            IFPStatusCode = "SUCCESS"
	IFPStatusCode_TimestampError     IFPStatusCode = "TIMESTAMP_ERROR"
	IFPStatusCode_SignatureError     IFPStatusCode = "SIGNATURE_ERROR"
	IFPStatusCode_AccesskeyError     IFPStatusCode = "ACCESS_KEY_ERROR"
	IFPStatusCode_ParameterError     IFPStatusCode = "PARAMETER_ERROR"
	IFPStatusCode_NoAdvertisement    IFPStatusCode = "NO_ADVERTISEMENT"
	IFPStatusCode_AccountStatusError IFPStatusCode = "ACCOUNT_STATUS_ERROR"
	IFPStatusCode_SystemError        IFPStatusCode = "SYSTEM_ERROR"
	IFPStatusCode_TradeCanceled      IFPStatusCode = "TRADE_CANCELED"

	IFPStatusCode_NoOrder IFPStatusCode = "NO_ORDER"
)

type CurrencyCode = string

const (
	CurrencyCode_CNY  CurrencyCode = "CNY"
	CurrencyCode_HKD  CurrencyCode = "HKD"
	CurrencyCode_TWD  CurrencyCode = "TWD"
	CurrencyCode_VND  CurrencyCode = "VND"
	CurrencyCode_AUD  CurrencyCode = "AUD"
	CurrencyCode_USDT CurrencyCode = "USDT"
)

const (
	IFPHeaderKey_Accesskey = "access-key"
	IFPHeaderKey_Timestamp = "timestamp"
	IFPHeaderKey_Signature = "signature"
)

type IFPGenericResponse[T any] struct {
	Data       T             `json:"data"`
	StatusCode IFPStatusCode `json:"statusCode"`
	Message    string        `json:"message"`
	Success    bool          `json:"success"`
}

func (res *IFPGenericResponse[T]) IsSuccess() bool {
	return res.Success && res.StatusCode == IFPStatusCode_Success
}

type rawBuyRequest struct {
	*baseRequest
	// 买币模式
	Mode BuyCoinMode `json:"buyCoinMode"`
	// 币的数量, USDD数量模式必传
	Amount decimal.Decimal `json:"usddAmount,omitempty"` // Decimal
	// 总支付金额, 支付金额模式必传
	Price decimal.Decimal `json:"totalPrice,omitempty"` // Decimal
	// 外部订单号
	Ticket string `json:"externalOrderNumber"`
	// 买币成功后回调地址
	CallbackURL string `json:"callbackUrl"`
	// 展示语言
	Language LanguageCode `json:"supportLanguage"`
	// 所需支付币种
	Currency CurrencyCode `json:"currencyCode"`
	// 实际支付人姓名
	UserName string `json:"payerRealName"`
}

type FiatBuyRequest struct {
	Language language.Tag

	MerchantOrderID string
	Amount          decimal.Decimal
	Currency        string
	// 实际支付人姓名, kyc need
	UserName string
}
type BuyCoinReply struct {
	// 所需跳转的支付 URI
	RedirectURL string
	// 此次交易所匹配的广告编码
	AdvertisementCode string
}

func (req *FiatBuyRequest) toRaw(conf *Config) *rawBuyRequest {
	return &rawBuyRequest{
		baseRequest: newBaseRequest(),
		Mode:        BuyCoinMode_Fiat,
		Price:       req.Amount,
		Ticket:      req.MerchantOrderID,
		CallbackURL: conf.CallbackURL,
		Language:    getLanguageCode(req.Language),
		Currency:    req.Currency,
		UserName:    req.UserName,
	}
}

func (req *FiatBuyRequest) Validate() error {
	if req.MerchantOrderID == "" {
		return fmt.Errorf("merchant order id is empty")
	}
	if req.UserName == "" {
		return fmt.Errorf("user name is empty")
	}

	if req.Amount.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("amount is zero")
	}
	return nil
}

func (r *rawBuyRequest) isBuyUSDDMode() bool {
	return r.Mode == BuyCoinMode_USDD || r.Currency == CurrencyCode_USDT
}

func (r *rawBuyRequest) MarshalJSON() ([]byte, error) {
	if r.isBuyUSDDMode() {
		if r.Amount.LessThan(minBuyUSDDAmountD) {
			return nil, ErrInvalidAmount
		}
	}

	temp := map[string]any{
		"externalOrderNumber": r.Ticket,
		"callbackUrl":         r.CallbackURL,
		"supportLanguage":     r.Language,
		"currencyCode":        r.Currency,
		"payerRealName":       r.UserName,
	}

	if r.isBuyUSDDMode() {
		temp["buyCoinMode"] = BuyCoinMode_USDD
		temp["usddAmount"] = r.Amount
	} else {
		temp["buyCoinMode"] = BuyCoinMode_Fiat
		temp["totalPrice"] = r.Price
	}

	return json.Marshal(temp)
}

func (r *rawBuyRequest) GenerateSignedRequest(ctx context.Context, conf *Config) (*http.Request, error) {
	const (
		Endpoint_BuyCoin = "/api/buy-coin/transaction"
		Method           = http.MethodPost
	)
	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(r); err != nil {
		return nil, fmt.Errorf("json encode error: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, Method, Endpoint_BuyCoin, &body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(IFPHeaderKey_Accesskey, conf.AccessKey)
	req.Header.Set(IFPHeaderKey_Timestamp, r.ts)
	req.Header.Set(IFPHeaderKey_Signature, r.GenerateSignature(conf.AccessKey, conf.PrivateKey))

	return req, nil
}

type BuyResponseData struct {
	// 所需跳转的支付 URI
	RedirectURL string `json:"redirectUrl"`
	// 此次交易所匹配的广告编码
	AdvertisementCode string `json:"advertisementCode"`
	// 当前 UTC 时间戳
	CurrentTimestamp int64 `json:"currentTimestamp"`
	// 币安智能链(BSC)的收款地址
	ETH string `json:"eth"`
	// Tron链的收款地址
	TRX string `json:"trx"`
}

type BuyResponse = IFPGenericResponse[BuyResponseData]

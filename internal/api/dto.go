package api

// JSON 请求/响应结构。所有大整数用十进制字符串：JSON number 是 float64，超过 2^53 会丢精度。

// PlaceOrderRequest POST /order：客户端已经签好名的订单。
type PlaceOrderRequest struct {
	Perp         string `json:"perp"`         // 0x 地址
	Signer       string `json:"signer"`       // 0x 地址
	PaperAmount  string `json:"paperAmount"`  // 1e18 精度，带符号，例 "100000000000000000" = 0.1 张买
	CreditAmount string `json:"creditAmount"` // USDC 6 位，带符号，例 "-2600000000"
	Info         string `json:"info"`         // 0x + 64 位 hex（PackInfo 的结果）
	Signature    string `json:"signature"`    // 0x + 130 位 hex（r ‖ s ‖ v）
}

// PlaceOrderResponse 下单结果。
type PlaceOrderResponse struct {
	OrderID   string `json:"orderId"`          // digest hex
	Filled    string `json:"filled"`           // 本次成交 paper（1e18）
	Remaining string `json:"remaining"`        // 挂在簿上的剩余 paper（1e18），0 表示全部成交
	TxHash    string `json:"txHash,omitempty"` // 有成交时的上链交易
}

// LevelDTO 一档深度。
type LevelDTO struct {
	Price  string `json:"price"`  // USDC 6 位 / 1 张
	Amount string `json:"amount"` // paper 1e18
	Count  int    `json:"count"`
}

// OrderBookResponse GET /orderbook?perp=0x…&depth=N
type OrderBookResponse struct {
	Perp string     `json:"perp"`
	Bids []LevelDTO `json:"bids"`
	Asks []LevelDTO `json:"asks"`
}

// PositionResponse GET /position?trader=0x…&perp=0x…
type PositionResponse struct {
	Trader        string `json:"trader"`
	Perp          string `json:"perp"`
	Paper         string `json:"paper"`         // Perpetual.balanceOf 的 paper（1e18）
	Credit        string `json:"credit"`        // Perpetual.balanceOf 的 credit（USDC 6 位）
	PrimaryCredit string `json:"primaryCredit"` // Dealer 主账户余额（USDC 6 位）
	Safe          bool   `json:"safe"`          // Dealer.IsSafe
}

// ErrorResponse 统一的错误返回。
type ErrorResponse struct {
	Error string `json:"error"`
}

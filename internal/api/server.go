package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/xuefei016/my-perpetual-backend/internal/chain"
	"github.com/xuefei016/my-perpetual-backend/internal/chain/bindings/perpetual"
	"github.com/xuefei016/my-perpetual-backend/internal/chain/bindings/pricesource"
	"github.com/xuefei016/my-perpetual-backend/internal/engine"
)

// Server 把 HTTP 请求翻译成引擎/链上调用。
type Server struct {
	eng    *engine.Engine
	c      *chain.Chain
	dealer common.Address // 算 domainSeparator 用（verifyingContract）
}

func New(eng *engine.Engine, c *chain.Chain, dealer common.Address) *Server {
	return &Server{eng: eng, c: c, dealer: dealer}
}

var one = big.NewInt(1e18)
var errIMNotSafe = errors.New("IM not safe after fill")

const maxCancelWindow = 5 * time.Minute
const defaultDepth, maxDepth = 20, 100

// Handler 注册路由（Go 1.22+ 的 "方法 路径" 写法）。
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /order", s.handlePlaceOrder)
	mux.HandleFunc("DELETE /order/{id}", s.handleCancelOrder)
	mux.HandleFunc("GET /orderbook", s.handleOrderBook)
	mux.HandleFunc("GET /position", s.handlePosition)
	return mux
}

// handlePlaceOrder 解析 PlaceOrderRequest → 用 chain.RecoverSigner 做和合约一样的验签
// （签名人 == Signer，或是 Signer 授权的 operator）→ engine.PlaceOrder → PlaceOrderResponse。
func (s *Server) handlePlaceOrder(w http.ResponseWriter, r *http.Request) {
	var req PlaceOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	// ② 校验并转换字段 → order chain.Order
	//    地址：先 common.IsHexAddress(req.Perp / req.Signer)，再 common.HexToAddress
	//    大整数：new(big.Int).SetString(req.PaperAmount, 10)，第二个返回值 ok 要检查
	//    info：hexutil.Decode(req.Info) → []byte，len != 32 返回 400；copy(order.Info[:], b)
	//    签名：hexutil.Decode(req.Signature) → sig（长度留给 RecoverSigner 查）
	//    以上任何一个失败都返回 400
	if !common.IsHexAddress(req.Perp) {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid perp"))
		return
	}
	if !common.IsHexAddress(req.Signer) {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid signer"))
		return
	}

	perpAddress := common.HexToAddress(req.Perp)
	signerAddress := common.HexToAddress(req.Signer)

	paper, ok := new(big.Int).SetString(req.PaperAmount, 10)
	if !ok {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid paper amount"))
		return
	}

	credit, ok := new(big.Int).SetString(req.CreditAmount, 10)
	if !ok {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid credit amount"))
		return
	}
	infoBytes, err := hexutil.Decode(req.Info)
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid info"))
		return
	}

	if len(infoBytes) != 32 {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid info"))
		return
	}

	var info [32]byte
	copy(info[:], infoBytes)

	signatureBytes, err := hexutil.Decode(req.Signature)
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid signature"))
		return
	}

	// ③ 已知市场：s.eng.Snapshot(order.Perp, 0) 返回 err 就 400
	_, _, err = s.eng.Snapshot(perpAddress, 0)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	// ④ 验签
	//    digest := chain.Digest(chain.DomainSeparator(s.c.ChainId, s.dealer), order)
	//    recovered, err := chain.RecoverSigner(digest, sig)，err 返回 401
	//    recovered != order.Signer 时，再查 s.c.Dealer.IsOperatorValid(opts, order.Signer, recovered)
	//      查链出错返回 502（上游出错，不是客户端的错），false 返回 401
	//    opts := &bind.CallOpts{Context: r.Context()}
	order := chain.Order{
		Perp:         perpAddress,
		Signer:       signerAddress,
		Info:         info,
		PaperAmount:  paper,
		CreditAmount: credit,
	}
	digest := chain.Digest(chain.DomainSeparator(s.c.ChainId, s.dealer), order)
	recovered, err := chain.RecoverSigner(digest, signatureBytes)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err)
		return
	}
	if recovered != signerAddress {
		ok, err := s.c.Dealer.IsOperatorValid(&bind.CallOpts{Context: r.Context()}, order.Signer, recovered)
		if err != nil {
			writeError(w, http.StatusBadGateway, err)
			return
		}
		if !ok {
			writeError(w, http.StatusUnauthorized, fmt.Errorf("invalid signer"))
			return
		}
	}

	// ⑤ 过期：chain.UnpackInfo(order.Info).Expiration < uint64(time.Now().Unix()) 返回 400
	if chain.UnpackInfo(order.Info).Expiration < uint64(time.Now().Unix()) {
		writeError(w, http.StatusBadRequest, fmt.Errorf("expired order"))
		return
	}

	// ⑥ bo, err := engine.NewBookOrder(digest.Hex(), order, sig)，err 返回 400
	bo, err := engine.NewBookOrder(digest.Hex(), order, signatureBytes)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	// ⑦ Remaining 扣掉链上已成交量
	filledAmount, err := s.c.Dealer.GetOrderFilledAmount(&bind.CallOpts{Context: r.Context()}, common.HexToHash(bo.ID))
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	bo.Remaining = new(big.Int).Sub(bo.Remaining, filledAmount)
	if bo.Remaining.Sign() <= 0 {
		writeError(w, http.StatusConflict, fmt.Errorf("all orders are filled"))
		return
	}

	if _, err := s.checkIM(&bind.CallOpts{Context: r.Context()}, order, bo.Remaining); err != nil {
		if errors.Is(err, errIMNotSafe) {
			writeError(w, http.StatusUnprocessableEntity, err)
		} else {
			writeError(w, http.StatusBadGateway, err)
		}
		return
	}

	// ⑧ 上链用不随客户端断开而取消的 ctx
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	submittedRemaining := new(big.Int).Set(bo.Remaining)
	m, err := s.eng.PlaceOrder(ctx, bo)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}

	// ⑨ total := new(big.Int)，遍历 m.Fills 累加 f.Amount
	total := new(big.Int)

	for _, fill := range m.Fills {
		total.Add(total, fill.Amount)
	}

	// ⑩ writeJSON 200 PlaceOrderResponse
	//    OrderID: bo.ID；Filled: total.String()；Remaining: 提交前快照减本次成交量
	//    TxHash：只有 len(m.Fills) > 0 时才填 m.TxHash.Hex()（omitempty 才会把它省掉）
	placeOrderResponse := PlaceOrderResponse{
		OrderID:   bo.ID,
		Filled:    total.String(),
		Remaining: new(big.Int).Sub(submittedRemaining, total).String(), // 不读取解锁后可被其他请求修改的簿内对象
	}
	if len(m.Fills) > 0 {
		placeOrderResponse.TxHash = m.TxHash.Hex()
	}

	writeJSON(w, http.StatusOK, placeOrderResponse)
	return
}

// handleCancelOrder 撤单。谁有权撤（要不要签名）D9 再定。
func (s *Server) handleCancelOrder(w http.ResponseWriter, r *http.Request) {
	req := CancelOrderRequest{}

	// request.body校验
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	// 签名校验
	sig, err := hexutil.Decode(req.Signature)
	if err != nil || len(sig) != 65 {
		writeError(w, http.StatusBadRequest, errors.New("invalid signature"))
		return
	}

	// 解析orderId
	raw, err := hexutil.Decode(r.PathValue("id"))
	if err != nil || len(raw) != 32 {
		writeError(w, http.StatusBadRequest, errors.New("invalid orderId"))
		return
	}
	id := common.BytesToHash(raw).Hex()

	// perp校验+获取bookOrder
	if !common.IsHexAddress(req.Perp) {
		writeError(w, http.StatusBadRequest, errors.New("invalid perp"))
		return
	}
	perp := common.HexToAddress(req.Perp)
	bo, ok := s.eng.Order(perp, id)
	if !ok {
		writeError(w, http.StatusNotFound, fmt.Errorf("order not found"))
		return
	}

	// deadline校验
	now := time.Now().Unix()
	if req.Deadline < now || req.Deadline > now+int64(maxCancelWindow/time.Second) {
		writeError(w, http.StatusBadRequest, errors.New("invalid deadline"))
		return
	}
	// 验签
	hash := accounts.TextHash([]byte(fmt.Sprintf("cancel %s %d", id, req.Deadline)))
	if sig[64] >= 27 {
		sig[64] -= 27
	}
	pub, err := crypto.SigToPub(hash, sig)
	if err != nil {
		writeError(w, http.StatusBadRequest, errors.New("invalid signature"))
		return
	}

	ok, err = s.authorized(&bind.CallOpts{Context: r.Context()}, bo.Order.Signer, crypto.PubkeyToAddress(*pub))
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	if !ok {
		writeError(w, http.StatusUnauthorized, fmt.Errorf("not order owner"))
		return
	}

	// 查到之后、撤之前可能刚被别的 taker 吃完：引擎报 not found，按 404 返回
	removed, err := s.eng.CancelOrder(perp, id)
	if err != nil {
		writeError(w, http.StatusNotFound, fmt.Errorf("cancel order: %w", err))
		return
	}

	writeJSON(w, http.StatusOK, CancelOrderResponse{
		OrderID:   bo.ID,
		Remaining: removed.Remaining.String(),
	})
}

// handleOrderBook 返回 engine.Snapshot 的深度。
func (s *Server) handleOrderBook(w http.ResponseWriter, r *http.Request) {
	perp := r.URL.Query().Get("perp")
	if !common.IsHexAddress(perp) {
		writeError(w, http.StatusBadRequest, errors.New("invalid perp"))
		return
	}
	perpAddress := common.HexToAddress(perp)
	d := defaultDepth
	if s := r.URL.Query().Get("depth"); s != "" {
		v, err := strconv.Atoi(s)
		if err != nil || v < 1 || v > maxDepth {
			writeError(w, http.StatusBadRequest, errors.New("invalid depth"))
			return
		}
		d = v
	}

	bids, asks, err := s.eng.Snapshot(perpAddress, d)
	if err != nil { // 引擎只会因为未知市场报错
		writeError(w, http.StatusNotFound, err)
		return
	}

	writeJSON(w, http.StatusOK, OrderBookResponse{
		Perp: perpAddress.Hex(),
		Bids: toLevelDTOs(bids),
		Asks: toLevelDTOs(asks),
	})
}

// handlePosition 读链上 balanceOf / GetPrimaryCredit / IsSafe。
func (s *Server) handlePosition(w http.ResponseWriter, r *http.Request) {
	perp := r.URL.Query().Get("perp")
	if !common.IsHexAddress(perp) {
		writeError(w, http.StatusBadRequest, errors.New("invalid perp"))
		return
	}
	perpAddress := common.HexToAddress(perp)

	trader := r.URL.Query().Get("trader")
	if !common.IsHexAddress(trader) {
		writeError(w, http.StatusBadRequest, errors.New("invalid trader"))
		return
	}
	traderAddress := common.HexToAddress(trader)

	if _, _, err := s.eng.Snapshot(perpAddress, 0); err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}

	p, err := perpetual.NewPerpetual(perpAddress, s.c.Backend)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}

	opts := &bind.CallOpts{Context: r.Context()}
	bal, err := p.BalanceOf(opts, traderAddress)
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}

	primaryCredit, err := s.c.Dealer.GetPrimaryCredit(opts, traderAddress)
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}

	isSafe, err := s.c.Dealer.IsSafe(opts, traderAddress)
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}

	writeJSON(w, http.StatusOK, PositionResponse{
		Trader:        traderAddress.Hex(),
		Perp:          perpAddress.Hex(),
		Paper:         bal.Paper.String(),
		Credit:        bal.Credit.String(),
		PrimaryCredit: primaryCredit.String(),
		Safe:          isSafe,
	})
}

func (s *Server) checkIM(opts *bind.CallOpts, o chain.Order, remaining *big.Int) (ok bool, err error) {
	risk, err := s.c.Dealer.GetTraderRisk(opts, o.Signer)
	if err != nil {
		return false, fmt.Errorf("get trader risk: %w", err)
	}

	riskParam, err := s.c.Dealer.GetRiskParams(opts, o.Perp)
	if err != nil {
		return false, fmt.Errorf("get trader risk params: %w", err)
	}

	ps, err := pricesource.NewPriceSource(riskParam.MarkPriceSource, s.c.Backend)
	if err != nil {
		return false, fmt.Errorf("get pricesource: %w", err)
	}
	mark, err := ps.GetMarkPrice(opts)
	if err != nil {
		return false, fmt.Errorf("get trader mark price: %w", err)
	}

	perp, err := perpetual.NewPerpetual(o.Perp, s.c.Backend)
	if err != nil {
		return false, fmt.Errorf("new perpetual: %w", err)
	}
	bal, err := perp.BalanceOf(opts, o.Signer)
	if err != nil {
		return false, fmt.Errorf("get balance of perpetual: %w", err)
	}

	oldPaper := bal.Paper
	fillPaper := new(big.Int).Mul(remaining, big.NewInt(int64(o.PaperAmount.Sign())))
	fillCredit := new(big.Int).Mul(remaining, o.CreditAmount)
	fillCredit.Quo(fillCredit, new(big.Int).Abs(o.PaperAmount))

	newPaper := new(big.Int).Add(oldPaper, fillPaper)
	if oldPaper.Sign()*newPaper.Sign() >= 0 && oldPaper.CmpAbs(newPaper) >= 0 && oldPaper.Sign() != 0 {
		return true, nil
	}

	pnl := new(big.Int).Mul(fillPaper, mark)
	pnl.Quo(pnl, one)
	pnl.Add(pnl, fillCredit)
	newNet := new(big.Int).Add(risk.NetValue, pnl)
	diffIM := new(big.Int).Sub(im(newPaper, mark, riskParam.InitialMarginRatio), im(oldPaper, mark, riskParam.InitialMarginRatio))
	newIM := new(big.Int).Add(risk.InitialMargin, diffIM)
	if newNet.Cmp(newIM) < 0 {
		return false, fmt.Errorf("%w: net %s < im %s", errIMNotSafe, newNet, newIM)
	}
	return newNet.Cmp(newIM) >= 0, nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Println("Error writing response:", err)
	}
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, ErrorResponse{Error: err.Error()})
}

func im(paper, mark, imr *big.Int) *big.Int {
	v := new(big.Int).Mul(paper, mark)
	v.Quo(v, one)
	v.Abs(v)
	v.Mul(v, imr)
	return v.Quo(v, one)
}

// authorized 判断 recovered 能不能代表 signer：本人，或者 signer 授权过的 operator。
func (s *Server) authorized(opts *bind.CallOpts, signer, recovered common.Address) (bool, error) {
	if recovered == signer {
		return true, nil
	}
	ok, err := s.c.Dealer.IsOperatorValid(opts, signer, recovered)
	if err != nil {
		return false, fmt.Errorf("check operator: %w", err)
	}
	return ok, nil
}

func toLevelDTO(l engine.Level) LevelDTO {
	return LevelDTO{
		Price:  l.Price.String(),
		Amount: l.Amount.String(),
		Count:  l.Count,
	}
}

func toLevelDTOs(s []engine.Level) []LevelDTO {
	m := make([]LevelDTO, 0, len(s))
	for l := range s {
		m = append(m, toLevelDTO(s[l]))
	}
	return m
}

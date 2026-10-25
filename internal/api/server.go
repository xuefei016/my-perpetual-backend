package api

import (
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/xuefei016/my-perpetual-backend/internal/chain"
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
	panic("TODO(D9): handlePlaceOrder")
}

// handleCancelOrder 撤单。谁有权撤（要不要签名）D9 再定。
func (s *Server) handleCancelOrder(w http.ResponseWriter, r *http.Request) {
	panic("TODO(D9): handleCancelOrder")
}

// handleOrderBook 返回 engine.Snapshot 的深度。
func (s *Server) handleOrderBook(w http.ResponseWriter, r *http.Request) {
	panic("TODO(D9): handleOrderBook")
}

// handlePosition 读链上 balanceOf / GetPrimaryCredit / IsSafe。
func (s *Server) handlePosition(w http.ResponseWriter, r *http.Request) {
	panic("TODO(D9): handlePosition")
}

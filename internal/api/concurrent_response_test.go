package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http/httptest"
	"sync"
	"testing"
)

// 每个 POST 的 filled + remaining 应等于它提交时可用的数量，
// 后续请求的成交不能改变本次响应；多个请求会同时吃掉彼此的挂单。
func TestPlaceOrderConcurrentResponses(t *testing.T) {
	env := newTestEnv(t)
	alice, bob := newTrader(t), newTrader(t)
	const pairs = 128
	bodies := make([][]byte, 0, pairs*2)
	for i := 0; i < pairs; i++ {
		for side, actor := range []trader{alice, bob} {
			paper := amt(1)
			if side == 1 {
				paper.Neg(paper)
			}
			req, _ := sign(t, actor, orderSpec{signer: actor.addr, paper: paper, credit: at26000(paper), nonce: uint64(i*2 + side + 1)}, testChainID)
			body, err := json.Marshal(req)
			if err != nil {
				t.Fatal(err)
			}
			bodies = append(bodies, body)
		}
	}
	jobs := make(chan []byte, len(bodies))
	for _, body := range bodies {
		jobs <- body
	}
	close(jobs)
	problems := make(chan error, len(bodies))
	var wg sync.WaitGroup
	start := make(chan struct{})
	for worker := 0; worker < 16; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for body := range jobs {
				rec := httptest.NewRecorder()
				env.h.ServeHTTP(rec, httptest.NewRequest("POST", "/order", bytes.NewReader(body)))
				var result PlaceOrderResponse
				if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil || rec.Code != 200 {
					problems <- fmt.Errorf("POST status=%d body=%s decode=%v", rec.Code, rec.Body.String(), err)
					continue
				}
				filled, okFilled := new(big.Int).SetString(result.Filled, 10)
				remaining, okRemaining := new(big.Int).SetString(result.Remaining, 10)
				if !okFilled || !okRemaining || new(big.Int).Add(filled, remaining).Cmp(amt(1)) != 0 {
					problems <- fmt.Errorf("response mixes operations: filled=%s remaining=%s, want sum=%s", result.Filled, result.Remaining, amt(1))
				}
			}
		}()
	}
	close(start)
	wg.Wait()
	close(problems)
	for err := range problems {
		t.Error(err)
	}
}

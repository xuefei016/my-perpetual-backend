package keeper

import (
	"context"
	"errors"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/xuefei016/my-perpetual-backend/internal/chain/bindings/dealer"
	"github.com/xuefei016/my-perpetual-backend/internal/chaintest"
)

func TestFundingTickSkipsWithoutTransaction(t *testing.T) {
	for _, prices := range [][2]*big.Int{
		{nil, nil}, {usd(26000), nil}, {nil, usd(26200)}, {usd(26000), usd(26000)},
	} {
		env := newFundingEnv(t, time.Second, testBTC)
		env.book.best[testBTC] = prices
		// auth 和发送交易的 backend 均不可用：空更新必须在触碰它们之前返回。
		if err := env.k.tick(context.Background()); err != nil {
			t.Fatalf("prices %v: %v", prices, err)
		}
	}
}

func TestFundingTickWrapsReadError(t *testing.T) {
	env := newFundingEnv(t, time.Second, testBTC)
	env.book.err = chaintest.ErrNodeDown
	err := env.k.tick(context.Background())
	if !errors.Is(err, chaintest.ErrNodeDown) || !strings.Contains(err.Error(), "nextRates") {
		t.Fatalf("want nextRates wrapping node error, got %v", err)
	}
}

// 只拦截真实绑定发出的 eth_call，以观察 Run 传到链上的 context。
// 每轮读链失败，不进入假链尚未支持的发交易路径。
type fundingCallBackend struct {
	bind.ContractBackend
	call func(context.Context) ([]byte, error)
}

func (b *fundingCallBackend) CallContract(ctx context.Context, _ ethereum.CallMsg, _ *big.Int) ([]byte, error) {
	return b.call(ctx)
}

func interceptFundingCalls(t *testing.T, env *fundingEnv, call func(context.Context) ([]byte, error)) {
	t.Helper()
	b := &fundingCallBackend{ContractBackend: env.fb, call: call}
	d, err := dealer.NewDealer(testDealer, b)
	if err != nil {
		t.Fatal(err)
	}
	env.k.c.Backend, env.k.c.Dealer = b, d
	env.book.best[testBTC] = [2]*big.Int{usd(26000), usd(26200)}
}

func startFundingRun(t *testing.T, k *FundingKeeper, ctx context.Context, cancel context.CancelFunc) <-chan error {
	t.Helper()
	done := make(chan error, 1)
	go func() {
		done <- k.Run(ctx)
		close(done)
	}()
	t.Cleanup(func() {
		cancel()
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Error("funding Run did not exit after cancellation")
		}
	})
	return done
}

func receiveFundingContext(t *testing.T, calls <-chan context.Context) context.Context {
	t.Helper()
	select {
	case ctx := <-calls:
		return ctx
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for funding tick")
		return nil
	}
}

func TestFundingRunWaitsRetriesAndReleasesContexts(t *testing.T) {
	const interval = 40 * time.Millisecond
	env := newFundingEnv(t, interval, testBTC)
	calls := make(chan context.Context, 32)
	interceptFundingCalls(t, env, func(ctx context.Context) ([]byte, error) {
		calls <- ctx
		return nil, chaintest.ErrNodeDown
	})
	type traceKey struct{}
	parent, cancel := context.WithCancel(context.WithValue(context.Background(), traceKey{}, "D9"))
	started := time.Now()
	startFundingRun(t, env.k, parent, cancel)
	first := receiveFundingContext(t, calls)
	if time.Since(started) < interval {
		t.Error("first tick ran before the first interval elapsed")
	}
	deadline, ok := first.Deadline()
	if !ok || deadline.Before(started.Add(29*time.Second)) || deadline.After(time.Now().Add(30*time.Second)) {
		t.Errorf("tick deadline = %v (present %v), want a 30s timeout", deadline, ok)
	}
	if first.Value(traceKey{}) != "D9" {
		t.Error("tick context lost parent values")
	}
	second := receiveFundingContext(t, calls) // 第一轮失败后仍然运行。
	if first == second {
		t.Error("ticks reused one context")
	}
	if !errors.Is(first.Err(), context.Canceled) {
		t.Errorf("completed tick context not released: %v", first.Err())
	}
	if parent.Err() != nil {
		t.Error("releasing a tick canceled its parent")
	}
	receiveFundingContext(t, calls)
}

func TestFundingRunCancellationWhileWaiting(t *testing.T) {
	env := newFundingEnv(t, time.Hour, testBTC)
	ctx, cancel := context.WithCancel(context.Background())
	done := startFundingRun(t, env.k, ctx, cancel)
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("normal shutdown returned %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("cancellation waited for the hourly ticker")
	}
}

func TestFundingRunCancellationDuringRPC(t *testing.T) {
	env := newFundingEnv(t, 10*time.Millisecond, testBTC)
	calls := make(chan context.Context, 32)
	interceptFundingCalls(t, env, func(ctx context.Context) ([]byte, error) {
		calls <- ctx
		<-ctx.Done()
		return nil, ctx.Err()
	})
	ctx, cancel := context.WithCancel(context.Background())
	done := startFundingRun(t, env.k, ctx, cancel)
	tickCtx := receiveFundingContext(t, calls)
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("normal shutdown returned %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("parent cancellation did not interrupt in-flight RPC")
	}
	if !errors.Is(tickCtx.Err(), context.Canceled) {
		t.Errorf("in-flight context = %v, want canceled", tickCtx.Err())
	}
}

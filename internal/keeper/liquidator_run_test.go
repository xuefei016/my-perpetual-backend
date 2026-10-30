package keeper

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/xuefei016/my-perpetual-backend/internal/chain/bindings/dealer"
	"github.com/xuefei016/my-perpetual-backend/internal/chaintest"
)

func interceptLiquidatorQueries(t *testing.T, l *Liquidator, fb *chaintest.Backend, call func(context.Context) ([]byte, error)) {
	t.Helper()
	b := &fundingCallBackend{ContractBackend: fb, call: call}
	d, err := dealer.NewDealer(testDealer, b)
	if err != nil {
		t.Fatal(err)
	}
	l.c.Dealer = d
	l.traders = traderList{liqTrader}
}

func startLiquidatorRun(t *testing.T, l *Liquidator, ctx context.Context, cancel context.CancelFunc) <-chan error {
	t.Helper()
	done := make(chan error, 1)
	go func() {
		done <- l.Run(ctx)
		close(done)
	}()
	t.Cleanup(func() {
		cancel()
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Error("liquidator Run did not exit after cancellation")
		}
	})
	return done
}

func TestLiquidatorRunWaitsRetriesAndReleasesQueries(t *testing.T) {
	l, fb, _ := newLiquidationEnv(t)
	l.interval = 40 * time.Millisecond
	calls := make(chan context.Context, 32)
	interceptLiquidatorQueries(t, l, fb, func(ctx context.Context) ([]byte, error) {
		calls <- ctx
		return nil, chaintest.ErrNodeDown
	})
	ctx, cancel := context.WithCancel(context.WithValue(context.Background(), liquidationTraceKey{}, t.Name()))
	started := time.Now()
	startLiquidatorRun(t, l, ctx, cancel)
	first := receiveFundingContext(t, calls)
	if time.Since(started) < l.interval {
		t.Error("first scan did not wait for interval")
	}
	assertLiquidationChildContext(t, ctx, first, 10*time.Second)
	second := receiveFundingContext(t, calls)
	if first == second || !errors.Is(first.Err(), context.Canceled) {
		t.Errorf("query contexts reused or not released: same=%v err=%v", first == second, first.Err())
	}
	receiveFundingContext(t, calls)
	if ctx.Err() != nil {
		t.Error("releasing query context canceled Run")
	}
}

func TestLiquidatorRunCancellationWhileWaiting(t *testing.T) {
	l, _, _ := newLiquidationEnv(t)
	l.interval = time.Hour
	ctx, cancel := context.WithCancel(context.Background())
	done := startLiquidatorRun(t, l, ctx, cancel)
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("normal shutdown returned %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("shutdown waited for ticker")
	}
}

func TestLiquidatorRunCancellationDuringQuery(t *testing.T) {
	l, fb, _ := newLiquidationEnv(t)
	l.interval = 10 * time.Millisecond
	calls := make(chan context.Context, 32)
	interceptLiquidatorQueries(t, l, fb, func(ctx context.Context) ([]byte, error) {
		calls <- ctx
		<-ctx.Done()
		return nil, ctx.Err()
	})
	ctx, cancel := context.WithCancel(context.Background())
	done := startLiquidatorRun(t, l, ctx, cancel)
	queryCtx := receiveFundingContext(t, calls)
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("normal shutdown returned %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("parent cancellation did not interrupt query")
	}
	if !errors.Is(queryCtx.Err(), context.Canceled) {
		t.Errorf("query context = %v, want canceled", queryCtx.Err())
	}
}

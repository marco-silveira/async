package async_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/StudioSol/async"
)

func TestRun(t *testing.T) {
	t.Run("Two AsyncFunc success", func(t *testing.T) {
		var exec [2]bool
		f1 := func(_ context.Context) error {
			exec[0] = true
			return nil
		}

		f2 := func(_ context.Context) error {
			exec[1] = true
			return nil
		}

		err := async.Run(context.Background(), f1, f2)
		require.Nil(t, err)
		require.True(t, exec[0])
		require.True(t, exec[1])
	})

	t.Run("Two AsyncFunc, one fails", func(t *testing.T) {
		var errTest = errors.New("test error")
		f1 := func(_ context.Context) error {
			return errTest
		}

		f2 := func(_ context.Context) error {
			return nil
		}

		err := async.Run(context.Background(), f1, f2)
		require.True(t, errors.Is(errTest, err))
	})

	t.Run("Two AsyncFunc, one panics", func(t *testing.T) {
		f1 := func(_ context.Context) error {
			panic(errors.New("test panic"))
		}

		f2 := func(_ context.Context) error {
			return nil
		}

		err := async.Run(context.Background(), f1, f2)
		require.Error(t, err)
		require.Contains(t, err.Error(), "async.Run: panic test panic")
	})

	t.Run("Two AsyncFunc and one panics, the other doesn't execute", func(t *testing.T) {
		var mu sync.Mutex
		var exec [2]bool

		f1 := func(_ context.Context) error {
			exec[0] = true
			panic(errors.New("test panic"))
		}

		f2 := func(_ context.Context) error {
			time.Sleep(5 * time.Millisecond)
			mu.Lock()
			exec[1] = true
			mu.Unlock()
			return nil
		}

		_ = async.Run(context.Background(), f1, f2)
		mu.Lock()
		require.False(t, exec[1])
		mu.Unlock()
	})

	t.Run("If panics, cancel context", func(t *testing.T) {
		var copyCtx context.Context
		f1 := func(ctx context.Context) error {
			copyCtx = ctx
			panic(errors.New("test panic"))
		}

		err := async.Run(context.Background(), f1)
		require.Error(t, err)
		<-copyCtx.Done()
		require.Error(t, copyCtx.Err())
	})

	t.Run("cancel children when cancellable context is canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.TODO())
		cancel()

		var childCtx context.Context
		f1 := func(ctx context.Context) error {
			childCtx = ctx
			return nil
		}

		err := async.Run(ctx, f1)
		require.Nil(t, err)
		<-childCtx.Done()
		require.Error(t, childCtx.Err())
	})
}

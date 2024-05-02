package parallel

import (
	"context"
	"golang.org/x/sync/errgroup"
)

type mapFunc[T any, R any] func(context.Context, T) (R, error)

func Map[T any, R any](ctx context.Context, in []T, f mapFunc[T, R], numWorkers int) ([]R, error) {
	ch := make(chan R)
	defer close(ch)

	group, ctx := errgroup.WithContext(ctx)
	group.SetLimit(numWorkers)
	out := make([]R, len(in))

	// Go over every input to spawn workers
	for idx, item := range in {
		idx, item := idx, item // save item in context

		// Spawn a goroutine via error group
		group.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				r, err := f(ctx, item)
				if err != nil {
					return err
				}

				out[idx] = r

				return nil
			}
		})
	}

	if err := group.Wait(); err != nil {
		return nil, err
	}

	return out, nil
}

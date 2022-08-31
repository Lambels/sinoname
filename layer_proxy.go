package sinoname

import (
	"context"
	"errors"
	"sync"

	"golang.org/x/sync/errgroup"
)

// ErrQuit can be returned from a proxy function to indicate that the whole generation process
// should be stopped.
//
// No message will be recieved if this error is returned.
var ErrQuit error = errors.New("sinoname: abort process")

// ProxyFunc takes in a message and returns an error based on the input.
//
// If the error is not nil: the input is skipped.
//
// If the error is ErrQuit: the whole generation process gets closed.
//
// If the error is nil: the message is passed to further proxy functions.
type ProxyFunc func(string) error

// ProxyLayer holds all the proxy funcs it has, for a message to pass a proxy layer it must run
// through all the proxys without any error.
type ProxyLayer struct {
	proxys []ProxyFunc
}

// PumpOut recieves messages from the upstream layer via the in channel and passes them through the transformers.
// The end products of the transformers are fed in the returned channel.
func (l *ProxyLayer) PumpOut(ctx context.Context, g *errgroup.Group, in <-chan string) (<-chan string, error) {
	outC := make(chan string)

	if len(l.proxys) == 0 {
		return nil, errors.New("sinoname: layer has no proxys")
	}

	var wg sync.WaitGroup
	pumpOut := func(ctx context.Context, in string) func() error {
		f := func() error {
			defer wg.Done()

			for _, p := range l.proxys {
				if err := p(in); err != nil {
					// if err is ErrQuit abort process.
					if err == ErrQuit {
						return err
					}

					// just return
					return nil
				}
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			case outC <- in:
				return nil
			}
		}

		return f
	}

	go func() {
		defer close(outC)

		for {
			select {
			case <-ctx.Done():

			case v, ok := <-in:
				if !ok {
					return
				}

				g.Go(pumpOut(ctx, v))
			}
		}
	}()

	return outC, nil
}

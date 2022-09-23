package layer

import (
	"context"
	"errors"
	"sync"

	"github.com/Lambels/sinoname/config"
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

// ProxyFactory takes in a config object and returns a proxy function.
type ProxyFactory func(cfg *config.Config) ProxyFunc

// ProxyLayer holds all the proxy funcs it has, for a message to pass a proxy layer it must run
// through all the proxys without any error.
type ProxyLayer struct {
	Proxys []ProxyFunc
}

// PumpOut recieves messages from the upstream layer via the in channel and passes them through the transformers.
// The end products of the transformers are fed in the returned channel.
func (l *ProxyLayer) PumpOut(ctx context.Context, g *errgroup.Group, in <-chan string) (<-chan string, error) {
	if len(l.Proxys) == 0 {
		return nil, errors.New("sinoname: layer has no proxys")
	}

	outC := make(chan string)
	var wg sync.WaitGroup
	pumpOut := func(in string) func() error {
		f := func() error {
			defer wg.Done()

			for _, p := range l.Proxys {
				if err := p(in); err != nil {
					// if err is ErrQuit abort process.
					if err == ErrQuit {
						return err
					}

					// just return.
					return nil
				}
			}

			select {
			case <-ctx.Done():
				return nil
			case outC <- in:
				return nil
			}
		}

		return f
	}

	go func() {
		defer func() {
			if ctx.Err() != nil {
				return
			}

			wg.Wait()
			close(outC)
		}()

		for {
			select {
			case <-ctx.Done():

			case v, ok := <-in:
				if !ok {
					return
				}

				g.Go(pumpOut(v))
			}
		}
	}()

	return outC, nil
}

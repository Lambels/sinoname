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

// ProxyFactory takes in a config object and returns a proxy function and a state indicator.
//
// If the state indicator has true boolean value then the proxy layer using it is
// going to create a new ProxyFunc per each (sinoname.Layer).PumpOut() call.
//
// The state values is usefull for proxy functions since you might want to keep state across
// a pipeline with proxy functions.
type ProxyFactory func(cfg *Config) (ProxyFunc, bool)

// ProxyLayer holds all the proxy funcs it has (statefull or not), for a message to pass a proxy layer it must run
// through all the proxys without any error.
type ProxyLayer struct {
	cfg            *Config
	proxys         []ProxyFunc
	proxyFactories []ProxyFactory
}

// PumpOut recieves messages from the upstream layer via the in channel and passes them through the transformers.
// The end products of the transformers are fed in the returned channel.
func (l *ProxyLayer) PumpOut(ctx context.Context, g *errgroup.Group, in <-chan string) (<-chan string, error) {
	if len(l.proxys) == 0 && len(l.proxyFactories) == 0 {
		return nil, errors.New("sinoname: layer has no proxys")
	}

	statefullProxys := make([]ProxyFunc, len(l.proxyFactories))
	for i, f := range l.proxyFactories {
		p, _ := f(l.cfg)
		statefullProxys[i] = p
	}

	outC := make(chan string)
	// wg is used to monitor the local go routines of this layer.
	var wg sync.WaitGroup
	pumpOut := func(in string) func() error {
		f := func() error {
			defer wg.Done()

			// normal proxy run.
			for _, p := range l.proxys {
				if err := p(in); err != nil {
					// if err is ErrQuit abort process.
					if err == ErrQuit {
						return err
					}

					// just return.
					return nil
				}
			}

			// statefull proxy run.
			for _, p := range statefullProxys {
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
		// before the factory go-routine exits, either by a context cancelation or by the
		// upstream's out channel closure, cleanup.
		defer func() {
			if ctx.Err() != nil {
				return
			}

			// wait for all the transformers to send their value before closing.
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

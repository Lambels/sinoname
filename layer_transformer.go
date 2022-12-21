package sinoname

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"

	"golang.org/x/sync/errgroup"
)

// TransformerLayer holds all the transformers belonging to it (statefull or not),
// when the layer runs it fans out all the messages it gets to all
// the transformers it owns (first to the unstatefull then to the statefull).
//
// teoretically 1 message to a layer with 4 transformers results in 4 messages (1 * 4).
type TransformerLayer struct {
	cfg                  *Config
	init                 int32
	transformers         []Transformer
	transformerFactories []TransformerFactory
}

func (l *TransformerLayer) PumpOut(ctx context.Context, g *errgroup.Group, in <-chan MessagePacket) (<-chan MessagePacket, error) {
	if len(l.transformers) == 0 && len(l.transformerFactories) == 0 {
		return nil, errors.New("sinoname: layer has no transformers")
	}

	// local copy of statefull trasnformers.
	transformers := l.getStatefullTransformers()
	transformers = append(transformers, l.transformers...)

	outC := make(chan MessagePacket)
	handleValue := func(ctx context.Context, wg *sync.WaitGroup, _ int, v MessagePacket) error {
		defer wg.Done()

		select {
		case outC <- v:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	handleSkip := func(ctx context.Context, wg *sync.WaitGroup, id int, v MessagePacket) error {
		// layer skip (let this message pass through).
		if id == -1 {
			select {
			case outC <- v:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		wg.Done()
		return nil
	}
	handleExit := func(wg *sync.WaitGroup, _ bool) {
		wg.Wait()
		close(outC)
	}
	broadcast := newPacketBroadcatser(
		ctx,
		l.cfg,
		in,
		g,
		transformers,
		handleValue,
		handleSkip,
		handleExit,
	)
	broadcast.StartListen()

	return outC, nil
}

func (l *TransformerLayer) getStatefullTransformers() []Transformer {
	if len(l.transformerFactories) == 0 {
		return nil
	}

	statefullTransformers := make([]Transformer, len(l.transformerFactories))
	// get initiall values if the first caller.
	if atomic.CompareAndSwapInt32(&l.init, 0, 1) {
		diff := len(l.transformers) - len(l.transformerFactories)
		copy(statefullTransformers, l.transformers[diff:])
		l.transformers = l.transformers[:diff]
		return statefullTransformers
	}

	for i, f := range l.transformerFactories {
		t, _ := f(l.cfg)
		statefullTransformers[i] = t
	}
	return statefullTransformers
}

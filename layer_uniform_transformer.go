package sinoname

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"

	"golang.org/x/sync/errgroup"
)

// UniformTransformerLayer syncronises writes to the the downsream layer whilst not blocking
// the upstream layer.
//
// UniformTransformerLayer should be used when having transformers with very different
// transforming speeds. Instead of flooding the pipeline with lower speed messages from
// faster transformers the UniformTransformerLayer waits for the slower and faster transformers
// to write at the same time.
//
// Even though the layer waits for the messages to be written at the same time it doesent
// "sleep". Each transformer has a buffer so that when new messages come and a previous
// message is synced, the layer doesent wait for the previous message to be synced, it takes
// the new message, processes it and then writes it to the buffer. When the previous message
// is synced the layer pulls messages from each transformers buffer and syncs them.
type UniformTransformerLayer struct {
	cfg                  *Config
	init                 int32
	transformers         []Transformer
	transformerFactories []TransformerFactory
}

func (l *UniformTransformerLayer) PumpOut(ctx context.Context, g *errgroup.Group, in <-chan MessagePacket) (<-chan MessagePacket, error) {
	if len(l.transformers) == 0 && len(l.transformerFactories) == 0 {
		return nil, errors.New("sinoname: layer has no transformers")
	}

	// local copy of statefull trasnformers.
	transformers := l.getStatefullTransformers()
	transformers = append(transformers, l.transformers...)

	outC := make(chan MessagePacket)
	out := newSyncOut(len(l.transformers)+len(l.transformerFactories), outC)
	handleValue := func(_ context.Context, wg *sync.WaitGroup, id int, v MessagePacket) error {
		defer wg.Done()
		out.Write(id, v)
		return nil
	}
	handleSkip := func(ctx context.Context, wg *sync.WaitGroup, id int, v MessagePacket) error {
		if id == -1 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case outC <- v:
				return nil
			}
		}

		defer wg.Done()
		out.Advance(id)
		return nil
	}
	handleExit := func(wg *sync.WaitGroup, forced bool) {
		if forced {
			out.Close()
			return
		}

		wg.Wait()
		out.Close()
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

func (l *UniformTransformerLayer) getStatefullTransformers() []Transformer {
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

# Sinoname [![Go Report Card](https://goreportcard.com/badge/github.com/Lambels/sinoname)](https://goreportcard.com/report/github.com/Lambels/sinoname)

Sinoname is a highly customizable, flexible, extensible, out of the box ready, pipeline for generating similarly looking strings.

### Purpose:
Sinoname's aim is to provide a flexible and straightforward api to generate similar looking usernames when the original username isn't available.

### go get it:
```bash
go get github.com/Lambels/sinoname@latest
```

# Code Examples
## Example (2 Layers):
Sinoname configuration with 2 layers.

Layer1:
- [Camel case transformer](https://github.com/Lambels/sinoname/blob/main/transformer_camel_case.go)
- [Snake case transformer](https://github.com/Lambels/sinoname/blob/main/transformer_snake_case.go)
- [Kebab case transformer](https://github.com/Lambels/sinoname/blob/main/transformer_kebab_case.go)

Layer2:
- [Numbers prefix transformer](https://github.com/Lambels/sinoname/blob/main/transformer_numbers_prefix.go)
- [Plural transformer](https://github.com/Lambels/sinoname/blob/main/transformer_plural.go)
- [No op transformer](https://github.com/Lambels/sinoname/blob/main/transformer_noop.go)

### Code:
```go
gen := sinoname.New(someConfig)

gen.WithTransformers(
    sinoname.CamelCase,
    sinoname.SnakeCase,
    sinoname.KebabCase,
) // first layer.

gen.WithTransformers(
    sinoname.NumbersPrefix(""),
    sinoname.Plural,
    sinoname.Noop,
) // second layer.

vals, _ := gen.Generate(context.Background(), "Lam.bel2006") // error ignored for example.
fmt.Println(vals)

// Output:
// [2006LamBel 2006Lam_bel 2006Lam-bel LamBel2006s Lam_bel2006s Lam-bel2006s LamBel2006 Lam_bel2006 Lam-bel2006]
```

### Diagram:
<details><summary>Show</summary>

<img src=".github/2%20layers.svg" alt="diagram_2layers" width="500"/>

</details>

## Example (Uniform Transformer):
This example will show the use case of the [uniform transformer layer](https://github.com/Lambels/sinoname#uniformtransformerlayer).

In a diverse pipeline with 3 layers there might be points in your pipeline where a bit more work will be done by some transformers. We will simulate this work with the `timeoutTransformer`:

```go
type timeoutTransformer struct {
	add string
	d   time.Duration
}

func (t timeoutTransformer) Transform(ctx context.Context, val string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case <-time.After(t.d):
		return val + t.add, nil
	}
}
```

The `timeoutTransformer` simulates a possible IO operation which might take some time, by sleeping for `d` seconds and the returning the initial message with the `add` value concatenated to it.

Now lets look at 2 possible sinoname configurations:

### Code:
The code in both examples is the same, but for a subtle change. Where the second example uses `gen.WithUniformTransformers()` the first uses `gen.WithTransformers()`. Yet the change is small, the output is very different.

Example 1 (No Uniform Transformer Layer):
```go
gen := sinoname.New(&sinoname.Config{
		MaxVals: 10, // set limit to 10 values.
})

gen.WithTransformers(sinoname.Noop, sinoname.Noop, newTimeoutTransformer("bar", 5*time.Second)) // Layer 1 (Normal)
gen.WithTransformers(sinoname.Noop, sinoname.Noop, sinoname.Noop) // Layer 2
gen.WithTransformers(sinoname.Noop, sinoname.Noop, sinoname.Noop) // Layer 3

vals, _ := gen.Generate(context.Background(), "foo")
fmt.Println(vals)
// Output: (reproducible)
// [foo foo foo foo foo foo foo foo foo foo]
// Time Taken: < 5 seconds
```

Example 2 (With Uniform Transformer Layer):
```go
gen := sinoname.New(&sinoname.Config{
		MaxVals: 10, // set limit to 10 values.
})

gen.WithUniformTransformers(sinoname.Noop, sinoname.Noop, newTimeoutTransformer("bar", 5*time.Second)) // Layer 1 (Uniform Layer)
gen.WithTransformers(sinoname.Noop, sinoname.Noop, sinoname.Noop) // Layer 2
gen.WithTransformers(sinoname.Noop, sinoname.Noop, sinoname.Noop) // Layer 3

vals, _ := gen.Generate(context.Background(), "foo")
fmt.Println(vals)
// Output: (varies from run to run)
// [foobar foobar foobar foo foo foo foo foo foo foo]
// Time Taken: 5 seconds
```

By comparing the outputs, in the first example you will never get the values from the `timeoutTransformer` because the pipeline gets flooded with the much faster produced values from `NoopTransformer`.

When having very differently time paced transformers sticked in the same layer and only interested in the first N element (no point if you are going to wait for all the messages), consider using the `UniformLayerTransformer`
### Diagram:

Example 1 (No Uniform Transformer Layer):

Example 2 (With Uniform Transformer Layer):

# Docs
## Transformers:
Transformers as described by the interface:
```go
type Transformer interface {
	Transform(ctx context.Context, in string) (string, error)
}
```
Take in a string and return a modified string and an error, transformers are the building blocks of this pipeline. Above modifying the string thats coming in, the transformers need to validate the generated modified value against the config object they receive via:
```go
type TransformerFactory func(cfg *Config) (Transformer, bool)
```
`TransformerFactory` is used to create transformers since no layer accepts the raw `Transformer` type, but only the `TransformerFactory` type.

The transformers must check the modified values against these `cfg` fields:
- `cfg.Source` - check if the modified value is unique (see more details [here](https://github.com/Lambels/sinoname#Source))
- `cfg.MaxLen` - check that the modified value isn't over `cfg.MaxLen`
If the modified value doesn't meet one of these criteria the original value must be returned.

The context taken in provides contextual information about the current transaction via these [methods](https://github.com/Lambels/sinoname/blob/main/context.go). Also the context provides cancellation notifications to the transformers to return early and not stall the pipeline, these notifications should always (when possible and necessary) be respected. When the context is cancelled the error returned by the transformer must be non `nil`.

### Errors:
Transformers can return an errors, there are 3 scenarios possible:
1. The error is `nil`: The message gets sent further down the pipeline
2. The error is `ErrSkip`: The message gets skipped without shutting down the pipeline.
3. The error isn't `nil`: The whole pipeline gets shutdown and no messages are received by the sink (context gets cancelled).

### Stateful Transformers:
Stateful transformers are transformers which get re-initialized on each `(*sinoname.Generator).Generate()` call. They are so called "Stateful Transformers" because they are stateful in respect to each message sent through the pipeline. "Non Stateful Transformers" get re used for all `(*sinoname.Generator).Generate()` calls, therefor they have no state in respect to a particular message.

This is how to indicate whether a transformer is "stateful" or not:
```go
type TransformerFactory func(cfg *Config) (Transformer, bool)
```
The returned `bool` value is used, if its truth is `true` the `sinoname.Transformer` is marked as "stateful" and a new copy of the `sinoname.Transformer` will be made on each `(*sinoname.Generator).Generate()` call. Else the `sinoname.Transformer` behaves as normal and the transformer value gets re-used on each new message through the pipeline.

## Layers:
Layers are simply put, containers which hold [transformer functions](https://github.com/Lambels/sinoname#Transformers). Their job is to take in messages -> distribute them to the transformer functions -> collect the modified values from the transformers -> pass them to the next layer.

In the same time layers have to listen for context cancellations and handle graceful closing.

```go
type Layer interface {
	PumpOut(ctx context.Context, g *errgroup.Group, inBoundC <-chan string) (<-chan string, error)
}
```

A layer must exit when the inbound channel (`inBoundC`) gets closed or when the context gets cancelled. If the inbound channel gets closed, all the messages in the layer must be processed and sent before the layer exits. If the context gets cancelled the layer must exit as soon as possible whilst cleaning up all transformer go-routines.

Once the layer exits it must close its outbound channel to signal to the next layer in the pipeline to close, eventually this closing signal reaches the sink.

Sinoname has 2 `sinoname.Layer` implementations:
- `sinoname.TransformerLayer`
- `sinoname.UniformTransformerLayer`

### TransformerLayer:
This is the layer you will use most of the time since its the most simple and will cover most use-cases you will have.

To pair your transformers in this layer you need to use:
```go
gen := sinoname.New(someConfig)

gen.WithTransformers(tr1, tr2, tr3, tr4) // Transformer Layer with transformers: tr1, tr2, tr3 and tr4
```

The functionality of this layer is simple and can be described in 3 steps:
1. The layer receives a message and distributes it to all its transformers.
2. It runs all transformers in separate go routines via `(*errgroup.Group).Go()`
3. It collects all the values from the transformer go-routines and sends each one to the next layer.

### UniformTransformerLayer:
This layer is very similar to the former but buffers the return values in such a way that there is no faster or slower paced values. This layer must only be used when pairing slower with faster transformers in one layer and having a set `(sinoname.Config).MaxLen`.

Using this layer prevents the faster working transformers from flooding the pipeline with their faster paced messages and hitting the upper bound of `(sinoname.Config).MaxLen` without giving the slower paced messages a chance to reach the consumer.

To pair your transformers in this layer you need to use:
```go
gen := sinoname.New(someConfig)

gen.WithUniformTransformers(tr1, tr2, tr3, tr4) // Unfirom Transformer Layer with transformers: tr1, tr2, tr3 and tr4
```

## Config:
The [config struct](https://github.com/Lambels/sinoname/blob/main/config.go) is used to alter the behavior of `sinoname.Generator`.

It is passed to `sinoname.New()`, `sinoname.LayerFactory` and `sinoname.TransformerFactory` to create respectively `sinoname.Generator`, `sinoname.Layer` and `sinoname.Transformer`.

The base layer of sinoname is `sinoname.Generator` which holds all the layers (`sinoname.Layer`) of the pipeline which hold all the transformers (`sinoname.Transformer`).

### Config Fields:
| Field       | Type | Description |
| ----------- | --------- | ----------- |
| MaxLen      | `int` | MaxLen is used to set the max length of the returned values, it must be enforced by the transformers to make sure that they don't return values longer then MaxLen.        |
| MaxVals   | `int` | MaxVals is used to set the max number of returned values. The consumer reads up to MaxVals values.       |
| PreventDefault | `bool` | PreventDefault prevents the default value from being read by the consumer. |
| Source | `sinoname.Source` | Source is used to validate if the products of the transformers are unique / valid. |
| SplitOn | `[]string` | SplitOn is a slice of symbols used by the case transformers (camel case, kebab case, ...) to decide where to split the word up and add their specific separator. |

## Source:
`sinoname.Source` is an interface which must be implemented by the client. It is used by [transformers](https://github.com/Lambels/sinoname#Transformers) to validate if their return value is unique.

```go
type Source interface {
	Valid(context.Context, string) (bool, error)
}
```

The return values indicate if the message is valid ("unique"), if the `bool` truth value is `true` then the value is unique, if its `false` its not and the transformer must return the initial value. If the error isn't `nil` then the transformer must return the error and shutdown the pipeline.

## Extending Sinoname:
Sinoname is very extensible by providing and accepting a variety of interfaces and being very simple.

### Contributing:
Do you have a working `sinoname.Transformer` or `sinoname.Layer` which you could see being broadly used? If so:
1. Fork it
2. Create your transformer branch (`git checkout -b some-transformer`)
3. Test your feature
4. Commit your changes (`git commit -m 'Some Cool Layer'`)
5. Push your code to the branch (`git push origin some-new-transformer`)
6. Create a new pull request

### Custom Transformers:
Creating a custom transformer is as simple as:

`MyCustomTransformer` implements `sinoname.Transformer`
```go
type MyCustomTransformer struct {
    addSuffix string
    source sinoname.Source
    maxLen int
}

// add the suffix t.addSuffix to the incoming message.
func (t MyCustomTransformer) Transform(ctx context.Context, in string) (string, error) {
    // new message would be too long, return the initial message with nil error.
    if len(in) + len(t.addSuffix) > t.maxLen {
        return in, nil
    }

    out := in + t.addSuffix

    ok, err := t.source.Valid(ctx, out)
    // source error, return it.
    if err != nil {
        return "", err
    }
    // invalid out value, return initial message.
    if !ok {
        return in, nil
    }

    return out, nil
}
```

`NewCustomTransformer` returns a closure which of signature `sinoname.TransformerFactory`
```go
func NewCustomTransformer(addSuffix string) sinoname.TransformerFactory {
    return func(cfg *sinoname.Config) (sinoname.Transformer, bool) {
        cstmTr := MyCustomTransformer{
            addSuffix: addSuffix,
            source: cfg.Source,
            maxLen: cfg.MaxLen,
        }

        return cstmTr, false
    }   
}
```

Using the custom transformer
```go
gen := sinoname.New(someConfig)

gen.WithTransformers(
    NewCustomTransformer("_some_suffix"),
    sinoname.CamelCase,
    sinoname.Noop,
)

vals, _ := gen.Generate(context.Background(), "lam.bels")
fmt.Println(vals)
// Output:
// [lam.bels_some_suffix, lam_bels, lam.bels]
```

### Custom Layers:
Layers are similarly easy to implement (check [transformer layer](https://github.com/Lambels/sinoname/blob/main/layer_transformer.go) for simple implementation).

Lets build a logging layer:

`LoggingLayer` implement `sinoname.Layer`
```go
type LoggingLayer struct {
    logger *log.Logger
}

// we ignore the error group since our layer cant produce any errors.
func (l *LoggingLayer) PumpOut(ctx context.Context, _ *errgroup.Group, in <-chan string) (<- chan string, error) {
    if l.logger == nil {
        return nil, errors.New("expected a logger but got nil")
    }

    // channel on which we send messages out.
    outC := make(chan string)
    go func() {
        defer close(outC)

        for {
            select {
                case <-ctx.Done():
                    // ctx cancelled, exit
                    return

                case msg, ok := <-in:
                    // in channel cancelled, exit
                    if !ok {
                        return
                    }

                    logger.Printf("received message: %s", msg)

                    // pass the value to the next layer:
                    select {
                        case <-ctx.Done():
                            return
                        case outC <- msg:
                    }
            }
        }
    }()

    return outC, nil
}
```

`NewLoggingLayer` returns a closure which of signature `sinoname.LayerFactory`
```go
func NewLoggingLayer(logger *log.Logger) sinoname.LayerFactory {
    // we don't use the config in this layer.
    return func(_ *Config) sinoname.Layer {
        return &LoggingLayer{logger}
    }
}
```

Using the custom layer:
```go
gen := sinoname.New(someConfig)

gen.WithLayers(
    NewLoggingLayer(log.New(os.Stdout, "sinoname:", 0))
)

gen.Generate(context.Background(), "lam.bels")
// Output:
// 
```
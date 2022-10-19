# Sinoname [![Go Report Card](https://goreportcard.com/badge/github.com/Lambels/sinoname)](https://goreportcard.com/report/github.com/Lambels/sinoname)

Sinoname is a highly customizable, flexible, extensible, out of the box ready, pipeline for generating similarly looking strings.

### Purpose:
Sinoname's aim is to provide a flexible and straightforward api to generate similar looking usernames when the original username isn't available.

### go get it:
```bash
go get github.com/Lambels/sinoname@latest
```

# Docs
## Transformers:
Transformers as described by the interface:
```go
type Transformer interface {
	Transform(ctx context.Context, in string) (string, error)
}
```
Take in a string and return a modified string and an error, transformers are the building blocks of this pipeline.

The context taken in provides contextual information about the current transaction via these [methods](https://github.com/Lambels/sinoname/blob/main/context.go). Also the context provides cancellation notifications to the transformers to return early and not stall the pipeline, these notifications should always (when possible) be respected. When the context is cancelled the error returned by the transformer must be non `nil`.

### Errors
Transformers can return an errors, there are 3 scenarios possible:
1. The error is `nil`: The message gets sent further down the pipeline
2. The error isn't `nil`: The whole pipeline gets shutdown and no messages are received by the sink.
3. The error is `ErrSkip`: The message gets skipped without shutting down the pipeline.

## Layers:
Layers are simply put, containers which hold [transformer functions](https://github.com/Lambels/sinoname#Transformers). Their job is to take in messages -> distribute them to the transformer functions -> collect the modified values from the transformers -> pass them to the next layer.

In the same time layers have to listen for context cancellations and handle graceful closing.

```go
type Layer interface {
	PumpOut(context.Context, *errgroup.Group, <-chan string) (<-chan string, error)
}
```

A layer must exit when the inbound channel (`<-chan string`) gets closed or when the context gets cancelled. Once the layer exits it must close its outbound channel to signal to the next layer in the pipeline to close, eventually this closing signal reaches the sink.

Sinoname has 2 `Layer` implementations:
- `TransformerLayer`
- `UniformTransformerLayer`

### TransformerLayer:

### UniformTransformerLayer:

## Config:

## Source:

# Code Examples
## Example (2 Layers):

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
![diagram_2layers](.github/2%20layers.svg)
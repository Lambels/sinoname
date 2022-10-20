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

### Errors:
Transformers can return an errors, there are 3 scenarios possible:
1. The error is `nil`: The message gets sent further down the pipeline
2. The error isn't `nil`: The whole pipeline gets shutdown and no messages are received by the sink.
3. The error is `ErrSkip`: The message gets skipped without shutting down the pipeline.

### Stateful Transformers:

## Layers:
Layers are simply put, containers which hold [transformer functions](https://github.com/Lambels/sinoname#Transformers). Their job is to take in messages -> distribute them to the transformer functions -> collect the modified values from the transformers -> pass them to the next layer.

In the same time layers have to listen for context cancellations and handle graceful closing.

```go
type Layer interface {
	PumpOut(ctx context.Context, g *errgroup.Group, inBoundC <-chan string) (<-chan string, error)
}
```

A layer must exit when the inbound channel (`inBoundC`) gets closed or when the context gets cancelled. Once the layer exits it must close its outbound channel to signal to the next layer in the pipeline to close, eventually this closing signal reaches the sink.

Sinoname has 2 `Layer` implementations:
- `TransformerLayer`
- `UniformTransformerLayer`

### TransformerLayer:

### UniformTransformerLayer:

## Config:

## Source:

## Extending Sinoname:

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

package sinoname

import (
	"context"
	"math/rand"
	"testing"
	"time"
)

var gen = New(&Config{
	Source:     noopSource{false},
	Adjectives: make([]string, 50000),
	RandSrc:    rand.New(rand.NewSource(time.Now().UnixNano())),
}).WithTransformers(
	Suffix(""),
	Prefix(""),
	Circumfix(""),
).WithTransformers(
	Suffix(""),
	Prefix(""),
	Circumfix(""),
)

func benchmarkAdjectivesShuffle(b *testing.B, n int) {
	chunks = n
	for i := 0; i < b.N; i++ {
		_, err := gen.Generate(context.Background(), "")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAdjectivesShuffle1Chunk(b *testing.B)    { benchmarkAdjectivesShuffle(b, 1) }
func BenchmarkAdjectivesShuffle2Chunks(b *testing.B)   { benchmarkAdjectivesShuffle(b, 2) }
func BenchmarkAdjectivesShuffle4Chunks(b *testing.B)   { benchmarkAdjectivesShuffle(b, 4) }
func BenchmarkAdjectivesShuffle8Chunks(b *testing.B)   { benchmarkAdjectivesShuffle(b, 8) }
func BenchmarkAdjectivesShuffle16Chunks(b *testing.B)  { benchmarkAdjectivesShuffle(b, 16) }
func BenchmarkAdjectivesShuffle32Chunks(b *testing.B)  { benchmarkAdjectivesShuffle(b, 32) }
func BenchmarkAdjectivesShuffle64Chunks(b *testing.B)  { benchmarkAdjectivesShuffle(b, 64) }
func BenchmarkAdjectivesShuffle128Chunks(b *testing.B) { benchmarkAdjectivesShuffle(b, 128) }
func BenchmarkAdjectivesShuffle256Chunks(b *testing.B) { benchmarkAdjectivesShuffle(b, 256) }
func BenchmarkAdjectivesShuffle512Chunks(b *testing.B) { benchmarkAdjectivesShuffle(b, 512) }

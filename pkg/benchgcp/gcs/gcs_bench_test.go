package benchgcs_test

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"cloud.google.com/go/storage"
	"github.com/vsekhar/fabula/internal/bigarray"
	"google.golang.org/api/googleapi"
)

const defaultBucketName = "fabula-nam"

var bucketName = flag.String("bucket", defaultBucketName, "bucket to test against")

const objPrefix = "__fabula_benchgcs__"

func setup(tb testing.TB, ctx context.Context) *storage.BucketHandle {
	client, err := storage.NewClient(ctx)
	if err != nil {
		tb.Fatal(err)
	}
	return client.Bucket(*bucketName)
}

var metadata = map[string]string{
	"abc": "123",
}

func writeObject(ctx context.Context, bkt *storage.BucketHandle, i int, metadata map[string]string) error {
	name := objPrefix + fmt.Sprintf("%d", i)
	obj := bkt.Object(name)
	w := obj.If(storage.Conditions{DoesNotExist: true}).NewWriter(ctx)
	w.ObjectAttrs.Metadata = metadata
	if err := w.Close(); err != nil {
		return err
	}
	return nil
}

func cleanup(tb testing.TB, ctx context.Context, bkt *storage.BucketHandle, n int) {
	wg := sync.WaitGroup{}
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) error {
			name := objPrefix + fmt.Sprintf("%d", i)
			obj := bkt.Object(name)
			err := obj.Delete(ctx)
			if err != nil {
				tb.Errorf("deleting %s: %v", name, err)
			}
			wg.Done()
			return nil
		}(i)
	}
	wg.Wait()
}

// BenchmarkGCSParallel benchmarks GCS read and write performance.
func BenchmarkGCSWriteParallel(b *testing.B) {
	ctx := context.Background()
	bkt := setup(b, ctx)
	b.ResetTimer()

	// TODO: user b.RunParallel

	wg := sync.WaitGroup{}
	wg.Add(b.N)
	for i := 0; i < b.N; i++ {
		go func(i int) {
			if err := writeObject(ctx, bkt, i, metadata); err != nil {
				b.Error(err)
			}
		}(i)
	}
	wg.Wait()
	b.StopTimer()
	cleanup(b, ctx, bkt, b.N)
}

func BenchmarkGCSWrite(b *testing.B) {
	ctx := context.Background()
	bkt := setup(b, ctx)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := writeObject(ctx, bkt, i, metadata); err != nil {
			b.Error(err)
		}
	}
	b.StopTimer()
	cleanup(b, ctx, bkt, b.N)
}

func BenchmarkAppend(b *testing.B) {
	ctx := context.Background()
	bkt := setup(b, ctx)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Assuming ideal information (we correctly guess the last value because
		// we know wrote it or because of really good caching), we still need
		// to at least check that that value doesn't exist (1 probe) to
		// determine that we guessed write.
		name := objPrefix + fmt.Sprintf("%d", i)
		_, err := bkt.Object(name).Attrs(ctx)
		if err == nil {
			b.Fatalf("expected object to not exist")
		}
		if err != storage.ErrObjectNotExist {
			b.Fatal(err)
		}

		if err := writeObject(ctx, bkt, i, metadata); err != nil {
			b.Error(err)
		}
	}
	b.StopTimer()
	cleanup(b, ctx, bkt, b.N)
}

func BenchmarkAppendParallel(b *testing.B) {
	ctx := context.Background()
	bkt := setup(b, ctx)
	var collisions uint64
	var probes uint64
	var latency uint64
	b.ResetTimer()
	b.SetParallelism(10) // I/O bound
	b.RunParallel(func(pb *testing.PB) {
		next := 0
		for pb.Next() {
			thisCollisions := 0
			thisProbes := 0
			start := time.Now()
			notExists := func(i int) bool {
				thisProbes++
				atomic.AddUint64(&probes, 1)
				name := objPrefix + fmt.Sprintf("%d", i)
				_, err := bkt.Object(name).Attrs(ctx)
				if err == nil {
					return false
				}
				if err == storage.ErrObjectNotExist {
					return true
				}
				b.Error(err)
				return true // stop search
			}
			tries := 50
			var t int
			for t = 0; t < tries; t++ {
				next = bigarray.Search(next, notExists)
				err := writeObject(ctx, bkt, next, metadata)
				if err == nil {
					break
				}
				switch ee := err.(type) {
				case *googleapi.Error:
					if ee.Code == http.StatusPreconditionFailed {
						next++
						thisCollisions++
						continue
					}
				}
				b.Error(err)
				break
			}
			if t == tries {
				b.Errorf("failed after %d tries: #%d", tries, next)
			}
			ms := time.Since(start).Milliseconds()
			atomic.AddUint64(&latency, uint64(ms))
			atomic.AddUint64(&collisions, uint64(thisCollisions))
			atomic.AddUint64(&probes, uint64(thisProbes))
			fmt.Printf("writing entry #%d took %d ms, %d probes, %d collisions\n", next, ms, thisProbes, thisCollisions)
		}
	})
	b.StopTimer()
	fN := float64(b.N)
	b.ReportMetric(float64(collisions)/fN, "collisions/op")
	b.ReportMetric(float64(probes)/fN, "probes/op")
	b.ReportMetric(float64(latency)/fN, "latency_ms/op")
	cleanup(b, ctx, bkt, b.N)
}

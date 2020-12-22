package atomicwriter

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
)

const filenamePrefix = "_atomicwriter_test_tmp_"

func tempFileSystemWriter(t *testing.T) (string, func()) {
	dir, err := ioutil.TempDir("", filenamePrefix)
	if err != nil {
		t.Fatal(err)
	}
	return dir, func() { os.Remove(dir) }
}

func randFilename() string {
	return fmt.Sprintf("%s%d", filenamePrefix, rand.Int63())
}

func TestFileSystemOverwrite(t *testing.T) {
	dir, cleanup := tempFileSystemWriter(t)
	defer cleanup()

	filename := randFilename()
	ctx := context.Background()

	// Create new
	d, err := NewDriver(ctx, dir)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := d.(*fsDriver); !ok {
		t.Fatalf("bad driver type, expected *fsDriver, got %T", d)
	}
	a1, err := d.NewAtomicWriter(ctx, filename)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a1.Write([]byte("abc")); err != nil {
		t.Fatal(err)
	}
	if exists, err := d.Exists(ctx, filename); err != nil || exists {
		t.Fatalf("expected false,nil; got %t, %s", exists, err)
	}
	if err := a1.CloseAtomically(); err != nil {
		t.Fatal(err)
	}
	if exists, err := d.Exists(ctx, filename); err != nil || !exists {
		t.Fatalf("expected true,nil; got %t, %s", exists, err)
	}

	// Create over existing, should fail
	a2, err := d.NewAtomicWriter(ctx, filename)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a2.Write([]byte("def")); err != nil {
		t.Fatal(err)
	}
	if err := a2.CloseAtomically(); !os.IsExist(err) {
		t.Errorf("expected os.IsExist, got %v", err)
	}
}

func TestFileSystemUnderwrite(t *testing.T) {
	dir, cleanup := tempFileSystemWriter(t)
	defer cleanup()

	filename := randFilename()
	ctx := context.Background()

	// Create new, write, but don't close
	d, err := NewDriver(ctx, dir)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := d.(*fsDriver); !ok {
		t.Fatalf("bad driver type, expected *fsDriver, got %T", d)
	}
	a1, err := d.NewAtomicWriter(ctx, filename)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a1.Write([]byte("abc")); err != nil {
		t.Fatal(err)
	}
	if exists, err := d.Exists(ctx, filename); err != nil || exists {
		t.Fatalf("expected false,nil; got %t, %s", exists, err)
	}

	// Create underneath existing writer, should succeed
	a2, err := d.NewAtomicWriter(ctx, filename)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a2.Write([]byte("def")); err != nil {
		t.Fatal(err)
	}
	if err := a2.CloseAtomically(); err != nil {
		t.Fatal(err)
	}
	if exists, err := d.Exists(ctx, filename); err != nil || !exists {
		t.Fatalf("expected true,nil; got %t, %s", exists, err)
	}

	// Attempt to close first writer, should fail with IsExist
	if err := a1.CloseAtomically(); !os.IsExist(err) {
		t.Fatal(err)
	}
}

const (
	// gcsBucket = ""
	gcsBucket = "gs://fabula-8589-public_storage"
)

func TestGCSOverwrite(t *testing.T) {
	if gcsBucket == "" {
		t.Skip("no bucket specified, skipping GCS tests")
		return
	}

	filename := randFilename()
	ctx := context.Background()

	d, err := NewDriver(ctx, gcsBucket)
	if err != nil {
		t.Fatal(err)
	}
	g, ok := d.(*gsDriver)
	if !ok {
		t.Fatalf("bad driver type, expected *gsDriver, got %T", d)
	}

	// Create new
	writerCtx, cancelWriter := context.WithCancel(ctx)
	defer cancelWriter()
	a1, err := d.NewAtomicWriter(writerCtx, filename)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a1.Write([]byte("abc")); err != nil {
		t.Fatal(err)
	}
	if exists, err := d.Exists(writerCtx, filename); err != nil || exists {
		t.Fatalf("expected false,nil; got %t, %s", exists, err)
	}
	if err := a1.CloseAtomically(); err != nil {
		t.Fatal(err)
	}
	if exists, err := d.Exists(writerCtx, filename); err != nil || !exists {
		t.Fatalf("expected true,nil; got %t, %s", exists, err)
	}
	defer func(ctx context.Context) {
		// Can't use a1.obj since it has a precondition DoesNotExist attached.
		err := g.bkt.Object(filepath.Join(g.prefix, filename)).Delete(ctx)
		if err != nil {
			t.Error(err)
		}
	}(writerCtx)

	// Create over existing, should fail
	writerCtx, cancelWriter = context.WithCancel(ctx)
	defer cancelWriter()
	a2, err := d.NewAtomicWriter(writerCtx, filename)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a2.Write([]byte("def")); err != nil {
		t.Fatal(err)
	}
	if err := a2.CloseAtomically(); !os.IsExist(err) {
		t.Errorf("expected os.IsExist, got %v", err)
	}
}

func TestGCSUnderwrite(t *testing.T) {
	if gcsBucket == "" {
		t.Skip("no bucket specified, skipping GCS tests")
		return
	}

	filename := randFilename()
	ctx := context.Background()

	d, err := NewDriver(ctx, gcsBucket)
	if err != nil {
		t.Fatal(err)
	}
	g, ok := d.(*gsDriver)
	if !ok {
		t.Fatalf("bad driver type, expected *gsDriver, got %T", d)
	}

	// Create new, write, but don't close
	writerCtx, cancelWriter := context.WithCancel(ctx)
	defer cancelWriter()
	a1, err := d.NewAtomicWriter(writerCtx, filename)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a1.Write([]byte("abc")); err != nil {
		t.Fatal(err)
	}

	// Create underneath existing writer, should succeed
	writerCtx, cancelWriter = context.WithCancel(ctx)
	defer cancelWriter()
	a2, err := d.NewAtomicWriter(writerCtx, filename)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a2.Write([]byte("def")); err != nil {
		t.Fatal(err)
	}
	if exists, err := d.Exists(writerCtx, filename); err != nil || exists {
		t.Fatalf("expected false,nil; got %t, %s", exists, err)
	}
	if err := a2.CloseAtomically(); err != nil {
		t.Fatal(err)
	}
	defer func(ctx context.Context) {
		// Can't use a1.obj since it has a precondition DoesNotExist attached.
		err := g.bkt.Object(filepath.Join(g.prefix, filename)).Delete(ctx)
		if err != nil {
			t.Error(err)
		}
	}(writerCtx)
	if exists, err := d.Exists(writerCtx, filename); err != nil || !exists {
		t.Fatalf("expected true,nil; got %t, %s", exists, err)
	}

	// Attempt to close first writer, should fail with IsExist
	if err := a1.CloseAtomically(); !os.IsExist(err) {
		t.Fatal(err)
	}
}

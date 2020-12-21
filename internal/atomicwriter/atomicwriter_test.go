package atomicwriter_test

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/vsekhar/fabula/internal/atomicwriter"
)

func TestAtomicWriterFileSystem(t *testing.T) {
	dir, err := ioutil.TempDir("", "atomicwritertesttmp-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(dir)

	const filename = "atomicwritertestfile"
	path := filepath.Join(dir, filename)
	ctx := context.Background()

	// Create new
	d := atomicwriter.NewFileSystemDriver()
	a, err := d.NewAtomicWriter(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a.Write([]byte("abc")); err != nil {
		t.Fatal(err)
	}
	if err := a.CloseAtomically(); err != nil {
		t.Fatal(err)
	}

	// Create over existing, should fail
	a, err = d.NewAtomicWriter(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a.Write([]byte("def")); err != nil {
		t.Fatal(err)
	}
	if err := a.CloseAtomically(); !os.IsExist(err) {
		t.Errorf("expected os.IsExist, got %v", err)
	}
}

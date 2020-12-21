package atomicwriter

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

// DriverInterface is the interface an atomic writer driver fulfills. It can be
// used to create atomic writers.
type DriverInterface interface {
	NewAtomicWriter(context.Context, string) (Interface, error)
}

// Interface is the interface an individual atomic writer fulfills.
type Interface interface {
	io.Writer
	CloseAtomically() error
}

const fsPattern = ".atomicwritertmp-*"

type fsDriverObject struct {
	path  string
	tfile *os.File
}

func (fdo *fsDriverObject) Write(b []byte) (int, error) {
	return fdo.tfile.Write(b)
}

func (fdo *fsDriverObject) CloseAtomically() error {
	if err := fdo.tfile.Sync(); err != nil {
		return err
	}
	// there's something on disk now, so ensure we cleanup. We know this is a
	// file so we don't need the extra protections os.Remove() provides.
	defer syscall.Unlink(fdo.tfile.Name())
	if err := fdo.tfile.Close(); err != nil {
		return err // shouldn't happen, we sync above
	}
	if err := os.Link(fdo.tfile.Name(), fdo.path); err != nil {
		return err
	}
	return nil
}

type fsDriver struct{}

func (f *fsDriver) NewAtomicWriter(_ context.Context, path string) (Interface, error) {
	d := filepath.Dir(path)
	tfile, err := ioutil.TempFile(d, fsPattern)
	if err != nil {
		return nil, err
	}
	return &fsDriverObject{
		path:  path,
		tfile: tfile,
	}, nil
}

// NewFileSystemDriver returns a new atomic writer backed by the local
// file system.
func NewFileSystemDriver() DriverInterface {
	return &fsDriver{}
}

type gsDriverObject struct {
	obj    *storage.ObjectHandle
	writer *storage.Writer
}

func (gsdo *gsDriverObject) Write(b []byte) (int, error) {
	return gsdo.writer.Write(b)
}

func (gsdo *gsDriverObject) CloseAtomically() error {
	// Just close it. Atomicity is assured with the storage condition defined
	// when creating the gsDriverObject in gsDriver.NewAtomicWriter().
	return gsdo.writer.Close()
}

type gsDriver struct {
	client *storage.Client
}

func (g *gsDriver) NewAtomicWriter(ctx context.Context, path string) (Interface, error) {
	bucket, name, err := parseGcsURI(path)
	if err != nil {
		return nil, err
	}
	// Important: DoesNotExist condition here is needed for atomicity.
	obj := g.client.
		Bucket(bucket).
		Object(name).
		If(storage.Conditions{DoesNotExist: true})
	w := obj.NewWriter(ctx)
	return &gsDriverObject{
		obj:    obj,
		writer: w,
	}, nil
}

// ParseGcsURI parses a "gs://" URI into a bucket, name pair.
// Inspired by:
// https://github.com/GoogleCloudPlatform/gifinator/blob/master/internal/gcsref/gcsref.go#L37
func parseGcsURI(uri string) (bucket, name string, err error) {
	const prefix = "gs://"
	if !strings.HasPrefix(uri, prefix) {
		return "", "", fmt.Errorf("bad GCS URI %q: scheme is not %q", uri, prefix)
	}
	uri = uri[len(prefix):]
	i := strings.IndexByte(uri, '/')
	if i == -1 {
		return "", "", fmt.Errorf("bad GCS URI %q: no object name", uri)
	}
	bucket, name = uri[:i], uri[i+1:]
	return bucket, name, nil
}

// NewGCSDriver returns a new atomic writer backed by Google Cloud Storage.
func NewGCSDriver(ctx context.Context, opts ...option.ClientOption) (DriverInterface, error) {
	c, err := storage.NewClient(ctx, opts...)
	if err != nil {
		return nil, err
	}
	return &gsDriver{
		client: c,
	}, nil
}

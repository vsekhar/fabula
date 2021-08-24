package atomicwriter

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	"cloud.google.com/go/storage"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

var errExist = os.ErrExist

// DriverInterface is the interface an atomic writer driver fulfills. It can be
// used to create atomic writers.
type DriverInterface interface {
	NewAtomicWriter(context.Context, string) (Interface, error)
	Exists(context.Context, string) (bool, error)
}

// Interface is the interface an individual atomic writer fulfills.
type Interface interface {
	io.Writer

	// CloseAtomically will attempt to close the file, atomically committing it
	// to storage. The file will either be fully written, or an error is
	// returned.
	//
	// Clients can use os.IsExist(err) to check if the error was due to a name
	// conflict.
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
		// Link will return something that satisfies os.IsExist(), so
		return err
	}
	return nil
}

type fsDriver struct {
	dir string
}

func (fd *fsDriver) NewAtomicWriter(_ context.Context, name string) (Interface, error) {
	path := filepath.Join(fd.dir, name)
	d := filepath.Dir(path) // name might include additional directory separators
	tfile, err := ioutil.TempFile(d, fsPattern)
	if err != nil {
		return nil, err
	}
	return &fsDriverObject{
		path:  path,
		tfile: tfile,
	}, nil
}

func (fd *fsDriver) Exists(_ context.Context, name string) (bool, error) {
	_, err := os.Stat(filepath.Join(fd.dir, name))
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// NewFileSystemDriver returns a new atomic writer backed by the local
// file system.
func NewFileSystemDriver(dir string) DriverInterface {
	return &fsDriver{dir: dir}
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

	err := gsdo.writer.Close()
	switch ee := err.(type) {
	case *googleapi.Error:
		if ee.Code == http.StatusPreconditionFailed {
			// TODO: can we determine if it was the DoesNotExist precondition
			// specifically, as opposed to any other precondition? Maybe it's ok
			// since this package is the only place where preconditions can be
			// applied and we only set DoesNotExist in NewAtomicWriter.
			return errExist
		}
	}
	return err
}

type gsDriver struct {
	bkt    *storage.BucketHandle
	prefix string
}

func (g *gsDriver) NewAtomicWriter(ctx context.Context, name string) (Interface, error) {
	path := filepath.Join(g.prefix, name)
	// Important: DoesNotExist condition here is needed for atomicity.
	obj := g.bkt.Object(path).If(storage.Conditions{DoesNotExist: true})
	return &gsDriverObject{
		obj:    obj,
		writer: obj.NewWriter(ctx),
	}, nil
}

func (g *gsDriver) Exists(ctx context.Context, name string) (bool, error) {
	_, err := g.bkt.Object(filepath.Join(g.prefix, name)).Attrs(ctx)
	if err == nil {
		return true, nil
	}
	if err == storage.ErrObjectNotExist {
		return false, nil
	}
	return false, err
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
		bucket = uri
		name = ""
	} else {
		bucket, name = uri[:i], uri[i+1:]
	}
	return bucket, name, nil
}

var gcsClient *storage.Client
var gcsClientOnce sync.Once

// NewGCSDriver returns a new atomic writer backed by Google Cloud Storage.
func NewGCSDriver(ctx context.Context, dir string, opts ...option.ClientOption) (DriverInterface, error) {
	var err error
	gcsClientOnce.Do(func() {
		gcsClient, err = storage.NewClient(ctx, opts...)
	})
	if err != nil {
		return nil, err
	}
	bucket, prefix, err := parseGcsURI(dir)
	if err != nil {
		return nil, err
	}
	// TODO: specify Bucket().UserProject(cred.ProjectID)
	return &gsDriver{
		bkt:    gcsClient.Bucket(bucket),
		prefix: prefix,
	}, nil
}

// NewDriver returns a driver for the specified path.
func NewDriver(ctx context.Context, dir string) (DriverInterface, error) {
	switch {
	case strings.HasPrefix(dir, "gs://"):
		return NewGCSDriver(ctx, dir)
	default:
		return NewFileSystemDriver(dir), nil
	}
}

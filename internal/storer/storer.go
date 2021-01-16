package storer

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"hash/crc32"
	"strings"

	"cloud.google.com/go/storage"
	pb "github.com/vsekhar/fabula/internal/api/storer"
	"golang.org/x/crypto/sha3"
	"google.golang.org/api/googleapi"
	"google.golang.org/protobuf/proto"
)

const hashLen = 64

// Storer permits uncoordinated writes to storage.
type Storer struct {
	bkt    string
	fld    string
	pfx    string
	objpfx string
	cli    *storage.Client
	bktH   *storage.BucketHandle
}

// New returns a new storer object.
func New(ctx context.Context, bkt, fld, pfx string) (*Storer, error) {
	r := &Storer{
		bkt: bkt,
		fld: fld,
		pfx: pfx,
		objpfx: strings.Join([]string{
			fld,
			pfx,
		}, "/"),
	}
	scli, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	r.cli = scli
	r.bktH = scli.Bucket(bkt)
	return r, nil
}

func (s *Storer) objName(sum []byte) string {
	return strings.Join([]string{
		s.objpfx,
		base64.RawURLEncoding.EncodeToString(sum),
	}, "")
}

// StoreOrVerify performs one of three actions:
//
//   1. Write req to storage, returning the objectName, true and nil
//   2. If an object already exists in storage with a matching name, verify that
//      the object in storage is identical to req; if so, return objName, false,
//      and nil
//   3. Return an error
func (s *Storer) StoreOrVerify(ctx context.Context, req *pb.StoreRequest) (objName string, stored bool, err error) {
	h := sha3.NewShake256()
	for _, b := range req.Batch.Batches {
		h.Write(b.BatchSha3512)
	}
	for _, e := range req.Batch.Entries {
		h.Write(e.NotarizationSha3512)
	}
	var sumArr [hashLen]byte
	sum := sumArr[:]
	if n, err := h.Read(sum); n != hashLen || err != nil {
		return "", false, fmt.Errorf("reading hash: %w", err)
	}
	if !bytes.Equal(sum, req.Ref.BatchSha3512) {
		return "", false, fmt.Errorf("storer and packer hashes don't match: storer: '%s', packer: '%s'", sum, req.Ref.BatchSha3512)
	}
	objName = s.objName(sum)
	obj := s.bktH.Object(objName).If(storage.Conditions{
		DoesNotExist: true,
	})
	buf, err := proto.Marshal(req.Batch)
	if err != nil {
		return "", false, err
	}
	w := obj.NewWriter(ctx)
	w.SendCRC32C = true
	w.CRC32C = crc32.ChecksumIEEE(buf)
	if n, err := w.Write(buf); n != len(buf) || err != nil {
		return "", false, fmt.Errorf("short write (%d bytes) or error: %w", n, err)
	}
	if err := w.Close(); err != nil {
		if e, ok := err.(*googleapi.Error); ok {
			if e.Code == 412 && e.Errors[0].Reason == "conditionNotMet" {
				// exists
				if err := s.Verify(ctx, objName, w.CRC32C); err != nil {
					return "", false, err // failed to verify
				}
				return objName, false, nil // verified existing object matches req
			}
		}
		// Some other error
		return "", false, fmt.Errorf("closing writer: %w", err)
	}
	return objName, true, nil // stored new object
}

// Verify verifies objName has crc32c in the bucket.
func (s *Storer) Verify(ctx context.Context, objName string, crc32c uint32) error {
	attrs, err := s.bktH.Object(objName).Attrs(ctx)
	if err != nil {
		return err
	}
	if attrs.CRC32C == crc32c {
		return nil
	}
	return fmt.Errorf("CRC32C does not match on existing object '%s': expected %d, got %d", objName, crc32c, attrs.CRC32C)
}

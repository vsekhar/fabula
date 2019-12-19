package notary_test

import (
	"fmt"
	"testing"

	"github.com/vsekhar/COMMIT/notary"
)

func Example() {
	svc, _ := notary.NewService()
	data := []byte("hello world")
	n, _ := svc.Notarize(data)
	fmt.Printf("Notarize('%s') --> (%s, %.8x..., %.8x...)", data, n.Timestamp, n.Salt, n.Signature)
}

func TestNotarize(t *testing.T) {
	svc, err := notary.NewService()
	if err != nil {
		t.Fatal(err)
	}
	data := []byte("hello world")
	n, err := svc.Notarize(data)
	if err != nil {
		t.Error(err)
	}
	if !notary.ValidateNotarization(data, n) {
		t.Errorf("Bad signature")
	}
}

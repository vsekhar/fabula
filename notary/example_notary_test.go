package notary_test

import "github.com/vsekhar/COMMIT/notary"

func Example() {
	svc, _ := notary.NewService()
	_ = notary.NewClient(svc)
}

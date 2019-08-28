package faketime

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vsekhar/COMMIT/pkg/commit"
)

func TestFakeTimeNow(t *testing.T) {
	req, err := http.NewRequest("GET", commit.TrueTimeNowPath, nil)
	if err != nil {
		t.Fatal(err)
	}

	jitterMs := float64(10)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(NowHandler(jitterMs))
	handler.ServeHTTP(rr, req)
	resp := rr.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", resp.StatusCode, http.StatusOK)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}
	if len(body) != 0 {
		t.Errorf("unexpected body: %s", body)
	}
	// TODO: check headers
}

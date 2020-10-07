package youtime

import (
	"context"
	"testing"
)

func TestYouTime(t *testing.T) {
	c := NewClient(context.Background())
	c.Ready()
}

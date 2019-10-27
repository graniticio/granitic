package ws

import (
	"context"
	"testing"
)

func TestStoreRecoverID(t *testing.T) {

	ctx := context.Background()

	if RecoverIDFunction(ctx) != nil {

		t.Fail()

	}

	if RequestID(ctx) != "" {
		t.Fail()
	}

	rf := func(context.Context) string {
		return "A"
	}

	ctx = StoreRequestIDFunction(ctx, rf)

	if RecoverIDFunction(ctx) == nil {

		t.Fail()

	}

	if RequestID(ctx) != "A" {
		t.Fail()
	}
}

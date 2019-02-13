package instance

import "testing"

func TestNewIdentifier(t *testing.T) {

	i := NewIdentifier("my-id")

	if i.ID != "my-id" {
		t.Error()
	}

}

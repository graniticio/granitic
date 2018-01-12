package schedule

import (
	"testing"
	"time"
)

func TestQueryAutoBinding(t *testing.T) {

	i, e := parseEvery("10 seconds")

	expect(t, time.Duration(0), time.Duration(10*time.Second), nil, i.OffsetFromStart, i.Frequency, e)

}

func expect(t *testing.T, ed time.Duration, ef time.Duration, ee error, ad time.Duration, af time.Duration, ae error) {

	if ed != ad {
		t.Errorf("Expected %v actual %v", ed, ad)
	}

	if ef != af {
		t.Errorf("Expected %v actual %v", ef, af)
	}

	if ee != ae {
		t.Errorf("Expected %v actual %v", ee, ae)
	}

}

package schedule

import (
	"testing"
	"time"
)

func TestEvery10Seconds(t *testing.T) {

	i, e := parseEvery("10 seconds")

	expect(t, time.Duration(0), time.Duration(10*time.Second), nil, i.OffsetFromStart, i.Frequency, e)

}

func TestEveryDay(t *testing.T) {

	_, e := parseEvery("day")

	if e != nil {
		t.Errorf(e.Error())
		t.FailNow()
	}
}

func TestEveryDayAtTime(t *testing.T) {

	_, e := parseEvery("day at 1200")

	if e != nil {
		t.Errorf(e.Error())
		t.FailNow()
	}
}

func TestEveryVariants(t *testing.T) {

	patterns := []string{
		"Day", "1 day", "7  days", "10 seconds", "second", " 8 hours ", "30 minutes", "MinuTE", "1 hour", "2 hours", "day at 1200", "minute at HHMM13",
	}

	for _, p := range patterns {

		if _, e := parseEvery(p); e != nil {
			t.Errorf(e.Error())
			t.FailNow()
		}

	}

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

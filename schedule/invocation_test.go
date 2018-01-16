package schedule

import (
	"testing"
)

func TestEmptyQueue(t *testing.T) {

	iq := new(invocationQueue)
	c := iq.Contents()

	if len(c) != 0 {
		t.Fail()
	}

}

func TestSingleQueue(t *testing.T) {

	iq := new(invocationQueue)

	i1 := new(invocation)
	i1.counter = 1

	iq.Enqueue(i1)
	c := iq.Contents()

	if len(c) != 1 {
		t.FailNow()
	}

	if c[0].counter != 1 {
		t.Fail()
	}

}

func TestQueueMulti(t *testing.T) {

	iq := new(invocationQueue)

	i1 := new(invocation)
	i1.counter = 1

	iq.Enqueue(i1)
	c := iq.Contents()

	if len(c) != 1 {
		t.FailNow()
	}

	if c[0].counter != 1 {
		t.Fail()
	}

	i2 := new(invocation)
	i2.counter = 2

	iq.Enqueue(i2)
	c = iq.Contents()

	if len(c) != 2 {
		t.FailNow()
	}

	if c[0].counter != 1 {
		t.Fail()
	}

	if c[1].counter != 2 {
		t.Fail()
	}

	i3 := new(invocation)
	i3.counter = 3

	iq.Enqueue(i3)
	c = iq.Contents()

	if len(c) != 3 {
		t.FailNow()
	}

	if c[0].counter != 1 {
		t.Fail()
	}

	if c[1].counter != 2 {
		t.Fail()
	}

	if c[2].counter != 3 {
		t.Fail()
	}
}

func TestDequeueEmpty(t *testing.T) {
	iq := new(invocationQueue)
	i := iq.Dequeue()

	if i != nil {
		t.FailNow()
	}
}

func TestRemoveEmpty(t *testing.T) {
	iq := new(invocationQueue)

	iq.Remove(0)
}

func TestRemoveOnly(t *testing.T) {
	iq := buildTo(1)

	iq.Remove(1)

	if iq.Size() != 0 {
		t.FailNow()
	}

}

func TestRemoveNonExistent(t *testing.T) {
	iq := buildTo(11)

	iq.Remove(88)

	if iq.Size() != 11 {
		t.FailNow()
	}

}

func TestRemoveMiddleMany(t *testing.T) {
	iq := buildTo(11)

	iq.Remove(7)

	if iq.Size() != 10 {
		t.FailNow()
	}

	c := iq.Contents()

	if c[6].counter != 8 {
		t.FailNow()
	}

}

func TestRemoveHeadMany(t *testing.T) {
	iq := buildTo(11)

	iq.Remove(1)

	if iq.Size() != 10 {
		t.FailNow()
	}

	h := iq.PeekHead()

	if h.counter != 2 {
		t.FailNow()
	}

}


func buildTo(max int) *invocationQueue {

	iq := new(invocationQueue)

	for i := 1; i <= max; i++ {

		in := new(invocation)
		in.counter = uint64(i)
		iq.Enqueue(in)
	}

	return iq

}

func TestDequeueMulti(t *testing.T) {
	iq := new(invocationQueue)
	i := iq.Dequeue()

	if i != nil {
		t.FailNow()
	}

	i1 := new(invocation)
	i1.counter = 1

	iq.Enqueue(i1)

	i2 := new(invocation)
	i2.counter = 2

	iq.Enqueue(i2)

	i3 := new(invocation)
	i3.counter = 3

	iq.Enqueue(i3)

	di := iq.Dequeue()

	if di.counter != 1 {
		t.FailNow()
	}

	c := iq.Contents()

	if len(c) != 2 {
		t.FailNow()
	}

	di = iq.Dequeue()

	if di.counter != 2 {
		t.FailNow()
	}

	c = iq.Contents()

	if len(c) != 1 {
		t.FailNow()
	}

	di = iq.Dequeue()

	if di.counter != 3 {
		t.FailNow()
	}

	c = iq.Contents()

	if len(c) != 0 {
		t.FailNow()
	}

	di = iq.Dequeue()

	if di != nil {
		t.FailNow()
	}

	i4 := new(invocation)
	i4.counter = 4

	iq.Enqueue(i4)

	i5 := new(invocation)
	i5.counter = 5

	iq.Enqueue(i5)

	i6 := new(invocation)
	i6.counter = 6

	iq.Enqueue(i6)

	c = iq.Contents()

	if len(c) != 3 {
		t.FailNow()
	}

	if c[0].counter != 4 || c[1].counter != 5 || c[2].counter != 6 {
		t.FailNow()
	}

}

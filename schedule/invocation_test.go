package schedule

import (
	"fmt"
	"testing"
)

func TestRetryManagement(t *testing.T) {

	i := newInvocation(0, 0)

	if !i.firstAttempt() {
		t.Fail()
	}

	i.attempt++

	if i.retryAllowed() {
		t.Fail()
	}

	if i.firstAttempt() {
		t.Fail()
	}

	i = newInvocation(0, 3)

	i.attempt++
	i.attempt++

	if !i.retryAllowed() {
		t.Fail()
	}

	i.attempt++

	if i.retryAllowed() {
		t.Fail()
	}

}

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

	iq.EnqueueAtTail(i1)
	c := iq.Contents()

	if len(c) != 1 {
		t.FailNow()
	}

	if c[0].counter != 1 {
		t.Fail()
	}

}

func TestEnqueueHead(t *testing.T) {

	iq := new(invocationQueue)

	i1 := new(invocation)
	i1.counter = 1

	iq.EnqueueAtTail(i1)

	i2 := new(invocation)
	i2.counter = 2

	iq.EnqueueAtHead(i2)

	c := iq.Contents()

	if c[0].counter != 2 || iq.PeekHead().counter != 2 {
		t.FailNow()
	}

	if c[1].counter != 1 || iq.PeekTail().counter != 1 {
		t.FailNow()
	}

	iq = new(invocationQueue)

	iq.EnqueueAtHead(i2)

	if c[0].counter != 2 || iq.PeekHead().counter != 2 {
		t.FailNow()
	}

	iq = new(invocationQueue)

	iq.EnqueueAtTail(i1)
	iq.EnqueueAtTail(i2)

	i3 := new(invocation)
	i3.counter = 3

	iq.EnqueueAtHead(i3)

	c = iq.Contents()

	if c[0].counter != 3 || iq.PeekHead().counter != 3 {
		t.FailNow()
	}

	if c[2].counter != 2 || iq.PeekTail().counter != 2 {
		t.FailNow()
	}

}

func TestQueueMulti(t *testing.T) {

	iq := new(invocationQueue)

	i1 := new(invocation)
	i1.counter = 1

	iq.EnqueueAtTail(i1)
	c := iq.Contents()

	if len(c) != 1 {
		t.FailNow()
	}

	if c[0].counter != 1 {
		t.Fail()
	}

	i2 := new(invocation)
	i2.counter = 2

	iq.EnqueueAtTail(i2)
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

	iq.EnqueueAtTail(i3)
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

func TestRemoveTailMany(t *testing.T) {
	iq := buildTo(11)

	iq.Remove(11)

	if iq.Size() != 10 {
		t.FailNow()
	}

	h := iq.PeekTail()

	if h.counter != 10 {
		t.FailNow()
	}

}

func TestRemoveMultiMany(t *testing.T) {
	iq := buildTo(11)

	iq.Remove(10)
	iq.Remove(11)
	iq.Remove(5)
	iq.Remove(9)
	iq.Remove(1)
	iq.Remove(2)

	if iq.Size() != 5 {
		t.FailNow()
	}

	h := iq.PeekTail()

	if h.counter != 8 {
		fmt.Println(h.counter)

		t.FailNow()
	}

}

func buildTo(max int) *invocationQueue {

	iq := new(invocationQueue)

	for i := 1; i <= max; i++ {

		in := new(invocation)
		in.counter = uint64(i)
		iq.EnqueueAtTail(in)
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

	iq.EnqueueAtTail(i1)

	i2 := new(invocation)
	i2.counter = 2

	iq.EnqueueAtTail(i2)

	i3 := new(invocation)
	i3.counter = 3

	iq.EnqueueAtTail(i3)

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

	iq.EnqueueAtTail(i4)

	i5 := new(invocation)
	i5.counter = 5

	iq.EnqueueAtTail(i5)

	i6 := new(invocation)
	i6.counter = 6

	iq.EnqueueAtTail(i6)

	c = iq.Contents()

	if len(c) != 3 {
		t.FailNow()
	}

	if c[0].counter != 4 || c[1].counter != 5 || c[2].counter != 6 {
		t.FailNow()
	}

}

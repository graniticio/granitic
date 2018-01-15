// Copyright 2018 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package schedule

type invocation struct {
	nextInvocation *invocation
	counter        uint64
}

type invocationQueue struct {
	head *invocation
	tail *invocation
}

func (iq *invocationQueue) Enqueue(i *invocation) {

	if iq.head == nil {
		iq.head = i
	} else if iq.tail == nil {
		iq.head.nextInvocation = i
		iq.tail = i
	} else {
		iq.tail.nextInvocation = i
		iq.tail = i
	}

}

func (iq *invocationQueue) Dequeue() *invocation {

	ch := iq.head

	if ch != nil {

		iq.head = ch.nextInvocation

		if ch.nextInvocation == nil || iq.head.nextInvocation == nil {
			iq.tail = nil
		}
	}

	return ch
}

func (iq *invocationQueue) Counters() []uint64 {

	c := make([]uint64, 0)

	return iq.addCounter(iq.head, c)
}

func (iq *invocationQueue) addCounter(i *invocation, c []uint64) []uint64 {

	if i == nil {

		return c
	} else {
		c = append(c, i.counter)
		return iq.addCounter(i.nextInvocation, c)
	}
}

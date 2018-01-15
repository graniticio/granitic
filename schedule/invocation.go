// Copyright 2018 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package schedule

import "time"

type invocation struct {
	counter   uint64
	runAt     time.Time
	startedAt time.Time
}

type invocationQueue struct {
	head *queueMember
	tail *queueMember
}

type queueMember struct {
	i *invocation
	n *queueMember
}

func (iq *invocationQueue) Enqueue(i *invocation) {

	qm := new(queueMember)
	qm.i = i

	if iq.head == nil {
		iq.head = qm
	} else if iq.tail == nil {
		iq.head.n = qm
		iq.tail = qm
	} else {
		iq.tail.n = qm
		iq.tail = qm
	}

}

func (iq *invocationQueue) Peek() *invocation {

	if iq.head == nil {
		return nil
	}

	return iq.head.i
}

func (iq *invocationQueue) Dequeue() *invocation {

	ch := iq.head

	if ch != nil {

		iq.head = ch.n

		if ch.n == nil || iq.head.n == nil {
			iq.tail = nil
		}

		return ch.i

	} else {
		return nil
	}

}

func (iq *invocationQueue) Contents() []*invocation {

	c := make([]*invocation, 0)

	return iq.addToContents(iq.head, c)
}

func (iq *invocationQueue) addToContents(qm *queueMember, c []*invocation) []*invocation {

	if qm == nil {

		return c
	} else {
		c = append(c, qm.i)
		return iq.addToContents(qm.n, c)
	}
}

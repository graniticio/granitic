// Copyright 2018 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package schedule

import (
	"sync"
	"time"
)

type invocation struct {
	counter   uint64
	runAt     time.Time
	startedAt time.Time
}

type invocationQueue struct {
	head *queueMember
	tail *queueMember
	mux  sync.Mutex
	size uint64
}

type queueMember struct {
	i *invocation
	n *queueMember
}

func (iq *invocationQueue) Enqueue(i *invocation) {

	iq.mux.Lock()

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

	iq.size++

	iq.mux.Unlock()

}

func (iq *invocationQueue) PeekHead() *invocation {

	var result *invocation

	iq.mux.Lock()

	if iq.head == nil {
		result = nil
	}

	result = iq.head.i

	iq.mux.Unlock()

	return result
}

func (iq *invocationQueue) Remove(c uint64) {

	iq.mux.Lock()

	var previous *queueMember

	qm := iq.head

	if qm != nil {

		for qm != nil {

			if qm.i.counter == c {

				if previous == nil {
					//Removing head
					iq.head = nil

				} else if qm.n == nil {
					//Removing tail
					iq.tail = nil
					previous = nil

				} else {
					previous.n = qm.n
				}

				qm = nil
				iq.size--

			} else {
				previous = qm
				qm = qm.n
			}

		}

	}

	iq.mux.Unlock()

}

func (iq *invocationQueue) Size() uint64 {
	return iq.size
}

func (iq *invocationQueue) Dequeue() *invocation {

	iq.mux.Lock()

	ch := iq.head

	var result *invocation

	if ch != nil {

		iq.head = ch.n

		if ch.n == nil || iq.head.n == nil {
			iq.tail = nil
		}

		result = ch.i

	}

	iq.size--

	iq.mux.Unlock()

	return result

}

func (iq *invocationQueue) Contents() []*invocation {

	iq.mux.Lock()

	c := make([]*invocation, 0)

	result := iq.addToContents(iq.head, c)

	iq.mux.Unlock()

	return result

}

func (iq *invocationQueue) addToContents(qm *queueMember, c []*invocation) []*invocation {

	if qm == nil {

		return c
	} else {
		c = append(c, qm.i)
		return iq.addToContents(qm.n, c)
	}
}

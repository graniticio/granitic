// Copyright 2018-2022 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package schedule

import (
	"sync"
	"time"
)

const (
	//Scheduled indicates that this is the invocation was due to a schedule
	Scheduled = "Scheduled"
	//Retry indicates that this is the invocation is a retry of a previously failed invocation
	Retry = "Retry"
	//Manual indicates that this is the invocation was triggered outside of a schedule
	Manual = "Manual"
)

func newInvocation(counter uint64, retries int, reason string) *invocation {
	i := new(invocation)
	i.counter = counter
	i.attempt = 1
	i.maxAttempts = retries + 1
	i.reason = reason

	return i
}

type invocation struct {
	counter     uint64
	runAt       time.Time
	startedAt   time.Time
	attempt     int
	maxAttempts int
	reason      string
}

func (i *invocation) firstAttempt() bool {
	return i.attempt == 1
}

func (i *invocation) retryAllowed() bool {
	return i.attempt < i.maxAttempts
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

func (iq *invocationQueue) EnqueueAtHead(i *invocation) {

	iq.mux.Lock()

	newHead := new(queueMember)
	newHead.i = i

	if iq.head == nil {
		iq.head = newHead
	} else {

		newHead.n = iq.head

		if iq.tail == nil {
			iq.tail = iq.head
		}

		iq.head = newHead

	}

	iq.mux.Unlock()

}

func (iq *invocationQueue) EnqueueAtTail(i *invocation) {

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
	} else {
		result = iq.head.i
	}

	iq.mux.Unlock()

	return result
}

func (iq *invocationQueue) PeekTail() *invocation {

	var result *invocation

	iq.mux.Lock()

	if iq.tail == nil {
		result = nil
	} else {
		result = iq.tail.i
	}

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
					iq.head = qm.n

				} else if qm.n == nil {
					//Removing tail
					iq.tail = previous
					iq.tail.n = nil

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
	}

	c = append(c, qm.i)
	return iq.addToContents(qm.n, c)

}

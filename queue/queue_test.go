// Copyright (C) 2015 The Protogalaxy Project
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
package queue_test

import (
	"sync"
	"testing"
	"time"

	"github.com/protogalaxy/service-notify/queue"
)

func TestChannelQueueSendMessage(t *testing.T) {
	t.Parallel()
	result := make(chan queue.QueuedMessage, 1)
	q := queue.NewChannelQueue(func(c <-chan queue.QueuedMessage) {
		result <- <-c
	})
	go q.Start()
	defer q.Close()
	msg := queue.QueuedMessage{
		UserId: "userid",
		Data:   []byte("data"),
	}
	select {
	case q.Messages() <- msg:
	case <-time.After(time.Millisecond):
		t.Fatalf("Unable to send message to queue")
	}
	select {
	case msg := <-result:
		if msg.UserId != "userid" {
			t.Errorf("Wrong message user id: 'userid' != %s", msg.UserId)
		}
		if string(msg.Data) != "data" {
			t.Errorf("Wrong message data: 'data' != %s", string(msg.Data))
		}
	case <-time.After(time.Millisecond):
		t.Fatal("Message was not sent to worker")
	}
}

func TestChannelQueueNumWorkersStarted(t *testing.T) {
	t.Parallel()
	var lock sync.Mutex
	var counter int
	q := queue.NewChannelQueue(func(c <-chan queue.QueuedMessage) {
		lock.Lock()
		defer lock.Unlock()
		counter += 1
	}, func(p *queue.Properties) {
		p.NumWorkers = 4
	})
	go q.Close()
	q.Start()

	lock.Lock()
	defer lock.Unlock()
	if counter != 4 {
		t.Fatalf("Wrong numbeer of workers started: expecting 4 but got %d", counter)
	}
}

func TestChannelQueueSize(t *testing.T) {
	t.Parallel()
	var lock sync.Mutex
	var size int
	q := queue.NewChannelQueue(func(c <-chan queue.QueuedMessage) {
		lock.Lock()
		defer lock.Unlock()
		size = cap(c)
	}, func(p *queue.Properties) {
		p.NumWorkers = 1
		p.QueueSize = 11
	})
	// We must start and block so we make sure that worker goroutines are created.
	go q.Close()
	q.Start()

	lock.Lock()
	defer lock.Unlock()
	if size != 11 {
		t.Fatalf("Wrong queue size: expecting 11 but got %d", size)
	}
}

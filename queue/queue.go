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
package queue

import "sync"

type QueuedMessage struct {
	UserId string
	Data   []byte
}

const (
	DefaultWorkers   = 10
	DefaultQueueSize = 100
)

type MessageQueue interface {
	Messages() chan<- QueuedMessage
}

type Properties struct {
	NumWorkers int
	QueueSize  int
}

type ChannelQueue struct {
	properties *Properties
	worker     func(<-chan QueuedMessage)
	messages   chan QueuedMessage
	done       chan struct{}
	once       sync.Once
}

func NewChannelQueue(worker func(<-chan QueuedMessage), conf ...func(*Properties)) *ChannelQueue {
	p := &Properties{
		NumWorkers: DefaultWorkers,
		QueueSize:  DefaultQueueSize,
	}
	for _, f := range conf {
		f(p)
	}
	return &ChannelQueue{
		properties: p,
		worker:     worker,
		messages:   make(chan QueuedMessage, p.QueueSize),
		done:       make(chan struct{}),
	}
}

func (q *ChannelQueue) Start() {
	for i := 0; i < q.properties.NumWorkers; i++ {
		go q.worker(q.messages)
	}

	<-q.done
}

func (q *ChannelQueue) Messages() chan<- QueuedMessage {
	return q.messages
}

func (q *ChannelQueue) Close() {
	q.once.Do(func() {
		close(q.done)
		close(q.messages)
	})
}

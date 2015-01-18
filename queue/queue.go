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

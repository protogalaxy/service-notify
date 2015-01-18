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
		counter += 1
		lock.Unlock()
	}, func(p *queue.Properties) {
		p.NumWorkers = 4
	})
	go q.Close()
	q.Start()
	if counter != 4 {
		t.Fatalf("Wrong numbeer of workers started: expecting 4 but got %d", counter)
	}
}

func TestChannelQueueSize(t *testing.T) {
	t.Parallel()
	var size int
	q := queue.NewChannelQueue(func(c <-chan queue.QueuedMessage) {
		size = cap(c)
	}, func(p *queue.Properties) {
		p.NumWorkers = 1
		p.QueueSize = 11
	})
	// We must start and block so we make sure that worker goroutines are created.
	go q.Close()
	q.Start()
	if size != 11 {
		t.Fatalf("Wrong queue size: expecting 11 but got %d", size)
	}
}

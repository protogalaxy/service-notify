package service_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/arjantop/saola/httpservice"
	"github.com/protogalaxy/service-notify/queue"
	"github.com/protogalaxy/service-notify/service"
	"golang.org/x/net/context"
)

type QueueMock struct {
	Chan chan queue.QueuedMessage
}

func NewQueueMock(n int) *QueueMock {
	return &QueueMock{
		Chan: make(chan queue.QueuedMessage, n),
	}
}

func (q *QueueMock) Messages() chan<- queue.QueuedMessage {
	return q.Chan
}

func TestNotifyUserDevices(t *testing.T) {
	t.Parallel()
	queue := NewQueueMock(10)
	s := &service.NotifyUserDevices{
		Queue: queue,
	}
	w := httptest.NewRecorder()
	ps := httpservice.EmptyParams()
	ps.Set("userId", "user")
	r, _ := http.NewRequest("POST", "", strings.NewReader("test"))
	if err := s.DoHTTP(httpservice.WithParams(context.Background(), ps), w, r); err != nil {
		t.Fatalf("No error expected but got: %s", err)
	}
	if w.Code != http.StatusAccepted {
		t.Fatalf("Unexpected status code: %d != %d", http.StatusAccepted, w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json; utf-8" {
		t.Fatalf("Invalid context type header: %s", ct)
	}
	if strings.TrimSpace(w.Body.String()) != "{}" {
		t.Fatalf("Body should be an empty json document but got '%s'", w.Body.String())
	}
	select {
	case msg := <-queue.Chan:
		if msg.UserId != "user" {
			t.Errorf("Unexpected user id: 'user' != '%s'", msg.UserId)
		}
		if string(msg.Data) != "test" {
			t.Errorf("Unexpected user id: 'test' != '%s'", string(msg.Data))
		}
	case <-time.After(time.Millisecond):
		t.Fatalf("No value in the queue")
	}
}

type errorReader struct{}

func (r errorReader) Read(b []byte) (int, error) {
	return 0, errors.New("error")
}

func TestNotifyUserDevicesBodyDataError(t *testing.T) {
	t.Parallel()
	s := &service.NotifyUserDevices{}
	r, _ := http.NewRequest("POST", "", errorReader{})
	if err := s.DoHTTP(context.Background(), nil, r); err == nil {
		t.Fatal("Error expected but got nil")
	}
}

func TestNotifyUserDevicesContextCancelled(t *testing.T) {
	t.Parallel()
	queue := NewQueueMock(0)
	s := &service.NotifyUserDevices{
		Queue: queue,
	}
	r, _ := http.NewRequest("POST", "", strings.NewReader(""))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := s.DoHTTP(ctx, nil, r); err != context.Canceled {
		t.Fatalf("Expecting context cancelled error but got: %s", err)
	}
}

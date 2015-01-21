package queue_test

import (
	"errors"
	"testing"
	"time"

	"github.com/protogalaxy/service-notify/queue"
	"golang.org/x/net/context"
)

type messagingMock struct {
	GetUserDevicesFunc    func(ctx context.Context, userId string) (*queue.UserDevices, error)
	SocketSendMessageFunc func(ctx context.Context, deviceId string, data []byte) error
}

func (m messagingMock) GetUserDevices(ctx context.Context, userId string) (*queue.UserDevices, error) {
	if m.GetUserDevicesFunc != nil {
		return m.GetUserDevicesFunc(ctx, userId)
	}
	return nil, nil
}

func (m messagingMock) SocketSendMessage(ctx context.Context, deviceId string, data []byte) error {
	if m.SocketSendMessageFunc != nil {
		return m.SocketSendMessageFunc(ctx, deviceId, data)
	}
	return nil
}

func TestWorkerHandlerMessagesAreSent(t *testing.T) {
	called := make(map[string][]byte)
	m := messagingMock{
		GetUserDevicesFunc: func(ctx context.Context, userId string) (*queue.UserDevices, error) {
			return &queue.UserDevices{
				UserId: "user1",
				Devices: []queue.Device{
					queue.Device{"websocket", "id1"},
					queue.Device{"websocket", "id2"},
				},
			}, nil
		},
		SocketSendMessageFunc: func(ctx context.Context, deviceId string, data []byte) error {
			called[deviceId] = data
			return nil
		},
	}
	h := queue.MessageHandler(m)
	h(queue.QueuedMessage{"user1", []byte("data")})
	if len(called) != 2 {
		t.Errorf("Was expecting sending data to 2 devices but only sent %d", len(called))
	}
	if d := called["id1"]; string(d) != "data" {
		t.Errorf("Incorrect data sent to device: 'data' != '%s'", string(d))
	}
	if d := called["id2"]; string(d) != "data" {
		t.Errorf("Incorrect data sent to device: 'data' != '%s'", string(d))
	}
}

func TestWorkerHandlerInvalidDeviceType(t *testing.T) {
	var called bool
	m := messagingMock{
		GetUserDevicesFunc: func(ctx context.Context, userId string) (*queue.UserDevices, error) {
			return &queue.UserDevices{
				UserId: "user1",
				Devices: []queue.Device{
					queue.Device{"sometype", ""},
				},
			}, nil
		},
		SocketSendMessageFunc: func(ctx context.Context, deviceId string, data []byte) error {
			called = true
			return nil
		},
	}
	h := queue.MessageHandler(m)
	h(queue.QueuedMessage{"user1", []byte("data")})
	if called {
		t.Error("No data should be sent if the device type is unsupported")
	}
}

func TestWorkerGetUserDevicesError(t *testing.T) {
	var called bool
	m := messagingMock{
		GetUserDevicesFunc: func(ctx context.Context, userId string) (*queue.UserDevices, error) {
			return nil, errors.New("error")
		},
		SocketSendMessageFunc: func(ctx context.Context, deviceId string, data []byte) error {
			called = true
			return nil
		},
	}
	h := queue.MessageHandler(m)
	h(queue.QueuedMessage{"user1", []byte("data")})
	if called {
		t.Error("No data should be sent if user devices are not retrieved")
	}
}

func TestWorkerSendsMessagesToHandler(t *testing.T) {
	t.Parallel()
	c := make(chan queue.QueuedMessage, 100)
	w := queue.Worker{
		MessageHandler: func(msg queue.QueuedMessage) {
			c <- msg
		},
	}
	go w.Do(c)
	sendMessage(t, c, "u1", "d1")
	sendMessage(t, c, "u2", "d2")
	close(c)
	assertReceived(t, c, "u1", "d1")
	assertReceived(t, c, "u2", "d2")
}

func sendMessage(t *testing.T, c chan<- queue.QueuedMessage, userId, data string) {
	select {
	case c <- queue.QueuedMessage{userId, []byte(data)}:
	case <-time.After(time.Millisecond):
		t.Fatalf("Unable to send the message the worker")
	}
}

func assertReceived(t *testing.T, c <-chan queue.QueuedMessage, userId, data string) {
	select {
	case msg := <-c:
		if msg.UserId != userId {
			t.Errorf("Invalid user id: '%s' != '%s'", userId, msg.UserId)
		}
		if string(msg.Data) != data {
			t.Errorf("Invalid message data: '%s' != '%s'", data, string(msg.Data))
		}
	case <-time.After(time.Millisecond):
		t.Fatalf("Unable to send the message the worker")
	}
}

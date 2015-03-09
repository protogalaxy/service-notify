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
	"errors"
	"io"
	"testing"
	"time"

	"github.com/protogalaxy/service-notify/devicepresence"
	"github.com/protogalaxy/service-notify/queue"
	"github.com/protogalaxy/service-notify/socket"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type SenderMock struct {
	OnSendMessage func(ctx context.Context, in *socket.SendRequest, opts ...grpc.CallOption) (*socket.SendReply, error)
}

func (m SenderMock) SendMessage(ctx context.Context, in *socket.SendRequest, opts ...grpc.CallOption) (*socket.SendReply, error) {
	return m.OnSendMessage(ctx, in, opts...)
}

type PresenceManagerMock struct {
	OnGetDevices func(ctx context.Context, in *devicepresence.DevicesRequest, opts ...grpc.CallOption) (devicepresence.PresenceManager_GetDevicesClient, error)
}

func (m PresenceManagerMock) GetDevices(ctx context.Context, in *devicepresence.DevicesRequest, opts ...grpc.CallOption) (devicepresence.PresenceManager_GetDevicesClient, error) {
	return m.OnGetDevices(ctx, in, opts...)
}

type DeviceStream struct {
	devices []*devicepresence.Device
	grpc.ClientStream
}

func MockDeviceStream(devices ...*devicepresence.Device) *DeviceStream {
	return &DeviceStream{
		devices: devices,
	}
}

func (s *DeviceStream) Recv() (*devicepresence.Device, error) {
	if len(s.devices) == 0 {
		return nil, io.EOF
	}
	d := s.devices[0]
	s.devices = s.devices[1:]
	return d, nil
}

func TestWorkerHandlerMessagesAreSent(t *testing.T) {
	pmm := PresenceManagerMock{
		OnGetDevices: func(ctx context.Context, in *devicepresence.DevicesRequest, opts ...grpc.CallOption) (devicepresence.PresenceManager_GetDevicesClient, error) {
			if in.UserId != "user1" {
				t.Errorf("Unexpected user: %s", in.UserId)
			}

			return MockDeviceStream(&devicepresence.Device{
				Id:   "111",
				Type: devicepresence.Device_WS,
			}, &devicepresence.Device{
				Id:   "222",
				Type: devicepresence.Device_WS,
			}), nil
		},
	}

	var messagesSent int
	sm := SenderMock{
		OnSendMessage: func(ctx context.Context, in *socket.SendRequest, opts ...grpc.CallOption) (*socket.SendReply, error) {
			messagesSent += 1
			if in.SocketId == 111 {
				if "data" != string(in.Data) {
					t.Errorf("Unexpected message body")
				}
			} else if in.SocketId == 222 {
				if "data" != string(in.Data) {
					t.Errorf("Unexpected message body")
				}
			} else {
				t.Error("Unexpected socket id")
			}
			return nil, nil
		},
	}

	h := queue.MessageHandler(pmm, sm)
	h(queue.QueuedMessage{"user1", []byte("data")})
	if messagesSent != 2 {
		t.Errorf("Message sent to too many devices: %d", messagesSent)
	}
}

func TestWorkerHandlerInvalidDeviceType(t *testing.T) {
	pmm := PresenceManagerMock{
		OnGetDevices: func(ctx context.Context, in *devicepresence.DevicesRequest, opts ...grpc.CallOption) (devicepresence.PresenceManager_GetDevicesClient, error) {
			return MockDeviceStream(&devicepresence.Device{
				Type: devicepresence.Device_Type(9999),
			}), nil
		},
	}

	h := queue.MessageHandler(pmm, nil)
	h(queue.QueuedMessage{"user1", []byte("data")})
}

func TestWorkerGetUserDevicesError(t *testing.T) {
	pmm := PresenceManagerMock{
		OnGetDevices: func(ctx context.Context, in *devicepresence.DevicesRequest, opts ...grpc.CallOption) (devicepresence.PresenceManager_GetDevicesClient, error) {
			return nil, errors.New("error")
		},
	}

	h := queue.MessageHandler(pmm, nil)
	h(queue.QueuedMessage{"user1", []byte("data")})
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

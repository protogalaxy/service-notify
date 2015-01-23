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

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/arjantop/cuirass"
	"github.com/arjantop/saola/httpservice"
	"github.com/golang/glog"
	"github.com/protogalaxy/common/serviceerror"
	"golang.org/x/net/context"
)

type DeviceMessaging interface {
	GetUserDevices(ctx context.Context, userId string) (*UserDevices, error)
	SocketSendMessage(ctx context.Context, deviceId string, data []byte) error
}

type Worker struct {
	MessageHandler func(QueuedMessage)
}

func (w *Worker) Do(messages <-chan QueuedMessage) {
	for msg := range messages {
		w.MessageHandler(msg)
	}
}

func MessageHandler(messaging DeviceMessaging) func(QueuedMessage) {
	return func(msg QueuedMessage) {
		userDevices, err := messaging.GetUserDevices(context.Background(), msg.UserId)
		if err != nil {
			glog.Errorf("Unable to retrieve devices for user %s", msg.UserId)
			return
		}

		for _, device := range userDevices.Devices {
			if device.Type == "websocket" {
				err := messaging.SocketSendMessage(context.Background(), device.Id, msg.Data)
				if err != nil {
					glog.Errorf("Unable to send message to device '%s:%s'", device.Type, device.Id)
				}
			} else {
				glog.Warning("Unsupported device type: %s", device.Type)
			}
		}
	}
}

type Device struct {
	Type string `json:"device_type"`
	Id   string `json:"device_id"`
}

type UserDevices struct {
	UserId  string   `json:"user_id"`
	Devices []Device `json:"devices"`
}

type MessagingClient struct {
	Client   *httpservice.Client
	Executor cuirass.Executor
}

func (c MessagingClient) GetUserDevices(ctx context.Context, userId string) (*UserDevices, error) {
	cmd := cuirass.NewCommand("GetUserDevices", func(ctx context.Context) (interface{}, error) {
		req, err := http.NewRequest("GET", "http://localhost:10000/users/"+userId+"/devices", nil)
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
		if err != nil {
			return nil, err
		}

		res, err := c.Client.Do(ctx, req)
		if err != nil {
			return nil, err
		}
		if res.StatusCode != http.StatusOK {
			return nil, serviceerror.Decode(res.Body)
		}
		var userDevices UserDevices
		decoder := json.NewDecoder(res.Body)
		if err := decoder.Decode(&userDevices); err != nil {
			return nil, err
		}
		return &userDevices, nil
	}).Build()

	r, err := c.Executor.Exec(ctx, cmd)
	if err != nil {
		return nil, err
	}
	if result, ok := r.(*UserDevices); ok {
		return result, nil
	}
	return r.(*UserDevices), nil
}

func (c MessagingClient) SocketSendMessage(ctx context.Context, deviceId string, data []byte) error {
	cmd := cuirass.NewCommand("SocketSendMessage", func(ctx context.Context) (interface{}, error) {
		req, err := http.NewRequest("POST", "http://localhost:11001/websocket/"+deviceId+"/send", bytes.NewReader(data))
		req.Header.Set("Content-Type", "application/octet-stream")
		if err != nil {
			return nil, err
		}

		res, err := c.Client.Do(ctx, req)
		if err != nil {
			return nil, err
		}
		if res.StatusCode != http.StatusAccepted {
			return nil, serviceerror.Decode(res.Body)
		}
		return nil, nil
	}).Build()

	_, err := c.Executor.Exec(ctx, cmd)
	return err
}

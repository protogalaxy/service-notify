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
	"errors"
	"net/http"

	"github.com/arjantop/cuirass"
	"github.com/arjantop/saola/httpservice"
	"github.com/golang/glog"
	"golang.org/x/net/context"
)

type Worker struct {
	Client   *httpservice.Client
	Executor *cuirass.Executor
}

func (w *Worker) Do(messages <-chan QueuedMessage) {
	for msg := range messages {
		getUserDevices := NewGetUserDevicesCommand(w.Client, msg.UserId)
		userDevices, err := ExecGetUserDevicesCommand(context.Background(), w.Executor, getUserDevices)
		if err != nil {
			glog.Errorf("Unable to retrieve devices for user %s", msg.UserId)
			continue
		}

		for _, device := range userDevices.Devices {
			if device.Type == "websocket" {
				socketSendMessage := NewSocketSendMessageCommand(w.Client, device.Id, msg.Data)
				err := ExecSocketSendMessageCommand(context.Background(), w.Executor, socketSendMessage)
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

func NewGetUserDevicesCommand(client *httpservice.Client, userId string) *cuirass.Command {
	return cuirass.NewCommand("GetUserDevices", func(ctx context.Context) (interface{}, error) {
		req, err := http.NewRequest("GET", "http://localhost:10000/users/"+userId+"/devices", nil)
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
		if err != nil {
			return nil, err
		}

		res, err := client.Do(ctx, req)
		if res.StatusCode == http.StatusOK {
			var userDevices UserDevices
			decoder := json.NewDecoder(res.Body)
			if err := decoder.Decode(&userDevices); err != nil {
				return nil, err
			}
			return &userDevices, nil
		} else {
			return nil, err
		}
	}).Build()
}

func ExecGetUserDevicesCommand(ctx context.Context, ex *cuirass.Executor, cmd *cuirass.Command) (*UserDevices, error) {
	r, err := ex.Exec(ctx, cmd)
	if err != nil {
		return nil, err
	}
	if result, ok := r.(*UserDevices); ok {
		return result, nil
	}
	return nil, errors.New("invalid response type")
}

func NewSocketSendMessageCommand(client *httpservice.Client, deviceId string, data []byte) *cuirass.Command {
	return cuirass.NewCommand("SocketSendMessage", func(ctx context.Context) (interface{}, error) {
		req, err := http.NewRequest("POST", "http://localhost:11001/websocket/"+deviceId+"/send", bytes.NewReader(data))
		req.Header.Set("Content-Type", "application/octet-stream")
		if err != nil {
			return nil, err
		}

		_, err = client.Do(ctx, req)
		return nil, err
	}).Build()
}

func ExecSocketSendMessageCommand(ctx context.Context, ex *cuirass.Executor, cmd *cuirass.Command) error {
	_, err := ex.Exec(ctx, cmd)
	return err
}

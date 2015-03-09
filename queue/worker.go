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
	"io"
	"strconv"

	"github.com/golang/glog"
	"github.com/protogalaxy/service-notify/devicepresence"
	"github.com/protogalaxy/service-notify/socket"
	"golang.org/x/net/context"
)

type Worker struct {
	MessageHandler func(QueuedMessage)
}

func (w *Worker) Do(messages <-chan QueuedMessage) {
	for msg := range messages {
		w.MessageHandler(msg)
	}
}

func MessageHandler(presenceClient devicepresence.PresenceManagerClient, socketClient socket.SenderClient) func(QueuedMessage) {
	return func(msg QueuedMessage) {
		// TODO: add timeout
		stream, err := presenceClient.GetDevices(context.Background(), &devicepresence.DevicesRequest{
			UserId: msg.UserId,
		})
		if err != nil {
			glog.Errorf("Unable to retrieve devices for user %s", msg.UserId)
			return
		}

		for {
			device, err := stream.Recv()
			if err == io.EOF {
				break
			} else if err != nil {
				glog.Error("Unable to receive the next device")
			}
			sendMessageToDevice(socketClient, device, msg.Data)
		}
	}
}

func sendMessageToDevice(socketClient socket.SenderClient, device *devicepresence.Device, data []byte) {
	switch device.Type {
	case devicepresence.Device_WS:
		socketID, err := strconv.ParseInt(device.Id, 10, 64)
		if err != nil {
			glog.Errorf("Invalid socket id: %s", err)
			return
		}
		// TODO: add timeout
		_, err = socketClient.SendMessage(context.Background(), &socket.SendRequest{
			SocketId: socketID,
			Data:     data,
		})
		if err != nil {
			glog.Errorf("Unable to send message to device '%s:%s'", device.Type, device.Id)
			return
		}
	default:
		glog.Warning("Unsupported device type: %s", device.Type)
	}
}

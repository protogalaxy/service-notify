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
package service

import (
	"io/ioutil"
	"net/http"

	"github.com/arjantop/saola/httpservice"
	"github.com/golang/glog"
	"github.com/protogalaxy/service-notify/queue"
	"golang.org/x/net/context"
)

type NotifyUserDevices struct {
	Queue queue.MessageQueue
}

// DoHTTP implements saola.HttpService.
func (h *NotifyUserDevices) DoHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	userId := httpservice.GetParams(ctx).Get("userId")
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	msg := queue.QueuedMessage{
		UserId: userId,
		Data:   data,
	}

	select {
	case h.Queue.Messages() <- msg:
		glog.V(3).Infof("Message to user '%s' queued", userId)
	case <-ctx.Done():
		return ctx.Err()
	}

	w.WriteHeader(http.StatusAccepted)
	w.Header().Set("Content-Type", "application/json; utf-8")
	w.Write([]byte("{}\n"))
	return nil
}

// Do implements saola.Service.
func (h *NotifyUserDevices) Do(ctx context.Context) error {
	return httpservice.Do(h, ctx)
}

func (h *NotifyUserDevices) Name() string {
	return "notifyuserdevices"
}

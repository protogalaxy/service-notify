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

//go:generate protoc --go_out=plugins=grpc:. -I ../protos ../protos/notify.proto

package notify

import (
	"errors"

	"github.com/golang/glog"
	"github.com/protogalaxy/service-notify/queue"
	"golang.org/x/net/context"
)

type Notifier struct {
	Queue queue.MessageQueue
}

func (n *Notifier) Send(ctx context.Context, req *SendRequest) (*SendReply, error) {
	if err := validateRequest(req); err != nil {
		return nil, err
	}
	msg := queue.QueuedMessage{
		UserId: req.UserId,
		Data:   req.Data,
	}

	select {
	case n.Queue.Messages() <- msg:
		glog.V(3).Infof("Message for user '%s' queued", req.UserId)
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	return &SendReply{}, nil
}

func validateRequest(req *SendRequest) error {
	if req.UserId == "" {
		return errors.New("missing user id")
	}
	if len(req.Data) == 0 {
		return errors.New("empty message")
	}
	return nil
}

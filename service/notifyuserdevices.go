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

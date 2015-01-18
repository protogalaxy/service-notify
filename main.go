package main

import (
	"flag"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/arjantop/cuirass"
	"github.com/arjantop/saola"
	"github.com/arjantop/saola/httpservice"
	"github.com/arjantop/vaquita"
	"github.com/protogalaxy/common/serviceerror"
	"github.com/protogalaxy/service-notify/queue"
	"github.com/protogalaxy/service-notify/service"
)

func main() {
	flag.Parse()
	rand.Seed(time.Now().UnixNano())

	httpClient := &httpservice.Client{
		Transport: &http.Transport{},
	}

	config := vaquita.NewEmptyMapConfig()
	exec := cuirass.NewExecutor(config)

	worker := &queue.Worker{
		Client:   httpClient,
		Executor: exec,
	}
	queue := queue.NewChannelQueue(worker.Do)
	go queue.Start()

	endpoint := httpservice.NewEndpoint()
	endpoint.POST("/users/:userId/send", saola.Apply(
		&service.NotifyUserDevices{
			Queue: queue,
		},
		httpservice.NewCancellationFilter(),
		serviceerror.NewErrorResponseFilter(),
		serviceerror.NewErrorLoggerFilter()))

	log.Fatal(httpservice.Serve(":12000", saola.Apply(
		endpoint,
		httpservice.NewStdRequestLogFilter())))
}

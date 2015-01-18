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

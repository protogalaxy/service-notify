package main

import (
	"flag"
	"log"
	"os"
	"runtime/pprof"

	"github.com/arjantop/cuirass"
	"github.com/arjantop/vaquita"
	"golang.org/x/net/context"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

func NewBenchCommand() *cuirass.Command {
	return cuirass.NewCommand("BenchCommand", func(ctx context.Context) (interface{}, error) {
		return nil, nil
	}).Build()
}

func main() {
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	cfg := vaquita.NewEmptyMapConfig()
	ex := cuirass.NewExecutor(cfg)
	for i := 0; i < 1000000; i++ {
		ex.Exec(context.Background(), NewBenchCommand())
	}
}

package main

import (
	"net"
	"net/http"
	"os"

	"github.com/heptio/workgroup"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/skpr/k8s-cloudwatchlogs/internal/awslogs"
	"github.com/skpr/k8s-cloudwatchlogs/internal/fileutils"
)

var (
	cliPrometheus = kingpin.Flag("prometheus", "Endpoint which Prometheus metrics can be scraped.").Default(":8080").String()
	cliRegion     = kingpin.Flag("region", "Region which logs will be stored.").Default("ap-southeast-2").String()
	cliIgnore     = kingpin.Flag("ingore", "Ignore lines by using regex.").Default("liveness|healthz").String()
	cliDirectory  = kingpin.Arg("directory", "Directory which contains Kubernetes Pod logs.").Required().String()
)

func main() {
	kingpin.Parse()

	wg := workgroup.Group{}

	// Expose metrics for debugging.
	wg.Add(metrics)

	// Start watching files.
	wg.Add(watcher)

	if err := wg.Run(); err != nil {
		panic(err)
	}
}

// Exposes Prometheus metrics.
func metrics(stop <-chan struct{}) error {
	log.Infoln("Starting Prometheus metrics server")

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	l, err := net.Listen("tcp", *cliPrometheus)
	if err != nil {
		return err
	}

	go func() {
		<-stop
		l.Close()
	}()

	return http.Serve(l, mux)
}

// Starts to stream logs.
func watcher(stop <-chan struct{}) error {
	log.Infoln("Starting directory watcher")

	existing, err := fileutils.List(*cliDirectory)
	if err != nil {
		return err
	}

	for _, file := range existing {
		go func(file os.FileInfo) {
			log.Infof("Starting to stream existing file: %s", file.Name())

			params := awslogs.StreamParams{
				Region:      *cliRegion,
				Directory:   *cliDirectory,
				SkipPattern: *cliIgnore,
				File:        file,
				New:         false,
			}

			err := awslogs.Stream(params)
			if err != nil {
				log.Errorf("Failed to stream existing file: %s: %s", file.Name(), err)
			} else {
				log.Infof("Finished streaming existing file: %s", file.Name())
			}
		}(file)
	}

	created, err := fileutils.Watch(*cliDirectory)
	if err != nil {
		return err
	}

	for {
		file, more := <-created
		if !more {
			break
		}

		go func(file os.FileInfo) {
			log.Infof("Starting to stream new file: %s", file.Name())

			params := awslogs.StreamParams{
				Region:      *cliRegion,
				Directory:   *cliDirectory,
				SkipPattern: *cliIgnore,
				File:        file,
				New:         true,
			}

			err := awslogs.Stream(params)
			if err != nil {
				log.Errorf("Failed to stream new file: %s: %s", file.Name(), err)
			} else {
				log.Infof("Finished streaming new file: %s", file.Name())
			}
		}(file)
	}

	return nil
}

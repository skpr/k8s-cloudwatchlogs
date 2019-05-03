package awslogs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/docker/docker/daemon/logger"
	"github.com/hpcloud/tail"
	"github.com/moby/moby/daemon/logger/awslogs"
	"github.com/prometheus/common/log"

	"github.com/skpr/k8s-cloudwatchlogs/internal/fileutils"
	"github.com/skpr/k8s-cloudwatchlogs/internal/metadata"
)

const (
	// ConfigRegion is required to the "awslogs" Docker client.
	ConfigRegion = "awslogs-region"
	// ConfigCreateGroup is required to the "awslogs" Docker client.
	ConfigCreateGroup = "awslogs-create-group"
	// ConfigGroup is required to the "awslogs" Docker client.
	ConfigGroup = "awslogs-group"
	// ConfigStream is required to the "awslogs" Docker client.
	ConfigStream = "awslogs-stream"
)

// StreamParams passed into the Stream function.
type StreamParams struct {
	Region      string
	Directory   string
	SkipPattern string
	File        os.FileInfo
	New         bool
}

// Stream the contents of a file to CloudWatch Logs.
func Stream(params StreamParams) error {
	namespace, pod, container, err := metadata.Extract(params.File)
	if err != nil {
		return fmt.Errorf("failed to extract namespace, pod and container metadata: %s", err)
	}

	filelogger := log.With("namespace", namespace).With("pod", pod).With("container", container)

	filelogger.Infoln("Starting CloudWatch Logs client")

	// We load the stream backend so we can push these logs to the channel.
	cw, err := awslogs.New(logger.Info{
		Config: map[string]string{
			ConfigRegion:      params.Region,
			ConfigCreateGroup: "true",
			ConfigGroup:       namespace,
			ConfigStream:      fmt.Sprintf("%s-%s", pod, container),
		},
	})
	if err != nil {
		return err
	}
	defer cw.Close()

	filelogger.Infoln("Starting file watcher")

	watcher, err := fileutils.Tail(filepath.Join(params.Directory, params.File.Name()), params.New)
	if err != nil {
		return fmt.Errorf("failed to start tail: %s", err)
	}

	// We also want to monitor to make sure that the file still exists.
	go func(watcher *tail.Tail) {
		limiter := time.Tick(time.Second * 15)

		for {
			<-limiter

			if _, err := os.Stat(watcher.Filename); os.IsNotExist(err) {
				watcher.Stop()
				return
			}
		}
	}(watcher)

	filelogger.Infoln("Starting to stream")

	re := regexp.MustCompile(params.SkipPattern)

	for {
		line, more := <-watcher.Lines
		if !more {
			break
		}

		var message Message

		err := json.Unmarshal([]byte(line.Text), &message)
		if err != nil {
			return fmt.Errorf("failed to unmarshal line for %s/%s/%s: %s", namespace, pod, container, err)
		}

		if re.MatchString(message.Log) {
			continue
		}

		err = cw.Log(&logger.Message{
			Line:      []byte(message.Log),
			Timestamp: message.Time,
		})
		if err != nil {
			return fmt.Errorf("failed to push log: %s", err)
		}
	}

	filelogger.Infoln("Finished streaming")

	return nil
}

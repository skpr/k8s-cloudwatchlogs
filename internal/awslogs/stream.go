package awslogs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/docker/docker/daemon/logger"
	"github.com/heptio/workgroup"
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

	path := filepath.Join(params.Directory, params.File.Name())

	var wg workgroup.Group

	// We also want to monitor to make sure that the file still exists.
	wg.Add(func(stop <-chan struct{}) error {
		filelogger.Infoln("Starting file watcher")

		ticker := time.NewTicker(time.Second * 15)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if _, err := os.Stat(path); os.IsNotExist(err) {
					filelogger.Infoln("File has been deleted from filesystem")
					return nil
				}
			case <-stop:
				return nil
			}
		}
	})

	wg.Add(func(stop <-chan struct{}) error {
		filelogger.Infoln("Starting file tailer")

		regex := regexp.MustCompile(params.SkipPattern)

		watcher, err := fileutils.Tail(path, params.New)
		if err != nil {
			return fmt.Errorf("failed to start tail: %s", err)
		}

		go func() {
			<-stop
			watcher.Stop()
		}()

		for {
			line, more := <-watcher.Lines
			if !more {
				break
			}

			var message Message

			err = json.Unmarshal([]byte(line.Text), &message)
			if err != nil {
				return fmt.Errorf("failed to unmarshal line for %s/%s/%s: %s", namespace, pod, container, err)
			}

			if regex.MatchString(message.Log) {
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

		return nil
	})

	err = wg.Run()
	if err != nil {
		return err
	}

	filelogger.Infoln("Finished streaming")

	return nil
}

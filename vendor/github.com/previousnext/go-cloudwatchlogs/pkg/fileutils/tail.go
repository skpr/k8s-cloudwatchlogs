package fileutils

import (
	"io"

	"github.com/hpcloud/tail"
)

// Tail the contents of a file.
func Tail(file string, start bool) (*tail.Tail, error) {
	config := tail.Config{
		Follow:    true,
		MustExist: true,
	}

	if !start {
		config.Location = &tail.SeekInfo{
			Offset: 0,
			Whence: io.SeekEnd,
		}
	}

	return tail.TailFile(file, config)
}

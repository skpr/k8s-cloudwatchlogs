package cloudwatchlogs

import (
	"os"
	"time"
)

// QueryParams which get passed to the Query function.
type QueryParams struct {
	Group  string
	Prefix string
	Start  int64
	End    int64
}

// QueryOutput which gets returned fromm Query function.
type QueryOutput struct {
	Logs []Log
}

// Log which contains messages from streams.
type Log struct {
	Stream    string
	Timestamp int64
	Message   string
}

// Message is for each line in a Docker container log file.
type Message struct {
	Log    string    `json:"log"`
	Stream string    `json:"stream"`
	Time   time.Time `json:"time"`
}

// StreamParams passed into the Stream function.
type StreamParams struct {
	Region      string
	Directory   string
	Prefix      string
	SkipPattern string
	File        os.FileInfo
	New         bool
}

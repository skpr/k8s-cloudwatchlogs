package awslogs

import "time"

// Message is for each line in a Docker container log file.
type Message struct {
	Log    string    `json:"log"`
	Stream string    `json:"stream"`
	Time   time.Time `json:"time"`
}

# AWS CloudWatch Logs Client

This go package allows read/write operations from CloudWatch Logs.

## Usage

### Read Logs

```go
import (
    "github.com/previousnext/go-cloudwatchlogs"
    "github.com/aws/aws-sdk-go/aws"
)

params := cloudwatchlogs.QueryParams{
    Group:  "loggroup-name",
    Prefix: "logstream-name",
    Start:  aws.TimeUnixMilli(time.Now().Add(-startDuration).UTC()),
    End:    aws.TimeUnixMilli(time.Now().Add(-endDuration).UTC()),
}

resp, err := cloudwatchlogs.Query(client, params)
if err != nil {
    panic(err)
}

for _, log := range resp.Logs {
    fmt.Printf("%s | %s", log.Stream, log.Message)
}
```

### Write Logs

```go
import "github.com/previousnext/go-cloudwatchlogs"

file := os.Open("/var/log/containers/path-to.log")
params := cloudwatchlogs.StreamParams{
    // AWS CloudWatch region. 
    Region:      "us-east-1",
    // Directory where logs are located.
    Directory:   "/var/log/containers",
    // Prefix to apply to loggroup.
    Prefix:      "",
    // Regex pattern to use for excluding certain log messages.
    SkipPattern: "",
    // File handler.
    File:        file,
    // Whether this is a new or existing file.
    New:         false,
}
err := cloudwatchlogs.Stream(params)
if err != nil {
    log.Errorf("Failed to stream existing file: %s: %s", file.Name(), err)
} else {
    log.Infof("Finished streaming existing file: %s", file.Name())
}
```

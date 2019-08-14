package main

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	cliRegion = kingpin.Flag("region", "Region which logs reside").Default("ap-southeast-2").String()
	cliGroup  = kingpin.Flag("group", "CloudWatch Logs group").Required().String()
	cliStream = kingpin.Flag("stream", "CloudWatch Logs stream").String()
	cliStart  = kingpin.Flag("start", "Time ago to search from").Default("10m").String()
	cliEnd    = kingpin.Flag("end", "Time ago to end search").Default("0").String()
)

func main() {
	kingpin.Parse()

	client := cloudwatchlogs.New(session.New(), &aws.Config{Region: cliRegion})

	startDuration, err := time.ParseDuration(*cliStart)
	if err != nil {
		panic(err)
	}

	endDuration, err := time.ParseDuration(*cliEnd)
	if err != nil {
		panic(err)
	}

	params := QueryParams{
		Group:  *cliGroup,
		Prefix: *cliStream,
		Start:  aws.TimeUnixMilli(time.Now().Add(-startDuration).UTC()),
		End:    aws.TimeUnixMilli(time.Now().Add(-endDuration).UTC()),
	}

	resp, err := Query(client, params)
	if err != nil {
		panic(err)
	}

	for _, log := range resp.Logs {
		fmt.Printf("%s | %s", log.Stream, log.Message)
	}
}

type QueryParams struct {
	Group  string
	Prefix string
	Start  int64
	End    int64
}

type QueryOutput struct {
	Logs []Log
}

type Log struct {
	Stream    string
	Timestamp int64
	Message   string
}

func Query(client *cloudwatchlogs.CloudWatchLogs, params QueryParams) (QueryOutput, error) {
	var output QueryOutput

	streams, err := getStreams(client, params.Group, params.Prefix, params.Start)
	if err != nil {
		return output, err
	}

	if len(streams) == 0 {
		return output, nil
	}

	events, err := getLogs(client, params.Group, streams, params.Start, params.End)
	if err != nil {
		return output, err
	}

	for _, event := range events {
		output.Logs = append(output.Logs, Log{
			Stream:    *event.LogStreamName,
			Timestamp: *event.Timestamp,
			Message:   *event.Message,
		})
	}

	return output, nil
}

func getStreams(client *cloudwatchlogs.CloudWatchLogs, group, prefix string, start int64) ([]*string, error) {
	var streams []*string

	params := &cloudwatchlogs.DescribeLogStreamsInput{
		LogGroupName: aws.String(group),
		Descending:   aws.Bool(true),
		OrderBy:      aws.String(cloudwatchlogs.OrderByLastEventTime),
	}

	for {
		resp, err := client.DescribeLogStreams(params)
		if err != nil {
			return streams, err
		}

		for _, stream := range resp.LogStreams {
			if *stream.LastEventTimestamp < start {
				return streams, nil
			}

			streams = append(streams, stream.LogStreamName)
		}

		if resp.NextToken == nil {
			return streams, nil
		}

		params.NextToken = resp.NextToken
	}
}

func getLogs(client *cloudwatchlogs.CloudWatchLogs, group string, streams []*string, start, end int64) ([]*cloudwatchlogs.FilteredLogEvent, error) {
	var events []*cloudwatchlogs.FilteredLogEvent

	params := &cloudwatchlogs.FilterLogEventsInput{
		LogGroupName:   aws.String(group),
		LogStreamNames: streams,
		StartTime:      &start,
		EndTime:        &end,
		Interleaved:    aws.Bool(true),
	}

	for {
		resp, err := client.FilterLogEvents(params)
		if err != nil {
			return events, err
		}

		fmt.Printf("Found %d events\n", len(resp.Events))

		events = append(events, resp.Events...)

		fmt.Println(resp.SearchedLogStreams)

		if resp.NextToken == nil {
			return events, nil
		}

		params.NextToken = resp.NextToken
	}
}

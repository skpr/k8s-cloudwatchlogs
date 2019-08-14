package cloudwatchlogs

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
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

// Query for logs.
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

// Helper function to get streams.
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

// Helper function to get logs.
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

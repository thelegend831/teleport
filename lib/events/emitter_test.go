/*
Copyright 2020 Gravitational, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package events

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"time"

	"gopkg.in/check.v1"

	"github.com/gravitational/teleport/lib/fixtures"
	"github.com/gravitational/teleport/lib/session"
)

// TestProtoStreamer tests edge cases of proto streamer implementation
func (a *EventsTestSuite) TestProtoStreamer(c *check.C) {
	type generateEventsFn func() []AuditEvent
	type testCase struct {
		name           string
		minUploadBytes int64
		events         []AuditEvent
		err            error
		generateEvents generateEventsFn
	}
	testCases := []testCase{
		{
			name:           "5MB similar to S3 min size in bytes",
			minUploadBytes: 1024 * 1024 * 5,
			events:         []AuditEvent{&sessionStart, &sessionPrint, &sessionEnd},
		},
		{
			name:           "get a part per message",
			minUploadBytes: 1,
			events:         []AuditEvent{&sessionStart, &sessionPrint, &sessionEnd},
		},
		{
			name:           "small load test with some uneven numbers",
			minUploadBytes: 1024,
			generateEvents: func() []AuditEvent {
				events := []AuditEvent{&sessionStart}
				i := int64(0)
				for i = 0; i < 1000; i++ {
					event := &SessionPrint{
						Metadata: Metadata{
							Index: int64(i) + 1,
							Type:  SessionPrintEvent,
							Time:  time.Date(2020, 03, 30, 15, 58, 56, 959*int(time.Millisecond), time.UTC),
						},
						ChunkIndex:        int64(i),
						DelayMilliseconds: int64(i),
						Offset:            int64(i),
						Data:              bytes.Repeat([]byte("hello"), int(i%177+1)),
					}
					event.Bytes = int64(len(event.Data))
					event.Time = event.Time.Add(time.Duration(i) * time.Millisecond)
					events = append(events, event)
				}
				i++
				sessionEnd.Metadata.Index = i
				events = append(events, &sessionEnd)
				return events
			},
		},
		{
			name:           "no events",
			minUploadBytes: 1024*1024*5 + 64*1024,
		},
		{
			name:           "one event using the whole part",
			minUploadBytes: 1,
			events:         []AuditEvent{&sessionStart},
		},
	}

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

testcases:
	for i, tc := range testCases {
		uploader := NewMemoryUploader()
		streamer, err := NewProtoStreamer(ProtoStreamerConfig{
			Uploader:       uploader,
			MinUploadBytes: tc.minUploadBytes,
		})
		c.Assert(err, check.IsNil)

		sid := session.ID(fmt.Sprintf("test-%v", i))
		stream, err := streamer.CreateAuditStream(ctx, sid)
		c.Assert(err, check.IsNil)

		events := tc.events
		if tc.generateEvents != nil {
			events = tc.generateEvents()
		}

		for _, event := range events {
			err := stream.EmitAuditEvent(ctx, event)
			if tc.err != nil {
				c.Assert(err, check.FitsTypeOf, tc.err)
				continue testcases
			} else {
				c.Assert(err, check.IsNil)
			}
		}
		err = stream.Complete(ctx)
		c.Assert(err, check.IsNil)

		var outEvents []AuditEvent
		uploads, err := uploader.ListUploads(ctx)
		c.Assert(err, check.IsNil)
		parts, err := uploader.GetParts(uploads[0].ID)
		c.Assert(err, check.IsNil)

		for _, part := range parts {
			reader := NewProtoReader(bytes.NewReader(part))
			out, err := reader.ReadAll(ctx)
			c.Assert(err, check.IsNil, check.Commentf("part crash %#v", part))
			outEvents = append(outEvents, out...)
		}
		fmt.Printf("Test case %v\n", tc.name)
		fixtures.DeepCompareSlices(c, events, outEvents)
	}
}

func min(vars ...int) int {
	m := math.MaxInt64
	for _, zz := range vars {
		if zz < m {
			m = zz
		}
	}
	return m
}

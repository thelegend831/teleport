// +build dynamodb

/*
Copyright 2018 Gravitational, Inc.

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

package s3sessions

import (
	"fmt"
	"testing"

	"github.com/gravitational/teleport/lib/events/test"
	"github.com/gravitational/teleport/lib/utils"

	"gopkg.in/check.v1"
)

func Test(t *testing.T) { check.TestingT(t) }

type S3Suite struct {
	test.HandlerSuite
}

var _ = check.Suite(&S3Suite{})

func (s *S3Suite) SetUpSuite(c *check.C) {
	utils.InitLoggerForTests()

	var err error
	s.HandlerSuite.Handler, err = NewHandler(Config{
		Region: "us-west-1",
		Path:   "/test/",
		Bucket: fmt.Sprintf("teleport-unit-tests"),
	})
	c.Assert(err, check.IsNil)
}

func (s *S3Suite) TestUploadDownload(c *check.C) {
	s.UploadDownload(c)
}

func (s *S3Suite) TestDownloadNotFound(c *check.C) {
	s.DownloadNotFound(c)
}

func (s *S3Suite) TestStream(c *check.C) {
	s.Stream(c)
}

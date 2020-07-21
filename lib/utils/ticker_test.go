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

package utils

import (
	"time"

	"gopkg.in/check.v1"
)

// TickerSuite tests the behavior of custom tickers.
type TickerSuite struct {
}

var _ = check.Suite(&TickerSuite{})

// TestJitterTickerRange verifies that a JitterTicker produces a tick rate
// which hovers between 1x and 2x that of a normal ticker (this test is
// dependent on the properties of the current default Jitter implementation).
func (s *TickerSuite) TestJitterTickerRange(c *check.C) {
	const iterations = 20
	const duration = time.Millisecond * 2
	const maxElapsed = duration * time.Duration(iterations)
	const minElapsed = maxElapsed / 2

	ticker := NewJitterTicker(duration)
	defer ticker.Stop()

	start := time.Now()
	for i := 0; i < iterations; i++ {
		<-ticker.C
	}
	elapsed := time.Since(start)
	c.Assert(elapsed <= maxElapsed, check.Equals, true)
	c.Assert(elapsed <= minElapsed, check.Equals, false)
}

// TestJitterTickerRng verifies that JitterTicker instances are producing
// unique tick-rates over short durations (i.e. that jitter is uniquely
// and correctly applied).
func (s *TickerSuite) TestJitterTickerRng(c *check.C) {
	const duration = time.Millisecond * 2
	const iterations = 10
	const rounds = 10
	const maxSame = iterations / 4

	var elapsed []time.Duration

	for r := 0; r < rounds; r++ {
		ticker := NewJitterTicker(duration)
		defer ticker.Stop()
		start := time.Now()
		for i := 0; i < iterations; i++ {
			<-ticker.C
		}
		elapsed = append(elapsed, time.Since(start))
	}

	// sanity-check to ensure that our deduplication logic
	// behaves correctly.
	elapsed = append(elapsed, elapsed[0])

	// perform dedupe-in-place to determine the number of
	// unique elapsed values.  Correct randomization should
	// result in mostly unique results, event on the
	// extremely short time-scales used in this test.
	unique := elapsed[0:0]
Outer:
	for _, e := range elapsed {
		for _, u := range unique {
			if e == u {
				continue Outer
			}
		}
		unique = append(unique, e)
	}

	// verify that deduplication did at least catch the one duplicate
	// that we deliberately added.
	c.Assert(len(unique) <= iterations, check.Equals, true)

	// if less than half of our test cases resulted in a unique
	// duration, something is almost certainly wrong with our
	// randomization logic.
	c.Assert(len(unique) > iterations/2, check.Equals, true)
}

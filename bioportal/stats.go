// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package main

import (
	"fmt"
	"math"
	"sort"
	"time"
)

type statsMap map[time.Time]stats

type stats struct {
	pos, neg, other int   // number of tests by result
	delays          []int // reporting delays in ascending order
	agePos          map[ageRange]int
}

func (s stats) String() string {
	str := fmt.Sprintf("%4dp %4dn %2do (%0.1f%%)",
		s.pos, s.neg, s.other, 100*float64(s.pos)/float64(s.total()))
	if len(s.delays) > 0 {
		str += fmt.Sprintf(" [%d %d %d %d %d]",
			s.delayPct(0), s.delayPct(25), s.delayPct(50), s.delayPct(75), s.delayPct(100))
	}
	return str
}

func (s *stats) update(res result, ar ageRange, delay int) {
	switch res {
	case positive:
		s.pos++
		if s.agePos == nil {
			s.agePos = make(map[ageRange]int)
		}
		s.agePos[ar]++
	case negative:
		s.neg++
	default:
		s.other++
	}

	if delay >= 0 {
		// Insert delay at correct position.
		i := sort.SearchInts(s.delays, delay)
		s.delays = append(s.delays, 0)
		copy(s.delays[i+1:], s.delays[i:])
		s.delays[i] = delay
	}
}

func (s *stats) total() int {
	return s.pos + s.neg + s.other
}

func (s *stats) delayPct(pct float64) int {
	if len(s.delays) == 0 {
		return 0
	}
	return s.delays[int(math.Round(pct*float64(len(s.delays)-1)/100))]
}

// estInf returns the estimated number of new infections using Youyang Gu's method
// described at https://covid19-projections.com/estimating-true-infections/.
func (s *stats) estInf() int {
	// TODO: Maybe exclude "other" results?
	posRate := float64(s.pos) / float64(s.total())
	return int(math.Round(float64(s.pos) * (16*math.Pow(posRate, 0.5) + 2.5)))
}

// averageStats returns a new map with a numDays-day rolling average for each day in dm.
// The delays and agePos fields are not initialized.
func averageStats(dm statsMap, numDays int) statsMap {
	am := make(statsMap)
	days := sortedTimes(dm)
	for i, d := range days {
		var as stats
		nd := 0 // number of days
		for j := 0; j < numDays && i-j >= 0; j++ {
			s := dm[days[i-j]]
			as.pos += s.pos
			as.neg += s.neg
			as.other += s.other
			nd++
		}
		as.pos /= nd
		as.neg /= nd
		as.other /= nd
		am[d] = as
	}

	return am
}

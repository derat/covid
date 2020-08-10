// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package main

import (
	"fmt"
	"math"
	"time"
)

const maxDelay = 28 // max collect-to-report delay to track in days

type statsMap map[time.Time]*stats

// getStats returns the stats object for t, creating it if necessary.
func (m statsMap) get(t time.Time) *stats {
	if s, ok := m[t]; ok {
		return s
	}
	s := newStats()
	m[t] = s
	return s
}

type stats struct {
	pos, neg, other int   // number of tests by result
	delayCounts     []int // counts of tests indexed by collect-to-report delay in days
	numDelays       int   // total number of tests in delayCounts
	agePos          map[ageRange]int
}

func newStats() *stats {
	return &stats{
		delayCounts: make([]int, maxDelay+2), // extra for 0 and for overflow
		agePos:      make(map[ageRange]int),
	}
}

func (s stats) String() string {
	str := fmt.Sprintf("%3d %4d %2d %4.1f%%",
		s.pos, s.neg, s.other, 100*float64(s.pos)/float64(s.total()))
	str += fmt.Sprintf(" [%d %d %d %d %d]",
		s.delayPct(0), s.delayPct(25), s.delayPct(50), s.delayPct(75), s.delayPct(100))
	return str
}

// update incorporates a single test into s.
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
		idx := delay
		if idx >= len(s.delayCounts) {
			idx = len(s.delayCounts) - 1
		}
		s.delayCounts[idx]++
		s.numDelays++
	}
}

func (s *stats) total() int {
	return s.pos + s.neg + s.other
}

func (s *stats) delayPct(pct float64) int {
	if s.numDelays == 0 || pct < 0 || pct > 100 {
		return 0
	}

	total := 0
	target := 1 + int(math.Round(pct*float64(s.numDelays-1)/100))
	for d := 0; d < len(s.delayCounts); d++ {
		if total += s.delayCounts[d]; total >= target {
			return d
		}
	}
	panic("didn't find delay") // shouldn't be reached
}

// estInf returns the estimated number of new infections using Youyang Gu's method
// described at https://covid19-projections.com/estimating-true-infections/.
func (s *stats) estInf() int {
	// TODO: Maybe exclude "other" results?
	posRate := float64(s.pos) / float64(s.total())
	return int(math.Round(float64(s.pos) * (16*math.Pow(posRate, 0.5) + 2.5)))
}

// add incorporates o into s.
func (s *stats) add(o *stats) {
	s.pos += o.pos
	s.neg += o.neg
	s.other += o.other

	for i := 0; i < len(s.delayCounts); i++ {
		s.delayCounts[i] += o.delayCounts[i]
	}
	s.numDelays += o.numDelays

	for ar := ageMin; ar <= ageMax; ar++ {
		s.agePos[ar] += o.agePos[ar]
	}
}

// weeklyStats aggregates the stats in dm by week (starting on Sundays).
func weeklyStats(dm statsMap) statsMap {
	wm := make(statsMap)
	for d, s := range dm {
		week := d.AddDate(0, 0, -1*int(d.Weekday())) // subtract to sunday
		ws := wm[week]
		if ws == nil {
			ws = newStats()
			wm[week] = ws
		}
		ws.add(s)
	}
	return wm
}

// averageStats returns a new map with a numDays-day rolling average for each day in dm.
func averageStats(dm statsMap, numDays int) statsMap {
	am := make(statsMap)
	days := sortedTimes(dm)
	for i, d := range days {
		as := newStats()
		nd := 0 // number of days
		for j := 0; j < numDays && i-j >= 0; j++ {
			as.add(dm[days[i-j]])
			nd++
		}
		as.pos /= nd
		as.neg /= nd
		as.other /= nd
		// TODO: Update delays and age-positives.
		am[d] = as
	}
	return am
}

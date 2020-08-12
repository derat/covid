// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package main

import (
	"fmt"
	"math"
	"time"
)

const maxDelay = 28 // max collect-to-report delay to track in days

type stats struct {
	pos, neg, other int // number of molecular tests by result
	ab, ag, unk     int // number of serological, antigen, and unknown tests

	agePos map[ageRange]int // positive molecular results grouped by patient age

	delays, posDelays, negDelays *hist // delays for total, positive, and negative molecular results
}

func newStats() *stats {
	return &stats{
		agePos:    make(map[ageRange]int),
		delays:    newHist(maxDelay),
		posDelays: newHist(maxDelay),
		negDelays: newHist(maxDelay),
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
func (s *stats) update(t testType, res result, ar ageRange, delay int) {
	switch t {
	case molecular:
		switch res {
		case positive:
			s.pos++
			s.agePos[ar]++
			s.posDelays.inc(delay)
		case negative:
			s.neg++
			s.negDelays.inc(delay)
		default:
			s.other++
		}
		s.delays.inc(delay)
	case serological:
		s.ab++
	case antigen:
		s.ag++
	case unknownType:
		s.unk++
	}
}

func (s *stats) total() int {
	return s.pos + s.neg + s.other
}

func (s *stats) delayPct(pct float64) int {
	return s.delays.percentile(pct)
}

func (s *stats) posDelayPct(pct float64) int {
	return s.posDelays.percentile(pct)
}

func (s *stats) negDelayPct(pct float64) int {
	return s.negDelays.percentile(pct)
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
	if o == nil {
		return
	}
	s.pos += o.pos
	s.neg += o.neg
	s.other += o.other
	s.ab += o.ab
	s.ag += o.ag
	s.unk += o.unk
	s.delays.add(o.delays)
	s.posDelays.add(o.posDelays)
	s.negDelays.add(o.negDelays)
	for ar := ageMin; ar <= ageMax; ar++ {
		s.agePos[ar] += o.agePos[ar]
	}
}

// scale multiplies s's values by sc.
func (s *stats) scale(sc float64) {
	rs := func(v int) int { return int(math.Round(sc * float64(v))) }
	s.pos = rs(s.pos)
	s.neg = rs(s.neg)
	s.other = rs(s.other)
	s.ab = rs(s.ab)
	s.ag = rs(s.ag)
	s.unk = rs(s.unk)
	s.delays.scale(sc)
	s.posDelays.scale(sc)
	s.negDelays.scale(sc)
	for ar := ageMin; ar <= ageMax; ar++ {
		s.agePos[ar] = rs(s.agePos[ar])
	}
}

type hist struct {
	counts []int // bucketed counts
	total  int   // total number of tests in counts
}

func newHist(max int) *hist {
	return &hist{counts: make([]int, max+2)} // extra for 0 and for overflow
}

// inc increments the histogram for the supplied value. Negative values are ignored.
func (h *hist) inc(v int) {
	if v < 0 {
		return
	}

	i := v
	if i >= len(h.counts) {
		i = len(h.counts) - 1
	}
	h.counts[i]++
	h.total++
}

func (h *hist) percentile(p float64) int {
	if h.total == 0 || p < 0 || p > 100 {
		return 0
	}

	seen := 0
	target := 1 + int(math.Round(p*float64(h.total-1)/100))
	for i := 0; i < len(h.counts); i++ {
		if seen += h.counts[i]; seen >= target {
			return i
		}
	}
	panic("didn't find value for percentile") // shouldn't be reached
}

func (h *hist) add(o *hist) {
	if len(h.counts) != len(o.counts) {
		panic(fmt.Sprintf("can't add histograms with %v and %v bucket(s)", len(h.counts), len(o.counts)))
	}
	for i := 0; i < len(h.counts); i++ {
		h.counts[i] += o.counts[i]
	}
	h.total += o.total
}

// scale scales h's counts by sc.
func (h *hist) scale(sc float64) {
	h.total = 0
	for i := range h.counts {
		h.counts[i] = int(math.Round(sc * float64(h.counts[i])))
		h.total += h.counts[i]
	}
}

// statsMap holds stats indexed by time (typically days).
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
		as := am.get(d)
		nd := 0
		for j := 0; j < numDays && i-j >= 0; j++ {
			as.add(dm[days[i-j]])
			nd++
		}
		as.scale(1 / float64(nd))
	}
	return am
}

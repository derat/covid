package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"sort"
	"time"
)

const (
	filename = "minimal-info-unique-tests"
)

var (
	loc       *time.Location // PR time zone
	startDate time.Time      // earliest date to accept
)

func init() {
	var err error
	if loc, err = time.LoadLocation("America/Puerto_Rico"); err != nil {
		panic(err)
	}
	startDate = time.Date(2020, 2, 1, 0, 0, 0, 0, loc)
}

func main() {
	f, err := os.Open(filename)
	if err != nil {
		log.Fatal("Failed opening file: ", err)
	}
	defer f.Close()

	var tests []test
	if err := json.NewDecoder(f).Decode(&tests); err != nil {
		log.Fatal("Failed decoding data: ", err)
	}

	now := time.Now()
	colStats := make(map[time.Time]stats)
	repStats := make(map[time.Time]stats)

	for _, t := range tests {
		col := time.Time(t.Collected)
		colValid := !col.Before(startDate) && !col.After(now)

		rep := time.Time(t.Reported)
		repValid := !rep.Before(startDate) && !rep.After(now)

		delay := -1
		if colValid && repValid && !col.After(rep) {
			delay = int(math.Round(float64(rep.Sub(col)) / float64(24*time.Hour)))
		}

		if colValid {
			s := colStats[col]
			s.update(t.Result, delay)
			colStats[col] = s
		}

		if repValid {
			s := repStats[rep]
			s.update(t.Result, delay)
			repStats[rep] = s
		}
	}

	printStats(repStats)
}

type stats struct {
	pos, neg, other int   // number of tests by result
	delays          []int // reporting delays in ascending order
}

func (s *stats) update(res result, delay int) {
	switch res {
	case positive:
		s.pos++
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

func printStats(m map[time.Time]stats) {
	days := make([]time.Time, 0, len(m))
	for d := range m {
		days = append(days, d)
	}
	sort.Slice(days, func(i, j int) bool { return days[i].Before(days[j]) })

	for i, d := range days {
		s := m[d]

		// Sum the test results over the past week.
		var ws stats
		for j := int(math.Max(float64(i-6), 0)); j <= i; j++ {
			ds := m[days[j]]
			ws.pos += ds.pos
			ws.neg += ds.neg
			ws.other += ds.other
		}

		fmt.Printf("%s: %4d %4.1f%% [%d %d %d]\n", d.Format("2006-01-02"), s.total(),
			100*float64(ws.pos)/float64(ws.total()), s.delayPct(20), s.delayPct(50), s.delayPct(80))
	}
}

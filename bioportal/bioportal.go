package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"reflect"
	"sort"
	"time"
)

const (
	inputFile  = "minimal-info-unique-tests"
	posAgeFile = "positives-age.data"
	delaysFile = "delays.data"
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
	f, err := os.Open(inputFile)
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
			s.update(t.Result, t.AgeRange, delay)
			colStats[col] = s
		}

		if repValid {
			s := repStats[rep]
			s.update(t.Result, t.AgeRange, delay)
			repStats[rep] = s
		}
	}

	//printStats(repStats)
	writeAgeData(posAgeFile, repStats)
	writeDelaysData(delaysFile, repStats)
}

type stats struct {
	pos, neg, other int   // number of tests by result
	delays          []int // reporting delays in ascending order
	agePos          map[ageRange]int
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

func printStats(m map[time.Time]stats) {
	days := sortedTimes(m)

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

func writeAgeData(p string, m map[time.Time]stats) error {
	fw, err := newFileWriter(p)
	if err != nil {
		return err
	}

	// Aggregate positives by week.
	wm := make(map[time.Time]map[ageRange]int)
	for d, s := range m {
		wd := d.AddDate(0, 0, -1*int(d.Weekday())) // subtract to sunday
		am := wm[wd]
		if am == nil {
			am = make(map[ageRange]int)
		}
		for ar := age0To9; ar <= age100To109; ar++ {
			am[ar] += s.agePos[ar]
		}
		wm[wd] = am
	}

	fw.printf("X\tDate\tAge\tPositive Tests\n")
	for i, d := range sortedTimes(wm) {
		am := wm[d]
		for ar := age0To9; ar <= age100To109; ar++ {
			fw.printf("%d\t%s\t%d\t%d\n", i, d.Format("01/02"), ar.min(), am[ar])
		}
	}

	return fw.close()
}

func writeDelaysData(p string, m map[time.Time]stats) error {
	fw, err := newFileWriter(p)
	if err != nil {
		return err
	}

	wm := make(map[time.Time]stats)
	for d, s := range m {
		wd := d.AddDate(0, 0, -1*int(d.Weekday())) // subtract to sunday
		ws := wm[wd]
		ws.delays = append(ws.delays, s.delays...)
		wm[wd] = ws
	}

	fw.printf("Date\t20th\t50th\t80th\n")
	for _, d := range sortedTimes(wm) {
		s := wm[d]
		sort.Ints(s.delays)
		fw.printf("%s\t%d\t%d\t%d\n", d.Format("2006-01-02"),
			s.delayPct(20), s.delayPct(50), s.delayPct(80))
	}
	return fw.close()
}

// sortedTimes returns sorted keys from m, which must be a map with time.Time keys.
// See https://stackoverflow.com/a/35366762.
func sortedTimes(m interface{}) []time.Time {
	var keys []time.Time
	for _, k := range reflect.ValueOf(m).MapKeys() {
		keys = append(keys, k.Interface().(time.Time))
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i].Before(keys[j]) })
	return keys
}

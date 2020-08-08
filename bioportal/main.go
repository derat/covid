// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package main

import (
	"compress/gzip"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"time"
)

const (
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
	startDate = time.Date(2020, 3, 12, 0, 0, 0, 0, loc)
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %v <input>\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if len(flag.Args()) != 1 {
		flag.Usage()
		os.Exit(2)
	}

	fn := flag.Arg(0)
	f, err := os.Open(fn)
	if err != nil {
		log.Fatal("Failed opening input: ", err)
	}
	defer f.Close()

	var r io.Reader = f
	if filepath.Ext(fn) == ".gz" {
		gr, err := gzip.NewReader(f)
		if err != nil {
			log.Fatal("Failed decompressing input: ", err)
		}
		defer gr.Close()
		r = gr
	}

	now := time.Now()
	colStats := make(statsMap)
	repStats := make(statsMap)

	// Instead of unmarshaling all tests into slice all at once, strip off the
	// opening bracket so we can read them one at a time. See the "Stream"
	// example at https://golang.org/pkg/encoding/json/#Decoder.Decode.
	dec := json.NewDecoder(r)
	if t, err := dec.Token(); err != nil {
		log.Fatal("Failed reading opening bracket: ", err)
	} else if d, ok := t.(json.Delim); !ok || d != '[' {
		log.Fatalf("Data starts with %v instead of opening bracket", t)
	}

	for dec.More() {
		var t test
		if err := dec.Decode(&t); err != nil {
			log.Fatal("Failed reading test: ", err)
		}

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

	if t, err := dec.Token(); err != nil {
		log.Fatal("Failed reading closing bracket: ", err)
	} else if d, ok := t.(json.Delim); !ok || d != ']' {
		log.Fatalf("Data ends with %v instead of closing bracket", t)
	}

	for _, d := range sortedTimes(repStats) {
		fmt.Printf("%s: %s\n", d.Format("2006-01-02"), repStats[d])
	}
	writeAgeData(posAgeFile, repStats)
	writeDelaysData(delaysFile, repStats)
}

func writeAgeData(p string, m statsMap) error {
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

func writeDelaysData(p string, m statsMap) error {
	fw, err := newFileWriter(p)
	if err != nil {
		return err
	}

	wm := make(statsMap)
	for d, s := range m {
		wd := d.AddDate(0, 0, -1*int(d.Weekday())) // subtract to sunday
		ws := wm[wd]
		ws.delays = append(ws.delays, s.delays...)
		wm[wd] = ws
	}

	fw.printf("Date\t25th\t50th\t75th\n")
	for _, d := range sortedTimes(wm) {
		s := wm[d]
		sort.Ints(s.delays)
		fw.printf("%s\t%d\t%d\t%d\n", d.Format("2006-01-02"),
			s.delayPct(25), s.delayPct(50), s.delayPct(75))
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

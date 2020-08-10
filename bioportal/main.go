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

	"github.com/derat/covid/filewriter"
	"github.com/derat/covid/gnuplot"
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
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %v <input> [out-dir]\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if ln := len(flag.Args()); ln == 0 || ln > 2 {
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

	_, repStats, err := readTests(r)
	if err != nil {
		log.Fatal("Failed reading tests: ", err)
	}

	// If an output dir wasn't supplied, just print a summary.
	if len(flag.Args()) < 2 {
		for _, d := range sortedTimes(repStats) {
			fmt.Printf("%s: %s\n", d.Format("2006-01-02"), repStats[d])
		}
		return
	}

	outDir := flag.Arg(1)
	if err := os.MkdirAll(outDir, 0755); err != nil {
		log.Fatal("Failed creating output dir: ", err)
	}

	avgRepStats := averageStats(repStats, 7)
	weekRepStats := weeklyStats(repStats)

	for _, plot := range []struct {
		out  string                         // output file, e.g. "my-plot.png"
		tmpl string                         // gnuplot template data
		data func(w *filewriter.FileWriter) // writes gnuplot data to w
	}{
		{
			out:  "positives-age.png",
			tmpl: posAgeTmpl,
			data: func(w *filewriter.FileWriter) {
				w.Printf("X\tDate\tAge\tPositive Tests\n")
				for i, week := range sortedTimes(weekRepStats) {
					s := weekRepStats[week]
					for ar := age0To9; ar <= age100To109; ar++ {
						w.Printf("%d\t%s\t%d\t%d\n", i, week.Format("01/02"), ar.min(), s.agePos[ar])
					}
				}
			},
		},
		{
			out:  "reports-daily.png",
			tmpl: reportsTmpl,
			data: func(w *filewriter.FileWriter) {
				w.Printf("Date\tResults\n")
				for _, d := range sortedTimes(avgRepStats) {
					w.Printf("%s\t%d\n", d.Format("2006-01-02"), avgRepStats[d].total())
				}
			},
		},
		{
			out:  "result-delays.png",
			tmpl: delaysTmpl,
			data: func(w *filewriter.FileWriter) {
				w.Printf("Date\t25th\t50th\t75th\n")
				for _, week := range sortedTimes(weekRepStats) {
					s := weekRepStats[week]
					w.Printf("%s\t%d\t%d\t%d\n", week.Format("2006-01-02"),
						s.delayPct(25), s.delayPct(50), s.delayPct(75))
				}
			},
		},
	} {
		dp := filepath.Join("/tmp", "bioportal."+plot.out+".dat")
		dw := filewriter.New(dp)
		plot.data(dw)
		if err := dw.Close(); err != nil {
			log.Fatalf("Failed writing data for %v: %v", plot.out, err)
		}
		if err := gnuplot.ExecTemplate(plot.tmpl, templateData(dp, filepath.Join(outDir, plot.out), time.Now())); err != nil {
			log.Fatalf("Failed plotting %v: %v", plot.out, err)
		}
		os.Remove(dp)
	}
}

// readTests reads a JSON array of test objects from r and returns daily stats
// aggregated by collection date and by reporting date.
func readTests(r io.Reader) (colStats, repStats statsMap, err error) {
	// Instead of unmarshaling all tests into slice all at once, strip off the
	// opening bracket so we can read them one at a time. See the "Stream"
	// example at https://golang.org/pkg/encoding/json/#Decoder.Decode.
	dec := json.NewDecoder(r)
	if t, err := dec.Token(); err != nil {
		return nil, nil, fmt.Errorf("failed reading opening bracket: ", err)
	} else if d, ok := t.(json.Delim); !ok || d != '[' {
		return nil, nil, fmt.Errorf("data starts with %v instead of opening bracket", t)
	}

	now := time.Now()
	colStats = make(statsMap)
	repStats = make(statsMap)

	for dec.More() {
		var t test
		if err := dec.Decode(&t); err != nil {
			return nil, nil, fmt.Errorf("failed reading test: ", err)
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
			colStats.get(col).update(t.Result, t.AgeRange, delay)
		}
		if repValid {
			repStats.get(rep).update(t.Result, t.AgeRange, delay)
		}
	}

	if t, err := dec.Token(); err != nil {
		return nil, nil, fmt.Errorf("failed reading closing bracket: ", err)
	} else if d, ok := t.(json.Delim); !ok || d != ']' {
		return nil, nil, fmt.Errorf("data ends with %v instead of closing bracket", t)
	}
	return colStats, repStats, nil
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

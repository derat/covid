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

const (
	repoURL     = "https://github.com/derat/covid"
	fontName    = "Roboto"
	fontSize    = 22
	imageWidth  = 1280
	imageHeight = 960
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

	for _, plot := range []struct {
		out  string                         // output file, e.g. "my-plot.png"
		tmpl string                         // gnuplot template data
		data func(w *filewriter.FileWriter) // writes gnuplot data to w
	}{
		{
			out:  "positives-age.png",
			tmpl: posAgeTmpl,
			data: func(w *filewriter.FileWriter) {
				// Aggregate positives by week.
				m := make(map[time.Time]map[ageRange]int)
				for d, s := range repStats {
					week := d.AddDate(0, 0, -1*int(d.Weekday())) // subtract to sunday
					am := m[week]
					if am == nil {
						am = make(map[ageRange]int)
					}
					for ar := age0To9; ar <= age100To109; ar++ {
						am[ar] += s.agePos[ar]
					}
					m[week] = am
				}
				w.Printf("X\tDate\tAge\tPositive Tests\n")
				for i, week := range sortedTimes(m) {
					am := m[week]
					for ar := age0To9; ar <= age100To109; ar++ {
						w.Printf("%d\t%s\t%d\t%d\n", i, week.Format("01/02"), ar.min(), am[ar])
					}
				}
			},
		},
		{
			out:  "result-delays.png",
			tmpl: delaysTmpl,
			data: func(w *filewriter.FileWriter) {
				m := make(statsMap)
				for d, s := range repStats {
					week := d.AddDate(0, 0, -1*int(d.Weekday())) // subtract to sunday
					ws := m[week]
					ws.delays = append(ws.delays, s.delays...)
					m[week] = ws
				}
				w.Printf("Date\t25th\t50th\t75th\n")
				for _, week := range sortedTimes(m) {
					s := m[week]
					sort.Ints(s.delays)
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
		if err := gnuplot.ExecTemplate(plot.tmpl, struct{ DataPath, SetTerm, SetOutput, FooterLabel string }{
			DataPath:  dp,
			SetTerm:   fmt.Sprintf("set term pngcairo font '%s,%d' size %d,%d", fontName, fontSize, imageWidth, imageHeight),
			SetOutput: fmt.Sprintf("set output '%s'", filepath.Join(outDir, plot.out)),
			FooterLabel: fmt.Sprintf("set label front '{/*0.7 Generated on %s by %s}' at screen 0.99,0.025 right",
				time.Now().Format("2006-01-02"), repoURL),
		}); err != nil {
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

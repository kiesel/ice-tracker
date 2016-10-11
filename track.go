package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"text/tabwriter"
	"time"
)

const ENDPOINT = "http://ice.portal/api1/rs/status"

type JsonRecord struct {
	Connection   bool
	ServiceLevel string
	Speed        float32
	Longitude    float32
	Latitude     float32
	ServerTime   int64
	WagonClass   int
}

func (this *JsonRecord) String() string {
	return fmt.Sprintf("[%0.2f km/h @ {%0.5f,%0.5f}, %s]\n",
		this.Speed,
		this.Longitude,
		this.Latitude,
		time.Unix(this.ServerTime/1000, 0))
}

func (this *JsonRecord) ToSlice() []string {
	return []string{
		fmt.Sprintf("%0.3f", this.Speed),
		fmt.Sprintf("%0.5f", this.Longitude),
		fmt.Sprintf("%0.5f", this.Latitude),
		fmt.Sprintf("%d", this.ServerTime),
		fmt.Sprintf("%s", time.Unix(this.ServerTime/1000, 0)),
	}
}

func (this *JsonRecord) Time() time.Time {
	return time.Unix(this.ServerTime/1000, 0)
}

var filename = "ice-tracker.csv"
var frequency int = 5
var line = 0
var printer = new(tabwriter.Writer)
var lastSpeed = (float32)(0.0)
var maxSpeed = (float32)(0.0)

func init() {
	flag.StringVar(&filename, "filename", "ice-tracker.csv", "Track to this file")
	flag.IntVar(&frequency, "frequency", 5, "Retrieve values every n seconds")
	printer.Init(os.Stdout, 0, 8, 2, '\t', 0)
}

func main() {
	flag.Parse()

	for {
		next := time.Now().Add(time.Duration(frequency) * time.Second)

		rec, err := fetchRecord()
		if err != nil {
			fmt.Println(err)
			time.Sleep(1 * time.Second)
			continue
		}

		printRecord(*rec)
		storeRecord(*rec)

		// Try to achieve exact frequency timing
		time.Sleep(next.Sub(time.Now()))
	}
}

func fetchRecord() (*JsonRecord, error) {
	resp, err := http.Get(ENDPOINT)
	if err != nil {
		return nil, err
	}

	var rec JsonRecord
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&rec); err != nil {
		return nil, err
	}

	return &rec, nil
}

func printRecord(record JsonRecord) {
	if line%20 == 0 {
		fmt.Fprintf(printer, "TIME\tSPEED\tLONGITUDE\tLATITUDE\tTREND\n")
	}

	trend := "="
	if lastSpeed > record.Speed {
		trend = "v"
	} else if lastSpeed < record.Speed {
		trend = "^"
	}
	lastSpeed = record.Speed

	info := ""
	if record.Speed > maxSpeed {
		maxSpeed = record.Speed
		info = "RECORD!"
	}

	fmt.Fprintf(printer, "%s\t%0.1f km/h\t%0.3f\t%0.3f\t%s\t%s\n",
		record.Time(),
		record.Speed,
		record.Longitude,
		record.Latitude,
		trend,
		info,
	)

	printer.Flush()

	line++
}

func storeRecord(record JsonRecord) {
	f, err := os.OpenFile("ice-tracker.csv", os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)

	if err != nil {
		fmt.Print(err.Error)
		return
	}

	defer f.Close()
	w := csv.NewWriter(f)
	if err := w.Write(record.ToSlice()); err != nil {
		fmt.Print(err)
	}

	w.Flush()

	if err := w.Error(); err != nil {
		fmt.Println(err)
	}
}

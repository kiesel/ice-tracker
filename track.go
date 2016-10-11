package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
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

var filename = "ice-tracker.csv"
var frequency int = 5

func init() {
	flag.StringVar(&filename, "filename", "ice-tracker.csv", "Track to this file")
	flag.IntVar(&frequency, "frequency", 5, "Retrieve values every n seconds")
}

func main() {
	flag.Parse()

	maxSpeed := ((float32)(0.0))

	for {
		next := time.Now().Add(time.Duration(frequency) * time.Second)

		rec, err := fetchRecord()
		if err != nil {
			fmt.Println(err)
			time.Sleep(1 * time.Second)
			continue
		}

		fmt.Print(rec.String())
		storeRecord(*rec)

		if rec.Speed > maxSpeed {
			maxSpeed = rec.Speed
			fmt.Println("NEW RECORD!!!")
		}

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

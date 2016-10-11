package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
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

func main() {
	maxSpeed := ((float32)(0.0))

	for {
		resp, err := http.Get(ENDPOINT)
		if err != nil {
			panic(err.Error())
		}

		decoder := json.NewDecoder(resp.Body)
		var rec JsonRecord
		if err := decoder.Decode(&rec); err == io.EOF {
			break
		} else if err != nil {
			panic(err.Error())
		}

		fmt.Print(rec.String())
		storeRecord(rec)

		if rec.Speed > maxSpeed {
			maxSpeed = rec.Speed
			fmt.Println("NEW RECORD!!!")
		}

		time.Sleep(5 * time.Second)
	}
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

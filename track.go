package main 

import (
  "net/http"
  "fmt"
  "io"
  "encoding/json"
  "time"
)

const ENDPOINT = "http://ice.portal/api1/rs/status"

type JsonRecord struct {
  Connection bool 
  ServiceLevel string
  Speed float32 
  Longitude float32
  Latitude float32
  ServerTime int64
  WagonClass int
}

func (this *JsonRecord) String() string {
  return fmt.Sprintf("[%0.2f km/h @ {%0.5f,%0.5f}, time=%s]\n",
    this.Speed,
    this.Longitude,
    this.Latitude,
    time.Unix(this.ServerTime / 1000, 0))
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

    if rec.Speed > maxSpeed {
      maxSpeed = rec.Speed
      fmt.Println("NEW RECORD!!!")
    }

    time.Sleep(5 * time.Second)
  }
}
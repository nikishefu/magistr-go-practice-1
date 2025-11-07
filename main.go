package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type AvailabilityLogger struct {
	unavailableCount uint8
}

func (al *AvailabilityLogger) Unavailable() {
	if al.unavailableCount == 3 {
		fmt.Println("Unable to fetch server statistic")
	} else {
		al.unavailableCount++
	}
}

func (al AvailabilityLogger) Reset() {
	al.unavailableCount = 0
}

func CheckStats(statsStr string, al *AvailabilityLogger) {
	stats := make([]int64, 7)
	for i, stat := range strings.Split(string(statsStr), ",") {
		statInt, err := strconv.ParseInt(stat, 10, 64)
		if err != nil {
			al.Unavailable()
			continue
		}
		stats[i] = statInt
	}

	if stats[0] > 30 {
		fmt.Println("Load Average is too high:", stats[0])
	}
	if mem := float64(stats[2]) * 100 / float64(stats[1]); mem > 80 {
		fmt.Println("Memory usage too high:", mem)
	}
	if disk := float64(stats[4]) * 100 / float64(stats[3]); disk > 90 {
		mbLeft := (stats[4] - stats[3]) >> 20
		fmt.Printf("Free disk space is too low: %v Mb left", mbLeft)
	}
	if network := float64(stats[6]) * 100 / float64(stats[5]); network > 90 {
		bwLeft := (stats[4] - stats[3]) >> 17
		fmt.Println("Network bandwidth usage high: %v Mbit/s available", bwLeft)
	}
}

func main() {
	al := AvailabilityLogger{}
	for {
		time.Sleep(time.Second)

		r, err := http.Get("https://srv.msk01.gigacorp.local/_stats")
		if err != nil {
			al.Unavailable()
			continue
		}

		contType := r.Header.Get("Content-Type")
		if r.StatusCode != 200 || contType != "text/plain; charset=UTF-8" {
			al.Unavailable()
			continue
		}

		body, err := io.ReadAll(r.Body)
		r.Body.Close()
		if err != nil {
			al.Unavailable()
			continue
		}

		CheckStats(string(body), &al)
		al.Reset()
	}
}

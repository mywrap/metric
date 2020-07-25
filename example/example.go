package main

import (
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/mywrap/metric"
)

func main() {
	keys := map[string]string{
		"golang.org":    "https://golang.org/",
		"github.com":    "https://github.com/",
		"facebook.com":  "https://www.facebook.com/",
		"vnexpress.net": "https://vnexpress.net/",
	}
	keysList := make([]string, 0)
	for k, _ := range keys {
		keysList = append(keysList, k)
	}
	nKeys := len(keysList)
	rand.Seed(time.Now().UnixNano())

	m := metric.NewMemoryMetric()
	wg := &sync.WaitGroup{}
	for i := 0; i < 1000; i++ {
		if i%20 == 0 {
			log.Println("i:", i)
		}
		key := keysList[rand.Intn(nKeys)]
		i := i
		wg.Add(1)
		go func() {
			defer wg.Add(-1)
			beginT := time.Now()
			resp, err := http.Get(keys[key])
			m.Count(key)
			m.Duration(key, time.Since(beginT))
			if err != nil {
				log.Printf("err: %v, key: %v, i: %v\n", err, key, i)
				return
			}
			resp.Body.Close()
		}()
		time.Sleep(20 * time.Millisecond)
	}
	wg.Wait()
	rows := m.GetCurrentMetric()
	for _, row := range rows {
		log.Println(row)
	}
}

/*
key: facebook.com, count: 229, aveSecs: 0.294, percentiles:
	metric.PG1{P0:0.255, P25:0.278, P50:0.286, P75:0.297, P100:0.683},
	metric.PG2{P90:0.313, P95:0.334, P99:0.54, P995:0.663, P999:0.683}
key: github.com, count: 245, aveSecs: 0.163, percentiles:
	metric.PG1{P0:0.124, P25:0.142, P50:0.16, P75:0.177, P100:0.294},
	metric.PG2{P90:0.197, P95:0.21, P99:0.258, P995:0.291, P999:0.294}
key: golang.org, count: 273, aveSecs: 0.198, percentiles:
	metric.PG1{P0:0.189, P25:0.191, P50:0.193, P75:0.195, P100:0.693},
	metric.PG2{P90:0.198, P95:0.202, P99:0.359, P995:0.399, P999:0.693}
key: vnexpress.net, count: 253, aveSecs: 0.027, percentiles:
	metric.PG1{P0:0.022, P25:0.024, P50:0.024, P75:0.027, P100:0.137},
	metric.PG2{P90:0.03, P95:0.034, P99:0.109, P995:0.13, P999:0.137}
*/

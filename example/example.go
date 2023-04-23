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
	client := http.Client{Timeout: 10 * time.Second}
	m := metric.NewMemoryMetric()
	wg := &sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		if i%20 == 0 {
			log.Println("i:", i)
		}
		key := keysList[rand.Intn(nKeys)]
		i := i
		wg.Add(1)
		go func() {
			defer wg.Add(-1)
			beginT := time.Now()
			resp, err := client.Get(keys[key])
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

/* Output:
facebook.com, count: 24, aveSecs: 0.345, percentiles:
	{P0:0.301 P25:0.305 P50:0.31 P75:0.334 P90:0.482 P99:0.572 P100:0.572}
github.com, count: 27, aveSecs: 0.101, percentiles:
	{P0:0.058 P25:0.059 P50:0.06 P75:0.122 P90:0.205 P99:0.42 P100:0.42}
golang.org, count: 21, aveSecs: 0.639, percentiles:
	{P0:0.558 P25:0.585 P50:0.596 P75:0.661 P90:0.804 P99:0.887 P100:0.887}
vnexpress.net, count: 28, aveSecs: 0.092, percentiles:
	{P0:0.051 P25:0.053 P50:0.056 P75:0.06 P90:0.21 P99:0.361 P100:0.361}
*/

# metric

Package metric is used for observing request count and duration.  

It use an order statistic tree to store durations, so it can calculate
any duration percentile in time complexity log(n).

Wrapped [Go order statistic tree](https://github.com/daominah/GoLLRB).

## Usage

* Example 1: using MemoryMetric to measure http handlers duration:

  ````go
  metric := NewMemoryMetric()
  func EmbedHandler(method string, path string, handler http.HandlerFunc) (
  http.HandlerFunc) {
      metricKey := fmt.Sprintf("%v_%v", path, method)
      return func(w http.ResponseWriter, r *http.Request) {
          metric.Count(metricKey)
          beginTime := time.Now()
          handler(w, r)
          metric.Duration(metricKey, time.Since(beginTime))
      }
  }
  ````
  Detail in [httpsvr.go](
  https://github.com/mywrap/httpsvr/blob/master/httpsvr.go).

* Example 2: observing requests to websites [example.go](
  ./example/example.go):
  
  ````go
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
  ````

## Config

Default MemoryMetric will never auto reset. You can create a cron to
periodically reset the metric, example once per day.

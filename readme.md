# metric

Package metric is used for observing request count and duration.  

It use an order statistic tree to store durations, so it can calculate
any duration percentile in time complexity log(n).

Wrapped [Go order statistic tree](https://github.com/daominah/GoLLRB).

## Usage

* Example 1: using MemoryMetric to measure http handlers duration:

  ````go
  metric := NewMemoryMetric()
  func HandlerWithMetric(method string, path string, handler http.HandlerFunc) http.HandlerFunc {
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

* Example 2: observing requests to websitesin [example.go](example/example.go):
  
  ````go
  facebook.com, count: 24, aveSecs: 0.345, percentiles:
  	{P0:0.301 P25:0.305 P50:0.31 P75:0.334 P90:0.482 P99:0.572 P100:0.572}
  github.com, count: 27, aveSecs: 0.101, percentiles:
  	{P0:0.058 P25:0.059 P50:0.06 P75:0.122 P90:0.205 P99:0.42 P100:0.42}
  golang.org, count: 21, aveSecs: 0.639, percentiles:
  	{P0:0.558 P25:0.585 P50:0.596 P75:0.661 P90:0.804 P99:0.887 P100:0.887}
  vnexpress.net, count: 28, aveSecs: 0.092, percentiles:
  	{P0:0.051 P25:0.053 P50:0.056 P75:0.06 P90:0.21 P99:0.361 P100:0.361}
  ````

## Config

Default MemoryMetric will never auto reset. You can create a cron to
periodically reset the metric, example once per day.

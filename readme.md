# metric

Package metric is used for observing request count and duration.  

It use an order statistic tree to store durations, so it can calculate
any duration percentile in time complexity log(n).

Wrapped [Go order statistic tree](https://github.com/daominah/GoLLRB).

## Usage

Example using MemoryMetric to measure http handlers duration:

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

## Config

Default MemoryMetric will never auto reset. You can create a cron to
periodically reset the metric, example once per day.

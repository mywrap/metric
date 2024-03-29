// Package metric is used for observing request count and duration,
// it can calculate duration percentile in time complexity log(nRequests),
// default output durations at 0-25-50-75-100-90-95-99-99.5-99.9 percentile.
package metric

import (
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/daominah/GoLLRB/llrb"
)

// Metric monitors number of requests, duration of requests,
// Metric's methods must be safe for concurrent calls
type Metric interface {
	// Count increases key's counter by 1
	Count(key string)
	// Duration increases key's total duration by dur,
	// the dur will be saved in a order statistic tree
	Duration(key string, dur time.Duration)
	// Reset set all count and duration of requests to 0,
	// In a database implement, you can persist the metric before resetting
	Reset()

	GetLastReset() time.Time
	// each row is corresponding to a metric key
	GetCurrentMetric() []RowDisplay
	// percentile is in [0, 1]
	GetDurationPercentile(key string, percentile float64) time.Duration
}

// RowDisplay is human readable metric of a key,
// durations are measured in seconds, rounded to 3 decimal place.
type RowDisplay struct {
	// example of Key: http path_method
	Key            string
	RequestCount   int
	AverageSeconds float64
	Percentiles    struct {
		P0   float64 // min duration, best
		P25  float64 // 25 percentile
		P50  float64 // median
		P75  float64
		P90  float64
		P99  float64
		P100 float64 // max duration, can be infinity, so average is not a good metric
	}
}

func (r RowDisplay) String() string {
	return fmt.Sprintf(
		"key: %v, count: %v, aveSecs: %v, percentiles: %+v",
		r.Key, r.RequestCount, r.AverageSeconds, r.Percentiles)
}

// keys alphabet order
type SortByKey []RowDisplay

func (h SortByKey) Len() int           { return len(h) }
func (h SortByKey) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h SortByKey) Less(i, j int) bool { return h[i].Key < h[j].Key }

type SortByAveDur []RowDisplay

func (h SortByAveDur) Len() int           { return len(h) }
func (h SortByAveDur) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h SortByAveDur) Less(i, j int) bool { return h[i].AverageSeconds > h[j].AverageSeconds }

//
//
//

// MemoryMetric implements Metric interface
type MemoryMetric struct {
	lastReset time.Time
	current   map[string]*Row
	mutex     *sync.Mutex
}

// NewMemoryMetric returns an in-memory implementation of Metric interface
func NewMemoryMetric() *MemoryMetric {
	return &MemoryMetric{
		lastReset: time.Now(),
		current:   make(map[string]*Row),
		mutex:     &sync.Mutex{},
	}
}

// create new row if not existed
func (m *MemoryMetric) getRow(key string) *Row {
	m.mutex.Lock()
	row, found := m.current[key]
	if !found {
		m.current[key] = NewMetricRow()
		row = m.current[key]
	}
	m.mutex.Unlock()
	return row
}

func (m *MemoryMetric) Count(key string) {
	row := m.getRow(key)
	row.Lock()
	row.Count += 1
	row.Unlock()
}

func (m *MemoryMetric) Duration(key string, dur time.Duration) {
	row := m.getRow(key)
	row.Lock()
	row.TotalDuration += dur
	row.Durations.InsertNoReplace(Duration(dur))
	row.Unlock()
}

func (m *MemoryMetric) Reset() {
	m.mutex.Lock()
	m.lastReset = time.Now()
	m.current = make(map[string]*Row)
	m.mutex.Unlock()
}

func (m *MemoryMetric) GetLastReset() time.Time {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.lastReset
}

func (m *MemoryMetric) GetCurrentMetric() []RowDisplay {
	ret := make([]RowDisplay, 0)
	m.mutex.Lock()
	for key, row := range m.current {
		ret = append(ret, row.toDisplay(key))
	}
	m.mutex.Unlock()
	sort.Sort(SortByKey(ret))
	return ret
}

func (m *MemoryMetric) GetDurationPercentile(key string, percentile float64) time.Duration {
	row := m.getRow(key)
	row.Lock()
	ret := calcRowPercentile(row, percentile)
	row.Unlock()
	return ret
}

// Row is an in-memory representation of RowDisplay
type Row struct {
	Count         int
	TotalDuration time.Duration // TODO: handle int64 overflow
	Durations     *llrb.LLRB
	*sync.Mutex
}

func (r *Row) toDisplay(key string) RowDisplay {
	r.Lock()
	defer r.Unlock()
	ret := RowDisplay{Key: key, RequestCount: r.Count}
	if r.Count != 0 {
		aveDur := r.TotalDuration / time.Duration(r.Count)
		ret.AverageSeconds = round(aveDur.Seconds())
	}
	ret.Percentiles.P0 = round(calcRowPercentile(r, 0).Seconds())
	ret.Percentiles.P25 = round(calcRowPercentile(r, .25).Seconds())
	ret.Percentiles.P50 = round(calcRowPercentile(r, .5).Seconds())
	ret.Percentiles.P75 = round(calcRowPercentile(r, .75).Seconds())
	ret.Percentiles.P90 = round(calcRowPercentile(r, .9).Seconds())
	ret.Percentiles.P99 = round(calcRowPercentile(r, .99).Seconds())
	ret.Percentiles.P100 = round(calcRowPercentile(r, 1).Seconds())
	return ret
}

// round to 3 decimal place
func round(f float64) float64 { return math.Round(f*1000) / 1000 }

func NewMetricRow() *Row {
	return &Row{Durations: llrb.New(), Mutex: &sync.Mutex{}}
}

// Duration is time_Duration that implements LLRB's Item interface
type Duration time.Duration

func (d Duration) Less(than llrb.Item) bool {
	tmp, _ := than.(Duration)
	return d < tmp
}

// do not lock row in this func
func calcRowPercentile(row *Row, percentile float64) time.Duration {
	rank := int(math.Ceil(percentile * float64(row.Durations.Len())))
	item := row.Durations.GetByRank(rank)
	dur, ok := item.(Duration)
	if item == nil || !ok {
		return 0
	}
	return time.Duration(dur)
}

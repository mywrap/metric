package metric

import (
	"math/rand"
	"sync"
	"testing"
	"time"
)

func TestCalcPercentile(t *testing.T) {
	a := *NewMetricRow()
	for i := 1; i <= 100; i++ {
		a.Durations.InsertNoReplace(Duration(i))
	}
	for i, c := range []struct {
		percentile float64
		expect     time.Duration
	}{
		{0.6827, 69},
		{0.9545, 96},
		{0.9973, 100},
		{0.99, 99},
		{68, 100},
	} {
		reality := calcRowPercentile(a, c.percentile)
		if reality != c.expect {
			t.Errorf("CalcPercentile %v: expect %v, reality %v", i, c.expect, reality)
		}
	}
	// try empty row
	if calcRowPercentile(*NewMetricRow(), 0.5) != 0 {
		t.Error("try empty row")
	}
}

func TestMetric(t *testing.T) {
	m := NewMemoryMetric()
	wg := &sync.WaitGroup{}
	path0, path1 := "path0", "path1"
	a0, a1 := make([]int, 0), make([]int, 0)
	for i := 1; i <= 1000; i++ {
		a0 = append(a0, i)
		a0 = append(a0, i)
		a1 = append(a1, 2000+2*i)
	}
	rand.Shuffle(len(a0), func(i int, j int) {
		a0[i], a0[j] = a0[j], a0[i]
	})
	rand.Shuffle(len(a1), func(i int, j int) {
		a1[i], a1[j] = a1[j], a1[i]
	})
	for _, c := range []struct {
		path  string
		array []int
	}{{path0, a0}, {path1, a1}} {
		for i := 0; i < len(c.array); i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Add(-1)
				m.Count(c.path)
				m.Duration(c.path, time.Duration(c.array[i])*time.Millisecond)
			}(i)
		}
		wg.Wait()
	}
	a := m.GetCurrentMetric()
	if len(a) != 2 {
		t.Fatal(len(a))
	}
	if a[0].Key != path0 || a[1].Key != path1 {
		t.Error(a)
	}
	expectation := PG1{P0: 0.001, P25: 0.25, P50: 0.5, P75: 0.75, P100: 1}
	if a[0].RequestCount != 2000 || a[0].PercentilesG1 != expectation {
		t.Error(a[0])
	}
	expectation2 := PG2{P90: 3.8, P95: 3.9, P99: 3.98, P995: 3.99, P999: 3.998}
	if a[1].RequestCount != 1000 || a[1].PercentilesG2 != expectation2 {
		t.Error(a[1])
	}
}

func TestMemoryMetric_Reset(t *testing.T) {
	m := NewMemoryMetric()
	m.Count("key0")
	m.Duration("key0", 1*time.Second)
	m.Count("key0")
	m.Duration("key0", 2*time.Second)
	m.Reset()
	m.Count("key0")
	m.Duration("key0", 4*time.Second)
	rows := m.GetCurrentMetric()
	if rows[0].RequestCount != 1 ||
		rows[0].AverageSeconds != 4 {
		t.Error(rows[0])
	}
}

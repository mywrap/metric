package metric

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func TestCalcPercentile(t *testing.T) {
	a := NewMetricRow()
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
		{68, 100}, // percentile 68 equivalent to 1, test overflow
	} {
		reality := calcRowPercentile(a, c.percentile)
		if reality != c.expect {
			t.Errorf("CalcPercentile %v: expect %v, reality %v", i, c.expect, reality)
		}
	}
}

func TestCalcPercentileEmpty(t *testing.T) {
	if calcRowPercentile(NewMetricRow(), 0.5) != 0 {
		t.Error("try empty row")
	}
}

func TestMemoryMetric(t *testing.T) {
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
	// a[0] durations: [1, 1, 2, 2, 3, 3, .., 1000, 1000] milliseconds
	if a[0].Percentiles.P0 != 0.001 || a[0].Percentiles.P50 != 0.5 ||
		a[0].Percentiles.P75 != 0.75 || a[0].Percentiles.P100 != 1 {
		t.Errorf("error MemoryMetric Percentiles, got: %+v", a[0])
	}
	// a[0] durations: [2002, 2004, 2006, .., 3998, 4000] milliseconds
	if a[1].Percentiles.P25 != 2.5 || a[1].Percentiles.P90 != 3.8 ||
		a[1].Percentiles.P99 != 3.98 || a[1].Percentiles.P100 != 4 {
		t.Errorf("error MemoryMetric Percentiles, got: %+v", a[1])
	}

	if got, want := m.GetDurationPercentile(path1, 0.99), 3980*time.Millisecond; got != want {
		t.Errorf("error GetDurationPercentile got %v, want %v", got, want)
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

// CPU i7-1260P 1170 ns/op
func BenchmarkMemoryMetric_Duration(b *testing.B) {
	m := NewMemoryMetric()
	for i := 0; i < b.N; i++ {
		m.Duration(
			fmt.Sprintf("key%v", rand.Intn(50)),
			time.Duration(rand.Intn(b.N))*time.Millisecond,
		)
	}
	rows := m.GetCurrentMetric()
	for _, row := range rows {
		_ = row
		//b.Log(row)
	}
}

// CPU i7-1260P 774.3 ns/op
func BenchmarkMemoryMetric_Duration2(b *testing.B) {
	m := NewMemoryMetric()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() { // The loop body is executed b.N times in goroutines
			m.Duration(
				fmt.Sprintf("key%v", rand.Intn(50)),
				time.Duration(rand.Intn(b.N))*time.Millisecond,
			)
		}
	})
}

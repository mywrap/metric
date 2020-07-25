package main

import (
	"github.com/mywrap/metric"
)

func main() {
	m := metric.NewMemoryMetric()
	m.GetCurrentMetric()
}

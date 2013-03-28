// Copyright 2013 Prometheus Team
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"fmt"
	"github.com/prometheus/prometheus/model"
	"github.com/prometheus/prometheus/rules"
	"github.com/prometheus/prometheus/rules/ast"
	"github.com/prometheus/prometheus/storage/metric"
	"log"
	"os"
	"time"
)

// Commandline flags.
var (
	metricsStoragePath = flag.String("metricsStoragePath", "/tmp/metrics_bench", "Base path for metrics storage.")

	// Sample data generation flags.
	populateStorage        = flag.Bool("populateStorage", true, "Whether to populate the storage initially.")
	deleteStorage          = flag.Bool("deleteStorage", true, "Whether to remove the storage path before exiting.")
	numTimeseries          = flag.Int("numTimeseries", 100, "Number of timeseries to generate.")
	numValuesPerTimeseries = flag.Int("numValuesPerTimeseries", 10000, "Number of data points per timeseries to generate.")
	numLabels              = flag.Int("numLabels", 5, "Number of labels to attach to timeseries.")
	valueIntervalSeconds   = flag.Int("valueIntervalSeconds", 15, "Time in seconds between generated data points.")

	// Reading benchmark flags.
	numReadIterations   = flag.Int("numReadIterations", 5, "How often to run and time each expression.")
	evalIntervalSeconds = flag.Int("evalIntervalSeconds", 24*60*60, "Interval in seconds over which to evaluate expressions.")
)

var expressions = []string{
	// Simple vector literal lookup.
	"metric_0",
	// Interesting to see whether we load the same timeseries twice.
	"metric_0 + metric_0",
	// Comparing the above with two different timeseries.
	"metric_0 + metric_1",
	// Testing how timings increase with increasing rate ranges.
	"rate(metric_0[5m])",
	"rate(metric_0[15m])",
	"rate(metric_0[1h])",
}

// Creates a group timeseries with the format:
//
// metric_0{label0="value", label1="value", ...}
// metric_1{label0="value", label1="value", ...}
// metric_2{label0="value", label1="value", ...}
// ...
//
// Only the metric name is varied between timeseries, since label variance
// should otherwise not matter for lookup speeds (all labels are hashed
// together, and this fingerprint will be completely different for all these
// timeseries). The non-varying labels on the timeseries are mainly added to
// simulate real fingerprinting times, etc.
//
// The number of static labels is determined by -numLabels, while the number of
// timeseries is set by -numTimeseries.
func populateTestStorage(storage metric.Storage) {
	metric := model.Metric{}

	for i := 0; i < *numLabels; i++ {
		metric[model.LabelName(fmt.Sprintf("label%d", i))] = "value"
	}

	interval := time.Duration(*valueIntervalSeconds) * time.Second
	startTime := time.Time{}
	endTime := startTime.Add(interval * time.Duration(*numValuesPerTimeseries))
	for ts := 0; ts < *numTimeseries; ts++ {
		metric["name"] = model.LabelValue(fmt.Sprintf("metric_%d", ts))
		for t := startTime; t.Before(endTime); t = t.Add(interval) {
			sample := model.Sample{
				Metric:    metric,
				Value:     12345.6789, // The sample value shouldn't really matter.
				Timestamp: t,
			}
			storage.AppendSample(sample)
		}
	}
	storage.Flush()
}

func doBenchmark(storage metric.Storage, expression string) {
	fmt.Printf("\n==== Expression: '%s' ====\n", expression)
	exprNode, err := rules.LoadExprFromString(expression)
	if err != nil {
		log.Fatalf("Error parsing expression: %v", err)
	}

	startTime := time.Time{}
	endTime := startTime.Add(time.Duration(*evalIntervalSeconds) * time.Second)
	step := time.Duration(*valueIntervalSeconds) * time.Second
	fmt.Printf("Start time: %v\n", startTime)
	fmt.Printf("End time: %v\n", endTime)

	totalTime := time.Duration(0)
	for i := 0; i < *numReadIterations; i++ {
		evalStartTime := time.Now()
		_, err := ast.EvalVectorRange(exprNode.(ast.VectorNode), startTime, endTime, step)
		if err != nil {
			log.Fatalf("Error evaluating expression: %s", err)
		}
		iterationTime := time.Since(evalStartTime)
		//fmt.Printf("Iteration time: %v\n", iterationTime)
		totalTime += iterationTime
	}
	fmt.Printf("Average time: %v\n", totalTime/time.Duration(*numReadIterations))
}

func main() {
	flag.Parse()

	if *evalIntervalSeconds > (*valueIntervalSeconds * *numValuesPerTimeseries) {
		log.Fatalf("Evaluation interval (-evalIntervalSeconds) is greater than the stored range (-valueIntervalSeconds * -numValuesPerTimeseries)\n")
	}

	err := os.MkdirAll(*metricsStoragePath, 0755)
	if err != nil {
		log.Fatalf("Error creating storage directory: %v", err)
	}
	storage, err := metric.NewTieredStorage(5000, 5000, 100, time.Second*30, time.Second*1, time.Second*20, *metricsStoragePath)
	if err != nil {
		log.Fatalf("Error opening storage: %v", err)
	}
	go storage.Serve()
	defer func() {
		storage.Close()
		if *deleteStorage {
			os.RemoveAll(*metricsStoragePath)
		}
	}()

	ast.SetStorage(storage)
	if *populateStorage {
		populateTestStorage(storage)
	}

	for _, expression := range expressions {
		doBenchmark(storage, expression)
	}
}

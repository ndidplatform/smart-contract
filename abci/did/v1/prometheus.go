/**
 * Copyright (c) 2018, 2019 National Digital ID COMPANY LIMITED
 *
 * This file is part of NDID software.
 *
 * NDID is the free software: you can redistribute it and/or modify it under
 * the terms of the Affero GNU General Public License as published by the
 * Free Software Foundation, either version 3 of the License, or any later
 * version.
 *
 * NDID is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
 * See the Affero GNU General Public License for more details.
 *
 * You should have received a copy of the Affero GNU General Public License
 * along with the NDID source code. If not, see https://www.gnu.org/licenses/agpl.txt.
 *
 * Please contact info@ndid.co.th for any further questions
 *
 */

package did

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	prometheus.MustRegister(checkTxCounter)
	prometheus.MustRegister(checkTxFailCounter)
	prometheus.MustRegister(checkTxDurationHistogram)
	prometheus.MustRegister(deliverTxCounter)
	prometheus.MustRegister(deliverTxFailCounter)
	prometheus.MustRegister(deliverTxDurationHistogram)
	prometheus.MustRegister(queryCounter)
	prometheus.MustRegister(queryDurationHistogram)
	prometheus.MustRegister(commitDurationHistogram)
	prometheus.MustRegister(dbSaveDurationHistogram)
	prometheus.MustRegister(appHashDurationHistogram)
}

func recordCheckTxMetrics(fName string) {
	checkTxCounter.With(prometheus.Labels{"function": fName}).Inc()
}

var (
	checkTxCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Subsystem: "abci",
		Name:      "check_tx_total",
		Help:      "Total number of CheckTx function called",
	},
		[]string{"function"})
)

func recordCheckTxFailMetrics(fName string) {
	checkTxFailCounter.With(prometheus.Labels{"function": fName}).Inc()
}

var (
	checkTxFailCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Subsystem: "abci",
		Name:      "check_tx_fails_total",
		Help:      "Total number of failed CheckTx",
	},
		[]string{"function"})
)

func recordCheckTxDurationMetrics(startTime time.Time, fName string) {
	duration := time.Since(startTime)
	checkTxDurationHistogram.WithLabelValues(fName).Observe(duration.Seconds())
}

var (
	checkTxDurationHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Subsystem: "abci",
		Name:      "check_tx_duration_seconds",
		Help:      "Duration of CheckTx in seconds",
		Buckets:   []float64{0.05, 0.1, 0.25, 0.5, 0.75, 1},
	},
		[]string{"function"},
	)
)

func recordDeliverTxMetrics(fName string) {
	deliverTxCounter.With(prometheus.Labels{"function": fName}).Inc()
}

var (
	deliverTxCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Subsystem: "abci",
		Name:      "deliver_tx_total",
		Help:      "Total number of DeliverTx function called",
	},
		[]string{"function"},
	)
)

func recordDeliverTxFailMetrics(fName string) {
	deliverTxFailCounter.With(prometheus.Labels{"function": fName}).Inc()
}

var (
	deliverTxFailCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Subsystem: "abci",
		Name:      "deliver_tx_fails_total",
		Help:      "Total number of failed DeliverTx",
	},
		[]string{"function"},
	)
)

func recordDeliverTxDurationMetrics(startTime time.Time, fName string) {
	duration := time.Since(startTime)
	deliverTxDurationHistogram.WithLabelValues(fName).Observe(duration.Seconds())
}

var (
	deliverTxDurationHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Subsystem: "abci",
		Name:      "deliver_tx_duration_seconds",
		Help:      "Duration of DeliverTx in seconds",
		Buckets:   []float64{0.05, 0.1, 0.25, 0.5, 0.75, 1},
	},
		[]string{"function"},
	)
)

func recordQueryMetrics(fName string) {
	queryCounter.With(prometheus.Labels{"function": fName}).Inc()
}

var (
	queryCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Subsystem: "abci",
		Name:      "query_total",
		Help:      "Total number of Query function called",
	},
		[]string{"function"},
	)
)

func recordQueryDurationMetrics(startTime time.Time, fName string) {
	duration := time.Since(startTime)
	queryDurationHistogram.WithLabelValues(fName).Observe(duration.Seconds())
}

var (
	queryDurationHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Subsystem: "abci",
		Name:      "query_duration_seconds",
		Help:      "Duration of Query in seconds",
		Buckets:   []float64{0.05, 0.1, 0.25, 0.5, 0.75, 1},
	},
		[]string{"function"},
	)
)

func recordCommitDurationMetrics(startTime time.Time) {
	duration := time.Since(startTime)
	commitDurationHistogram.Observe(duration.Seconds())
}

var (
	commitDurationHistogram = prometheus.NewHistogram(prometheus.HistogramOpts{
		Subsystem: "abci",
		Name:      "commit_duration_seconds",
		Help:      "Duration of Commit in seconds",
		Buckets:   []float64{0.05, 0.1, 0.25, 0.5, 0.75, 1},
	},
	)
)

func recordDBSaveDurationMetrics(startTime time.Time) {
	duration := time.Since(startTime)
	dbSaveDurationHistogram.Observe(duration.Seconds())
}

var (
	dbSaveDurationHistogram = prometheus.NewHistogram(prometheus.HistogramOpts{
		Subsystem: "abci",
		Name:      "db_save_duration_seconds",
		Help:      "Duration of DB save in seconds",
		Buckets:   []float64{0.05, 0.1, 0.25, 0.5, 0.75, 1},
	},
	)
)

func recordAppHashDurationMetrics(startTime time.Time) {
	duration := time.Since(startTime)
	appHashDurationHistogram.Observe(duration.Seconds())
}

var (
	appHashDurationHistogram = prometheus.NewHistogram(prometheus.HistogramOpts{
		Subsystem: "abci",
		Name:      "app_hash_duration_seconds",
		Help:      "Duration of app hash calculation in seconds",
		Buckets:   []float64{0.05, 0.1, 0.25, 0.5, 0.75, 1},
	},
	)
)

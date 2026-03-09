package handler

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
)

// PgxPoolCollector exports pgxpool.Stat() as Prometheus metrics.
// It implements the prometheus.Collector interface so metrics are
// read on each /metrics scrape — no background goroutine needed.
type PgxPoolCollector struct {
	pool *pgxpool.Pool

	acquiredConns  *prometheus.Desc
	idleConns      *prometheus.Desc
	totalConns     *prometheus.Desc
	maxConns       *prometheus.Desc
	acquireCount   *prometheus.Desc
	acquireDurSec  *prometheus.Desc
	canceledAcqCnt *prometheus.Desc
	emptyAcqCount  *prometheus.Desc
}

// NewPgxPoolCollector creates a collector for the given pool.
func NewPgxPoolCollector(pool *pgxpool.Pool) *PgxPoolCollector {
	return &PgxPoolCollector{
		pool: pool,
		acquiredConns: prometheus.NewDesc(
			"pgx_pool_acquired_connections",
			"Number of currently acquired connections.",
			nil, nil,
		),
		idleConns: prometheus.NewDesc(
			"pgx_pool_idle_connections",
			"Number of idle connections in the pool.",
			nil, nil,
		),
		totalConns: prometheus.NewDesc(
			"pgx_pool_total_connections",
			"Total number of connections in the pool.",
			nil, nil,
		),
		maxConns: prometheus.NewDesc(
			"pgx_pool_max_connections",
			"Maximum number of connections allowed.",
			nil, nil,
		),
		acquireCount: prometheus.NewDesc(
			"pgx_pool_acquire_count_total",
			"Total number of successful connection acquires.",
			nil, nil,
		),
		acquireDurSec: prometheus.NewDesc(
			"pgx_pool_acquire_duration_seconds_total",
			"Total cumulative time spent acquiring connections.",
			nil, nil,
		),
		canceledAcqCnt: prometheus.NewDesc(
			"pgx_pool_canceled_acquire_count_total",
			"Total number of acquires canceled by context.",
			nil, nil,
		),
		emptyAcqCount: prometheus.NewDesc(
			"pgx_pool_empty_acquire_count_total",
			"Total number of acquires when pool was empty.",
			nil, nil,
		),
	}
}

func (c *PgxPoolCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.acquiredConns
	ch <- c.idleConns
	ch <- c.totalConns
	ch <- c.maxConns
	ch <- c.acquireCount
	ch <- c.acquireDurSec
	ch <- c.canceledAcqCnt
	ch <- c.emptyAcqCount
}

func (c *PgxPoolCollector) Collect(ch chan<- prometheus.Metric) {
	stat := c.pool.Stat()

	ch <- prometheus.MustNewConstMetric(c.acquiredConns, prometheus.GaugeValue, float64(stat.AcquiredConns()))
	ch <- prometheus.MustNewConstMetric(c.idleConns, prometheus.GaugeValue, float64(stat.IdleConns()))
	ch <- prometheus.MustNewConstMetric(c.totalConns, prometheus.GaugeValue, float64(stat.TotalConns()))
	ch <- prometheus.MustNewConstMetric(c.maxConns, prometheus.GaugeValue, float64(stat.MaxConns()))
	ch <- prometheus.MustNewConstMetric(c.acquireCount, prometheus.CounterValue, float64(stat.AcquireCount()))
	ch <- prometheus.MustNewConstMetric(c.acquireDurSec, prometheus.CounterValue, stat.AcquireDuration().Seconds())
	ch <- prometheus.MustNewConstMetric(c.canceledAcqCnt, prometheus.CounterValue, float64(stat.CanceledAcquireCount()))
	ch <- prometheus.MustNewConstMetric(c.emptyAcqCount, prometheus.CounterValue, float64(stat.EmptyAcquireCount()))
}

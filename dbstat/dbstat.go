package dbstat

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
)

type DBStats struct {
	db     *sql.DB
	prefix string
}

func New(db *sql.DB, prefix string) *DBStats {
	return &DBStats{
		db:     db,
		prefix: prefix,
	}
}

func (dbs *DBStats) Describe(ch chan<- *prometheus.Desc) {
	metricCh := make(chan prometheus.Metric)
	doneCh := make(chan struct{})

	go func() {
		for m := range metricCh {
			ch <- m.Desc()
		}
		close(doneCh)
	}()

	dbs.Collect(metricCh)
	close(metricCh)
	<-doneCh
}

func (dbs *DBStats) Collect(ch chan<- prometheus.Metric) {
	stats := dbs.db.Stats()

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(dbs.prefix+"_max_open_connections", "Maximum number of open connections to the database.", nil, nil),
		prometheus.GaugeValue,
		float64(stats.MaxOpenConnections),
	)

	// pool status
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(dbs.prefix+"_open_connections", "The number of established connections both in use and idle.", nil, nil),
		prometheus.GaugeValue,
		float64(stats.OpenConnections),
	)
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(dbs.prefix+"_in_use", "The number of connections currently in use.", nil, nil),
		prometheus.GaugeValue,
		float64(stats.InUse),
	)
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(dbs.prefix+"_idle", "The number of idle connections.", nil, nil),
		prometheus.GaugeValue,
		float64(stats.Idle),
	)

	// counters
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(dbs.prefix+"_wait_count_total", "The total number of connections waited for.", nil, nil),
		prometheus.CounterValue,
		float64(stats.WaitCount),
	)
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(dbs.prefix+"_wait_duration_seconds_total", "The total time blocked waiting for a new connection.", nil, nil),
		prometheus.CounterValue,
		stats.WaitDuration.Seconds(),
	)
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(dbs.prefix+"_max_idle_closed_total", "The total number of connections closed due to SetMaxIdleConns.", nil, nil),
		prometheus.CounterValue,
		float64(stats.MaxIdleClosed),
	)
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(dbs.prefix+"_max_lifetime_closed_total", "The total number of connections closed due to SetConnMaxLifetime.", nil, nil),
		prometheus.CounterValue,
		float64(stats.MaxLifetimeClosed),
	)
}

// check interfaces
var (
	_ prometheus.Collector = (*DBStats)(nil)
)

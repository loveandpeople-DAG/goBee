package metrics

import (
	"time"

	"github.com/loveandpeople-DAG/goHive/daemon"
	"github.com/loveandpeople-DAG/goHive/node"
	"github.com/loveandpeople-DAG/goHive/timeutil"

	"github.com/loveandpeople-DAG/goBee/pkg/shutdown"
)

var PLUGIN = node.NewPlugin("Metrics", node.Enabled, configure, run)

func configure(_ *node.Plugin) {
	// nothing
}

func run(_ *node.Plugin) {
	// create a background worker that "measures" the TPS value every second
	daemon.BackgroundWorker("Metrics TPS Updater", func(shutdownSignal <-chan struct{}) {
		timeutil.Ticker(measureTPS, 1*time.Second, shutdownSignal)
	}, shutdown.PriorityMetricsUpdater)
}

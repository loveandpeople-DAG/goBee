package curl

import (
	"sync"
	"time"

	"github.com/loveandpeople-DAG/goClient/consts"
	"github.com/loveandpeople-DAG/goHive/daemon"
	"github.com/loveandpeople-DAG/goHive/logger"
	"github.com/loveandpeople-DAG/goHive/node"

	"github.com/loveandpeople-DAG/goBee/pkg/batcher"
	"github.com/loveandpeople-DAG/goBee/pkg/shutdown"
)

const (
	inputSize = consts.TransactionTrinarySize
	timeout   = 50 * time.Millisecond
)

var (
	PLUGIN     = node.NewPlugin("Curl", node.Enabled, configure, run)
	log        *logger.Logger
	hasher     *batcher.Curl
	hasherOnce sync.Once
)

// Hasher returns the batched Curl singleton.
func Hasher() *batcher.Curl {
	hasherOnce.Do(func() {
		// create a new batched Curl instance to compute transaction hashes
		// on average amd64 hardware, even a single worker can hash about 100Mb/s; this is sufficient for all scenarios
		// TODO: verify performance on arm (especially 32bit) that >1 worker is indeed not needed and beneficial
		hasher = batcher.NewCurlP81(inputSize, timeout, 1)
	})
	return hasher
}

func configure(plugin *node.Plugin) {
	log = logger.NewLogger(plugin.Name)

	// assure that the hasher is initialized
	Hasher()
}

func run(_ *node.Plugin) {
	// close the hasher on shutdown
	daemon.BackgroundWorker("Curl batched hashing", func(shutdownSignal <-chan struct{}) {
		log.Info("Starting Curl batched hashing ... done")
		<-shutdownSignal
		log.Info("Stopping Curl batched hashing ...")
		if err := Hasher().Close(); err != nil {
			log.Errorf("Stopping Curl batched hashing: %s", err)
		} else {
			log.Info("Stopping Curl batched hashing ... done")
		}
	}, shutdown.PriorityCurlHasher)
}

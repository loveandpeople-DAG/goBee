package pow

import (
	"sync"
	"time"

	"github.com/loveandpeople-DAG/goHive/daemon"
	"github.com/loveandpeople-DAG/goHive/logger"
	"github.com/loveandpeople-DAG/goHive/node"

	"github.com/loveandpeople-DAG/goBee/pkg/config"
	powpackage "github.com/loveandpeople-DAG/goBee/pkg/pow"
	"github.com/loveandpeople-DAG/goBee/pkg/shutdown"
)

const (
	powsrvInitCooldown = 30 * time.Second
)

var (
	PLUGIN      = node.NewPlugin("PoW", node.Enabled, configure, run)
	log         *logger.Logger
	handler     *powpackage.Handler
	handlerOnce sync.Once
)

// Handler gets the pow handler instance.
func Handler() *powpackage.Handler {
	handlerOnce.Do(func() {
		// init the pow handler with all possible settings
		powsrvAPIKey, _ := config.LoadHashFromEnvironment("POWSRV_API_KEY", 12)
		handler = powpackage.New(log, powsrvAPIKey, powsrvInitCooldown)

	})
	return handler
}

func configure(plugin *node.Plugin) {
	log = logger.NewLogger(plugin.Name)

	// init pow handler
	Handler()
}

func run(_ *node.Plugin) {

	// close the PoW handler on shutdown
	daemon.BackgroundWorker("PoW Handler", func(shutdownSignal <-chan struct{}) {
		log.Info("Starting PoW Handler ... done")
		<-shutdownSignal
		log.Info("Stopping PoW Handler ...")
		Handler().Close()
		log.Info("Stopping PoW Handler ... done")
	}, shutdown.PriorityPoWHandler)
}

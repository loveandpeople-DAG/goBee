package dashboard

import (
	"github.com/loveandpeople-DAG/goHive/daemon"
	"github.com/loveandpeople-DAG/goHive/events"
	"github.com/loveandpeople-DAG/goHive/node"

	"github.com/loveandpeople-DAG/goBee/pkg/shutdown"
	"github.com/loveandpeople-DAG/goBee/pkg/tipselect"
	"github.com/loveandpeople-DAG/goBee/plugins/urts"
)

func runTipSelMetricWorker() {

	// check if URTS plugin is enabled
	if node.IsSkipped(urts.PLUGIN) {
		return
	}

	onTipSelPerformed := events.NewClosure(func(metrics *tipselect.TipSelStats) {
		hub.BroadcastMsg(&Msg{Type: MsgTypeTipSelMetric, Data: metrics})
	})

	daemon.BackgroundWorker("Dashboard[TipSelMetricUpdater]", func(shutdownSignal <-chan struct{}) {
		urts.TipSelector.Events.TipSelPerformed.Attach(onTipSelPerformed)
		<-shutdownSignal
		log.Info("Stopping Dashboard[TipSelMetricUpdater] ...")
		urts.TipSelector.Events.TipSelPerformed.Detach(onTipSelPerformed)
		log.Info("Stopping Dashboard[TipSelMetricUpdater] ... done")
	}, shutdown.PriorityDashboard)
}

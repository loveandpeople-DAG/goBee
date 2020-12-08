package urts

import (
	"time"

	"github.com/loveandpeople-DAG/goHive/daemon"
	"github.com/loveandpeople-DAG/goHive/events"
	"github.com/loveandpeople-DAG/goHive/logger"
	"github.com/loveandpeople-DAG/goHive/node"

	"github.com/loveandpeople-DAG/goBee/pkg/config"
	"github.com/loveandpeople-DAG/goBee/pkg/dag"
	"github.com/loveandpeople-DAG/goBee/pkg/model/tangle"
	"github.com/loveandpeople-DAG/goBee/pkg/shutdown"
	"github.com/loveandpeople-DAG/goBee/pkg/tipselect"
	"github.com/loveandpeople-DAG/goBee/pkg/whiteflag"
	tangleplugin "github.com/loveandpeople-DAG/goBee/plugins/tangle"
)

var (
	PLUGIN = node.NewPlugin("URTS", node.Enabled, configure, run)
	log    *logger.Logger

	TipSelector *tipselect.TipSelector

	// Closures
	onBundleSolid        *events.Closure
	onMilestoneConfirmed *events.Closure
)

func configure(plugin *node.Plugin) {
	log = logger.NewLogger(plugin.Name)

	TipSelector = tipselect.New(
		config.NodeConfig.GetInt(config.CfgTipSelMaxDeltaTxYoungestRootSnapshotIndexToLSMI),
		config.NodeConfig.GetInt(config.CfgTipSelMaxDeltaTxOldestRootSnapshotIndexToLSMI),
		config.NodeConfig.GetInt(config.CfgTipSelBelowMaxDepth),

		config.NodeConfig.GetInt(config.CfgTipSelNonLazy+config.CfgTipSelRetentionRulesTipsLimit),
		time.Duration(time.Second*time.Duration(config.NodeConfig.GetInt(config.CfgTipSelNonLazy+config.CfgTipSelMaxReferencedTipAgeSeconds))),
		config.NodeConfig.GetUint32(config.CfgTipSelNonLazy+config.CfgTipSelMaxApprovers),
		config.NodeConfig.GetInt(config.CfgTipSelNonLazy+config.CfgTipSelSpammerTipsThreshold),

		config.NodeConfig.GetInt(config.CfgTipSelSemiLazy+config.CfgTipSelRetentionRulesTipsLimit),
		time.Duration(time.Second*time.Duration(config.NodeConfig.GetInt(config.CfgTipSelSemiLazy+config.CfgTipSelMaxReferencedTipAgeSeconds))),
		config.NodeConfig.GetUint32(config.CfgTipSelSemiLazy+config.CfgTipSelMaxApprovers),
		config.NodeConfig.GetInt(config.CfgTipSelSemiLazy+config.CfgTipSelSpammerTipsThreshold),
	)

	configureEvents()
}

func run(_ *node.Plugin) {
	daemon.BackgroundWorker("Tipselection[Events]", func(shutdownSignal <-chan struct{}) {
		attachEvents()
		<-shutdownSignal
		detachEvents()
	}, shutdown.PriorityTipselection)

	daemon.BackgroundWorker("Tipselection[Cleanup]", func(shutdownSignal <-chan struct{}) {
		for {
			select {
			case <-shutdownSignal:
				return
			case <-time.After(time.Second):
				ts := time.Now()
				removedTipCount := TipSelector.CleanUpReferencedTips()
				log.Debugf("CleanUpReferencedTips finished, removed: %d, took: %v", removedTipCount, time.Since(ts).Truncate(time.Millisecond))
			}
		}
	}, shutdown.PriorityTipselection)
}

func configureEvents() {
	onBundleSolid = events.NewClosure(func(cachedBndl *tangle.CachedBundle) {
		cachedBndl.ConsumeBundle(func(bndl *tangle.Bundle) { // bundle -1
			// do not add tips during syncing, because it is not needed at all
			if !tangle.IsNodeSyncedWithThreshold() {
				return
			}

			if bndl.IsInvalidPastCone() || !bndl.IsValid() || !bndl.ValidStrictSemantics() {
				// ignore invalid bundles or semantically invalid bundles or bundles with invalid past cone
				return
			}

			TipSelector.AddTip(bndl)
		})
	})

	onMilestoneConfirmed = events.NewClosure(func(confirmation *whiteflag.Confirmation) {
		// do not propagate during syncing, because it is not needed at all
		if !tangle.IsNodeSyncedWithThreshold() {
			return
		}

		// propagate new transaction root snapshot indexes to the future cone for URTS
		ts := time.Now()
		dag.UpdateTransactionRootSnapshotIndexes(confirmation.Mutations.TailsReferenced, confirmation.MilestoneIndex)
		log.Debugf("UpdateTransactionRootSnapshotIndexes finished, took: %v", time.Since(ts).Truncate(time.Millisecond))

		ts = time.Now()
		removedTipCount := TipSelector.UpdateScores()
		log.Debugf("UpdateScores finished, removed: %d, took: %v", removedTipCount, time.Since(ts).Truncate(time.Millisecond))
	})
}

func attachEvents() {
	tangleplugin.Events.BundleSolid.Attach(onBundleSolid)
	tangleplugin.Events.MilestoneConfirmed.Attach(onMilestoneConfirmed)
}

func detachEvents() {
	tangleplugin.Events.BundleSolid.Detach(onBundleSolid)
	tangleplugin.Events.MilestoneConfirmed.Detach(onMilestoneConfirmed)
}

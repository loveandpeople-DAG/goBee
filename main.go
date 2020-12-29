package main

import (
	"github.com/loveandpeople-DAG/goHive/node"

	"github.com/loveandpeople-DAG/goBee/pkg/config"
	"github.com/loveandpeople-DAG/goBee/pkg/toolset"
	"github.com/loveandpeople-DAG/goBee/plugins/autopeering"
	"github.com/loveandpeople-DAG/goBee/plugins/cli"
	"github.com/loveandpeople-DAG/goBee/plugins/coordinator"
	"github.com/loveandpeople-DAG/goBee/plugins/curl"
	"github.com/loveandpeople-DAG/goBee/plugins/dashboard"
	"github.com/loveandpeople-DAG/goBee/plugins/database"
	"github.com/loveandpeople-DAG/goBee/plugins/gossip"
	"github.com/loveandpeople-DAG/goBee/plugins/gracefulshutdown"
	"github.com/loveandpeople-DAG/goBee/plugins/metrics"
	"github.com/loveandpeople-DAG/goBee/plugins/mqtt"
	"github.com/loveandpeople-DAG/goBee/plugins/peering"
	"github.com/loveandpeople-DAG/goBee/plugins/pow"
	"github.com/loveandpeople-DAG/goBee/plugins/profiling"
	"github.com/loveandpeople-DAG/goBee/plugins/prometheus"
	"github.com/loveandpeople-DAG/goBee/plugins/snapshot"
	"github.com/loveandpeople-DAG/goBee/plugins/spammer"
	"github.com/loveandpeople-DAG/goBee/plugins/tangle"
	"github.com/loveandpeople-DAG/goBee/plugins/urts"
	"github.com/loveandpeople-DAG/goBee/plugins/warpsync"
	"github.com/loveandpeople-DAG/goBee/plugins/webapi"
	"github.com/loveandpeople-DAG/goBee/plugins/zmq"
)

func main() {
	cli.ParseFlags()
	cli.PrintVersion()
	cli.ParseConfig()
	toolset.HandleTools()
	cli.PrintConfig()

	plugins := []*node.Plugin{
		cli.PLUGIN,
		gracefulshutdown.PLUGIN,
		profiling.PLUGIN,
		database.PLUGIN,
		curl.PLUGIN,
		autopeering.PLUGIN,
		webapi.PLUGIN,
	}

	if !config.NodeConfig.GetBool(config.CfgNetAutopeeringRunAsEntryNode) {
		plugins = append(plugins, []*node.Plugin{
			pow.PLUGIN,
			gossip.PLUGIN,
			tangle.PLUGIN,
			peering.PLUGIN,
			warpsync.PLUGIN,
			urts.PLUGIN,
			metrics.PLUGIN,
			snapshot.PLUGIN,
			dashboard.PLUGIN,
			zmq.PLUGIN,
			mqtt.PLUGIN,
			spammer.PLUGIN,
			coordinator.PLUGIN,
			prometheus.PLUGIN,
		}...)
	}

	node.Run(node.Plugins(plugins...))
}

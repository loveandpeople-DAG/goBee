package webapi

import (
	"github.com/loveandpeople-DAG/goBee/plugins/autopeering"
	"github.com/loveandpeople-DAG/goHive/node"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/loveandpeople-DAG/goBee/pkg/config"
	"github.com/loveandpeople-DAG/goBee/pkg/metrics"
	"github.com/loveandpeople-DAG/goBee/pkg/model/tangle"
	"github.com/loveandpeople-DAG/goBee/plugins/cli"
	"github.com/loveandpeople-DAG/goBee/plugins/gossip"
	"github.com/loveandpeople-DAG/goBee/plugins/peering"
	tangleplugin "github.com/loveandpeople-DAG/goBee/plugins/tangle"
	wapi "github.com/loveandpeople-DAG/goClient/api"
	"github.com/loveandpeople-DAG/goClient/consts"
)

var nodeStartAt = time.Now()

func init() {
	addEndpoint("getNodeInfo", getNodeInfo, implementedAPIcalls)
	addEndpoint("getNodeAPIConfiguration", getNodeAPIConfiguration, implementedAPIcalls)
}

func getNodeInfo(_ interface{}, c *gin.Context, _ <-chan struct{}) {
	// Basic info data
	result := wapi.GetNodeInfoResponse{
		AppName:    cli.AppName,
		AppVersion: cli.AppVersion,
	}

	// Node Alias
	if config.NodeConfig.GetBool(config.CfgNodeShowAliasInGetNodeInfo) {
		result.NodeAlias = config.NodeConfig.GetString(config.CfgNodeAlias)
	}

	// Number of peers
	result.Neighbors = uint(peering.Manager().ConnectedPeerCount())

	// Latest milestone index
	lmi := tangle.GetLatestMilestoneIndex()
	result.LatestMilestoneIndex = uint32(lmi)
	result.LatestMilestone = consts.NullHashTrytes

	// Latest milestone hash
	cachedLatestMs := tangle.GetMilestoneOrNil(lmi) // bundle +1
	if cachedLatestMs != nil {
		result.LatestMilestone = cachedLatestMs.GetBundle().GetMilestoneHash().Trytes()
		cachedLatestMs.Release(true) // bundle -1
	}

	// Solid milestone index
	smi := tangle.GetSolidMilestoneIndex()
	result.LatestSolidSubtangleMilestoneIndex = uint32(smi)
	result.LatestSolidSubtangleMilestone = consts.NullHashTrytes
	result.IsSynced = tangle.IsNodeSyncedWithThreshold()
	result.Health = tangleplugin.IsNodeHealthy()

	// Solid milestone hash
	cachedSolidMs := tangle.GetMilestoneOrNil(smi) // bundle +1
	if cachedSolidMs != nil {
		result.LatestSolidSubtangleMilestone = cachedSolidMs.GetBundle().GetMilestoneHash().Trytes()
		cachedSolidMs.Release(true) // bundle -1
	}

	// Milestone start index
	snapshotInfo := tangle.GetSnapshotInfo()
	if snapshotInfo != nil {
		result.MilestoneStartIndex = uint32(snapshotInfo.PruningIndex)
		result.LastSnapshottedMilestoneIndex = uint32(snapshotInfo.SnapshotIndex)
	}

	// System time
	result.Time = time.Now().Unix() * 1000

	// Features
	// Workaround until https://github.com/golang/go/issues/27589 is fixed
	if len(features) != 0 {
		result.Features = features
	} else {
		result.Features = []string{}
	}

	// Tips
	result.Tips = metrics.SharedServerMetrics.TipsNonLazy.Load() + metrics.SharedServerMetrics.TipsSemiLazy.Load()

	// TX to request
	queued, pending, _ := gossip.RequestQueue().Size()
	result.TransactionsToRequest = queued + pending

	// Coo addr
	result.CoordinatorAddress = config.NodeConfig.GetString(config.CfgCoordinatorAddress)

	// node status
	result.Uptime = time.Since(nodeStartAt).Milliseconds()
	if !node.IsSkipped(autopeering.PLUGIN) {
		result.AutopeeringID = autopeering.ID
	}
	result.IsHealthy = tangleplugin.IsNodeHealthy()
	result.NodeAlias = config.NodeConfig.GetString(config.CfgNodeAlias)

	// Return node info
	c.JSON(http.StatusOK, result)
}

func getNodeAPIConfiguration(_ interface{}, c *gin.Context, _ <-chan struct{}) {

	result := GetNodeAPIConfigurationReturn{
		MaxFindTransactions: config.NodeConfig.GetInt(config.CfgWebAPILimitsMaxFindTransactions),
		MaxRequestsList:     config.NodeConfig.GetInt(config.CfgWebAPILimitsMaxRequestsList),
		MaxGetTrytes:        config.NodeConfig.GetInt(config.CfgWebAPILimitsMaxGetTrytes),
		MaxBodyLength:       config.NodeConfig.GetInt(config.CfgWebAPILimitsMaxBodyLengthBytes),
	}

	// Milestone start index
	snapshotInfo := tangle.GetSnapshotInfo()
	if snapshotInfo != nil {
		result.MilestoneStartIndex = snapshotInfo.PruningIndex
	}

	c.JSON(http.StatusOK, result)
}

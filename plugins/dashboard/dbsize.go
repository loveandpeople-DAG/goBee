package dashboard

import (
	"encoding/json"
	"time"

	"github.com/loveandpeople-DAG/goHive/daemon"
	"github.com/loveandpeople-DAG/goHive/events"
	"github.com/loveandpeople-DAG/goHive/timeutil"

	"github.com/loveandpeople-DAG/goBee/pkg/model/tangle"
	"github.com/loveandpeople-DAG/goBee/pkg/shutdown"
	"github.com/loveandpeople-DAG/goBee/plugins/database"
)

var (
	lastDbCleanup       = &database.DatabaseCleanup{}
	cachedDbSizeMetrics []*DBSizeMetric
)

// DBSizeMetric represents database size metrics.
type DBSizeMetric struct {
	Tangle   int64
	Snapshot int64
	Spent    int64
	Time     time.Time
}

func (s *DBSizeMetric) MarshalJSON() ([]byte, error) {
	size := struct {
		Tangle   int64 `json:"tangle"`
		Snapshot int64 `json:"snapshot"`
		Spent    int64 `json:"spent"`
		Time     int64 `json:"ts"`
	}{
		Tangle:   s.Tangle,
		Snapshot: s.Snapshot,
		Spent:    s.Spent,
		Time:     s.Time.Unix(),
	}

	return json.Marshal(size)
}

func currentDatabaseSize() *DBSizeMetric {
	tangle, snapshot, spent := tangle.GetDatabaseSizes()
	newValue := &DBSizeMetric{
		Tangle:   tangle,
		Snapshot: snapshot,
		Spent:    spent,
		Time:     time.Now(),
	}
	cachedDbSizeMetrics = append(cachedDbSizeMetrics, newValue)
	if len(cachedDbSizeMetrics) > 600 {
		cachedDbSizeMetrics = cachedDbSizeMetrics[len(cachedDbSizeMetrics)-600:]
	}
	return newValue
}

func runDatabaseSizeCollector() {

	// Gather first metric so we have a starting point
	currentDatabaseSize()

	onDatabaseCleanup := events.NewClosure(func(cleanup *database.DatabaseCleanup) {
		lastDbCleanup = cleanup
		hub.BroadcastMsg(&Msg{Type: MsgTypeDatabaseCleanupEvent, Data: cleanup})
	})

	daemon.BackgroundWorker("Dashboard[DBSize]", func(shutdownSignal <-chan struct{}) {
		database.Events.DatabaseCleanup.Attach(onDatabaseCleanup)
		defer database.Events.DatabaseCleanup.Detach(onDatabaseCleanup)

		timeutil.Ticker(func() {
			dbSizeMetric := currentDatabaseSize()
			hub.BroadcastMsg(&Msg{Type: MsgTypeDatabaseSizeMetric, Data: []*DBSizeMetric{dbSizeMetric}})
		}, 1*time.Minute, shutdownSignal)
	}, shutdown.PriorityDashboard)
}

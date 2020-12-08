package tangle

import (
	"github.com/loveandpeople-DAG/goHive/events"
)

var Events = packageEvents{
	ReceivedValidMilestone:   events.NewEvent(BundleCaller),
	ReceivedInvalidMilestone: events.NewEvent(events.ErrorCaller),
	AddressSpent:             events.NewEvent(events.StringCaller),
}

type packageEvents struct {
	ReceivedValidMilestone   *events.Event
	ReceivedInvalidMilestone *events.Event
	AddressSpent             *events.Event
}

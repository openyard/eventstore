package edge

import (
	"github.com/openyard/eventstore/internal/app/eventstore/domain"
	"github.com/openyard/eventstore/pkg/genproto/grpcapi"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func api2domain(events []*grpcapi.Event) []*domain.Event {
	res := make([]*domain.Event, 0)
	for _, e := range events {
		res = append(res, domain.NewEventAt(
			e.ID,
			e.Name,
			e.AggregateID,
			e.OccurredAt.AsTime(),
			e.Payload,
		))
	}
	return res
}

func domain2api(streamName string, events map[uint64]*domain.Event) *grpcapi.Stream {
	res := &grpcapi.Stream{Name: streamName, Version: uint64(len(events)), Events: make([]*grpcapi.Event, 0, len(events))}
	for p, e := range events {
		res.Events = append(res.Events, &grpcapi.Event{
			ID:          e.ID(),
			Name:        e.Name(),
			AggregateID: e.AggregateID(),
			Pos:         p + 1,
			Payload:     e.Payload(),
			OccurredAt:  timestamppb.New(e.OccurredAt()),
		})
	}
	return res
}

package domain

import (
	"context"
	"time"
)

const (
	// domain/version.name
	appendToStreamCommandName      = "event-store/v1.append"
	readStreamCommandName          = "event-store/v1.read"
	readStreamAtCommandName        = "event-store/v1.readAt"
	subscribeCommandName           = "event-store/v1.subscribe"
	subscribeWithIDCommandName     = "event-store/v1.subscribeWithID"
	subscribeWithOffsetCommandName = "event-store/v1.subscribeWithOffset"
)

const (
	UnknownCmd CommandKind = iota
	AppendCmd
	ReadCmd
	ReadAtCmd
	SubscribeCmd
	SubscribeWithIDCmd
	SubscribeWithOffsetCmd
)

type CommandKind uint8

func NewCommand(ctx context.Context, kind CommandKind, payload any) Command {
	return Command{ctx, kind, payload}
}

type Command struct {
	ctx     context.Context
	kind    CommandKind
	payload any
}

func Append(streamData ...StreamData) AppendCommand {
	return AppendCommand{streamData: streamData}
}

type AppendCommand struct {
	streamData []StreamData
}

type StreamData struct {
	name            string
	expectedVersion uint64
	events          []*Event
}

func NewStreamData(name string, expectedVersion uint64, events ...*Event) StreamData {
	return StreamData{
		name:            name,
		expectedVersion: expectedVersion,
		events:          events,
	}
}

func Read(streams ...string) ReadCommand {
	return ReadCommand{
		streams: streams,
	}
}

type ReadCommand struct {
	streams []string
}

func ReadAt(at time.Time, streams ...string) ReadAtCommand {
	return ReadAtCommand{
		streams: streams,
		at:      at,
	}
}

type ReadAtCommand struct {
	streams []string
	at      time.Time
}

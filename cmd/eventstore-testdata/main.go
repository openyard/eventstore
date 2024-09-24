package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/bwmarrin/snowflake"
	"github.com/openyard/eventstore/pkg/genproto/grpcapi"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
	"os"
	"time"
)

const threshold = 100000 // threshold to print processing output
var (
	version  = "TRUNK"
	revision = "HEAD"
)

func main() {
	start := time.Now().UTC()

	printv := flag.Bool("version", false, "print version")
	verbose := flag.Bool("v", false, "verbose output")
	help := flag.Bool("h", false, "print usage")
	streams := flag.Int("s", 100, "quantity how many streams are generated (optional)")
	events := flag.Int("e", 10, "quantity how many events per stream are generated (optional)")
	file := flag.String("f", "", "filename to save the generated stream data to (optional)")

	flag.Parse()

	if *help {
		Usage()
		os.Exit(0)
	}

	if *printv {
		fmt.Printf("iban4go %s - Rev. %s\n", version, revision)
		os.Exit(0)
	}

	var done chan bool
	if *streams**events > threshold {
		go func() {
			for {
				select {
				case <-time.After(time.Second):
					fmt.Print(".")
				case <-done:
					break
				}
			}
		}()
	}

	node, _ := snowflake.NewNode(1)
	td := &grpcapi.AppendRequest{StreamData: make([]*grpcapi.StreamData, 0)}
	for i := 0; i < *streams; i++ {
		td.StreamData = append(td.StreamData, generateStream(node, i, *events))
	}

	go func() {
		done <- true
	}()

	if *streams**events > threshold {
		fmt.Println("")
	}

	if *file == "" {
		raw, _ := protojson.Marshal(td)
		fmt.Printf("%s\n", raw)
		if *verbose {
			fmt.Printf("finished after %s\n", time.Since(start))
		}
		os.Exit(0)
	}
	f, err := os.OpenFile(*file, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("could not create result file: %s\n", err.Error())
		os.Exit(1)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	defer w.Flush()

	raw, _ := json.Marshal(td)
	w.WriteString(fmt.Sprintf("%s\n", raw))

	fmt.Printf("writing results to [%s]\n", *file)
	if *verbose {
		fmt.Printf("finished after %s\n", time.Since(start))
	}
}

var Usage = func() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()
}

func generateStream(node *snowflake.Node, i, j int) *grpcapi.StreamData {
	streamName := fmt.Sprintf("test-stream-%d", i+1)
	return &grpcapi.StreamData{
		Name:            streamName,
		ExpectedVersion: 0,
		Events:          generateEvents(j, node, streamName),
	}
}

func generateEvents(count int, node *snowflake.Node, aggregateID string) []*grpcapi.Event {
	events := make([]*grpcapi.Event, 0, count)
	for j := 0; j < count; j++ {
		events = append(events, &grpcapi.Event{
			ID:          node.Generate().String(),
			Name:        fmt.Sprintf("v1/test-event"),
			AggregateID: aggregateID,
			Pos:         uint64(j + 1),
			Payload:     nil,
			OccurredAt:  timestamppb.New(time.Now()),
		})
	}
	return events
}

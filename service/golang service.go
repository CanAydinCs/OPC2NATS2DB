package main

import (
	"context"
	"fmt"
	"log"

	"github.com/nats-io/nats.go"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
)

func main() {
	connectAndReadOPCUAAndPublish("ns=2;i=3")
	connectAndReadOPCUAAndPublish("ns=2;i=2")
}

func connectAndReadOPCUAAndPublish(inputString string) {
	opcServerURL := "opc.tcp://192.168.56.1:4840"

	// Connect to OPC UA server
	ctx := context.Background()
	_, err := opcua.GetEndpoints(ctx, opcServerURL)
	if err != nil {
		log.Fatal(err)
	}

	c, _ := opcua.NewClient(opcServerURL, opcua.SecurityMode(ua.MessageSecurityModeNone))
	err = c.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close(context.Background())

	// Create node id
	id, err := ua.ParseNodeID(inputString)
	if err != nil {
		log.Fatal(err)
	}

	// Read value from node
	req := &ua.ReadRequest{
		MaxAge:             2000,
		NodesToRead:        []*ua.ReadValueID{{NodeID: id, AttributeID: ua.AttributeIDValue}},
		TimestampsToReturn: ua.TimestampsToReturnBoth,
	}
	resp, err := c.Read(ctx, req)
	if err != nil {
		log.Fatal(err)
	}

	if len(resp.Results) == 0 || resp.Results[0].Status != ua.StatusOK {
		log.Fatal("Failed to read value")
	}

	fmt.Println("Read value:", resp.Results[0].Value.Value())

	// Connect to a NATS server
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	// Publish a message
	value := resp.Results[0].Value.Value()
	if valueStr, ok := value.(string); ok {
		// Convert the string to bytes
		valueBytes := []byte(valueStr)
		nc.Publish("test.subject", valueBytes)
	} else if valueInt, ok := value.(int64); ok {
		// Convert the int64 to bytes
		valueBytes := []byte(fmt.Sprintf("%d", valueInt))
		nc.Publish("test.subject", valueBytes)
	} else {
		log.Fatalf("Unexpected value type: %T", value)
	}

	fmt.Println("Published message", resp.Results[0].Value.Value(), "to test.subject")
}

package main

import (
	"context"
	"fmt"
	"log"

	"bufio"
	"os"

	"github.com/nats-io/nats.go"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
)

var OPC_IP_ADRESS string = "localhost"
var NATS_IP_ADRESS string = nats.DefaultURL
var NATS_SUBJECT string = "test.subject"
var CONFIG_PATH string = "../my_config.txt"

func main() {
	fmt.Println("------------------------")
	//configs
	//OPC
	var errOPCIP error
	OPC_IP_ADRESS, errOPCIP = readNthLine(CONFIG_PATH, 1, OPC_IP_ADRESS)
	if errOPCIP != nil {
		fmt.Println("No IP config found for OPC, launching at local IP")
	}
	fmt.Println("Reading OPC IP adress at:", OPC_IP_ADRESS)
	fmt.Println()

	//NATS
	var errNATSIP error
	NATS_IP_ADRESS, errNATSIP = readNthLine(CONFIG_PATH, 2, NATS_IP_ADRESS)
	if errNATSIP != nil {
		fmt.Println("No IP config found for NATS, launching at default IP")
	}
	fmt.Println("Publishing NATS Server at:", NATS_IP_ADRESS)
	fmt.Println()

	//NATS SUBJECT
	var errNATSSUBJECT error
	NATS_SUBJECT, errNATSSUBJECT = readNthLine(CONFIG_PATH, 3, NATS_SUBJECT)
	if errNATSSUBJECT != nil {
		fmt.Println("No subject has been found, set as default")
	}
	fmt.Println("NATS Server subject set as:", NATS_SUBJECT)
	fmt.Println("------------------------")
	//start
	connectAndReadOPCUAAndPublish("ns=2;i=3")
	connectAndReadOPCUAAndPublish("ns=2;i=2")
}

func connectAndReadOPCUAAndPublish(inputString string) {
	opcServerURL := "opc.tcp://" + OPC_IP_ADRESS

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
	nc, err := nats.Connect(NATS_IP_ADRESS)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	// Publish a message
	value := resp.Results[0].Value.Value()
	if valueStr, ok := value.(string); ok {
		// Convert the string to bytes
		valueBytes := []byte(valueStr)
		nc.Publish(NATS_SUBJECT, valueBytes)
	} else if valueInt, ok := value.(int64); ok {
		// Convert the int64 to bytes
		valueBytes := []byte(fmt.Sprintf("%d", valueInt))
		nc.Publish(NATS_SUBJECT, valueBytes)
	} else {
		log.Fatalf("Unexpected value type: %T", value)
	}

	fmt.Println("Published message", resp.Results[0].Value.Value())
	fmt.Println("------------------------")
}

func readNthLine(filename string, lineNumber int, defaultValue string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return defaultValue, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineCount := 0
	for scanner.Scan() {
		lineCount++
		if lineCount == lineNumber {
			subject := scanner.Text()
			if subject != "" {
				return subject, nil
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err // Return an error if there was a problem scanning the file
	}

	// Return the default value if the requested line does not exist
	return defaultValue, fmt.Errorf("line %d does not exist", lineNumber)
}

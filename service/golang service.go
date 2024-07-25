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

var OPC_IP_ADRESS string = ""
var NATS_IP_ADRESS string = ""
var NATS_SUBJECT string
var CONFIG_PATH string = "../my_config.txt"

func main() {
	fmt.Println("------------------------")
	//configs
	//OPC
	OPC_IP_ADRESS, _ = readFirstLine(CONFIG_PATH)
	fmt.Println("Reading OPC IP adress at:", OPC_IP_ADRESS)
	fmt.Println()

	//NATS
	var err error
	NATS_IP_ADRESS, err = readSecondLine(CONFIG_PATH)
	if err != nil {
		fmt.Println("No url config found for NATS, launching at default url")
	}
	fmt.Println("Publishing NATS Server at:", NATS_IP_ADRESS)
	fmt.Println()

	//NATS SUBJECT
	var err2 error
	NATS_SUBJECT, err2 = readThirdLine(CONFIG_PATH)
	if err2 != nil {
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

func readFirstLine(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		return scanner.Text(), nil
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", err
}

func readSecondLine(filename string) (string, error) {
	file, err := os.Open(filename) // Open the file specified by filename
	if err != nil {
		return "", err // Return an error if there was a problem opening the file
	}
	defer file.Close() // Ensure the file will be closed when the function completes

	scanner := bufio.NewScanner(file) // Create a new scanner for the file
	lineCount := 0
	for scanner.Scan() {
		lineCount++
		// If we've read the first line, continue and read second line
		if lineCount == 2 {
			IP := scanner.Text() // Return the second line's text
			if IP != "" {
				return IP, nil
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err // Return an error if there was a problem scanning the file
	}

	return nats.DefaultURL, fmt.Errorf("file does not have a second line")
}

func readThirdLine(filename string) (string, error) {
	file, err := os.Open(filename) // Open the file specified by filename
	if err != nil {
		return "", err // Return an error if there was a problem opening the file
	}
	defer file.Close() // Ensure the file will be closed when the function completes

	scanner := bufio.NewScanner(file) // Create a new scanner for the file
	lineCount := 0
	for scanner.Scan() {
		lineCount++
		// If we've read the first two lines, continue and read third line
		if lineCount == 3 {
			subject := scanner.Text() // Return the third line's text
			if subject != "" {
				return subject, nil
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err // Return an error if there was a problem scanning the file
	}

	return "test.subject", fmt.Errorf("file does not have a third line")
}

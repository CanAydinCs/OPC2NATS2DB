package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/nats-io/nats.go"
)

var NATS_IP_ADRESS string = nats.DefaultURL
var NATS_SUBJECT string = "test.subject"
var CONFIG_PATH string = "../my_config.txt"

func main() {
	//NATS
	fmt.Println("------------------------")
	var err error
	NATS_IP_ADRESS, err = readNthLine(CONFIG_PATH, 2, NATS_IP_ADRESS)
	if err != nil {
		fmt.Println("No url config found for NATS, launching at default url")
	}
	fmt.Println("Receiving NATS Server at:", NATS_IP_ADRESS)
	fmt.Println()

	//NATS SUBJECT
	var err2 error
	NATS_SUBJECT, err2 = readNthLine(CONFIG_PATH, 3, NATS_SUBJECT)
	if err2 != nil {
		fmt.Println("No subject has been found, set as default value")
	}
	fmt.Println("NATS Server subject set as:", NATS_SUBJECT)
	fmt.Println("------------------------")

	getMessage(2)
}

func getMessage(expectedAmount int) {
	// Connect to a NATS server
	nc, err := nats.Connect(NATS_IP_ADRESS)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	// Use a WaitGroup to wait for messages to arrive
	var wg sync.WaitGroup
	wg.Add(expectedAmount) // Assuming you expect at least one message

	// Subscribe to NATS_SUBJECT
	if _, err := nc.Subscribe(NATS_SUBJECT, func(m *nats.Msg) {
		fmt.Printf("Received a message: %s\n", m.Data)
		wg.Done()
	}); err != nil {
		log.Fatal(err)
	}

	// Wait for messages to come in
	wg.Wait()
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

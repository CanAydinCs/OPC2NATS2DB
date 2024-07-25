package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/nats-io/nats.go"
)

var NATS_IP_ADRESS string
var NATS_SUBJECT string
var CONFIG_PATH string = "../my_config.txt"

func main() {
	//NATS
	fmt.Println("------------------------")
	var err error
	NATS_IP_ADRESS, err = readSecondLine(CONFIG_PATH)
	if err != nil {
		fmt.Println("No url config found for NATS, launching at default url")
	}
	fmt.Println("Receiving NATS Server at:", NATS_IP_ADRESS)
	fmt.Println()

	//NATS SUBJECT
	var err2 error
	NATS_SUBJECT, err2 = readThirdLine(CONFIG_PATH)
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

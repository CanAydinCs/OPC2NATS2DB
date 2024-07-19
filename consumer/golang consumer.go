package main

import (
	"fmt"
	"log"
	"sync"

	"github.com/nats-io/nats.go"
)

func main() {
	// Connect to a NATS server
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	// Use a WaitGroup to wait for messages to arrive
	var wg sync.WaitGroup
	wg.Add(2) // Assuming you expect at least one message

	// Subscribe to "test.subject"
	if _, err := nc.Subscribe("test.subject", func(m *nats.Msg) {
		fmt.Printf("Received a message: %s\n", m.Data)
		wg.Done()
	}); err != nil {
		log.Fatal(err)
	}

	// Wait for messages to come in
	wg.Wait()
}

package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/nats-io/nats.go"
)

var NATS_IP_ADRESS string = nats.DefaultURL
var NATS_SUBJECT string = "test.subject"
var CONFIG_PATH string = "../my_config.txt"
var DB_SERVER_IP string = "localhost"
var DB_SERVER_PORT string = "5432"
var DB_SERVER_SSL_MODE string = "disable"
var DB_USERNAME string = ""
var DB_PASSWORD string = ""
var DB_NAME string = ""

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
	err = nil
	NATS_SUBJECT, err = readNthLine(CONFIG_PATH, 3, NATS_SUBJECT)
	if err != nil {
		fmt.Println("No subject has been found, set as default value")
	}
	fmt.Println("NATS Server subject set as:", NATS_SUBJECT)

	// Database Server IP
	fmt.Println("------------------------")
	err = nil
	DB_SERVER_IP, err = readNthLine(CONFIG_PATH, 4, DB_SERVER_IP)
	if err != nil {
		fmt.Println("No IP config found for Database, connecting to default url")
	}
	fmt.Println("Database connected to:", DB_SERVER_IP)

	// Database Server Port
	err = nil
	DB_SERVER_PORT, err = readNthLine(CONFIG_PATH, 8, DB_SERVER_PORT)
	if err != nil {
		fmt.Println("No port config found for db connection. Using default port")
	}
	fmt.Println("Database connection port is:", DB_SERVER_PORT)

	// SSL Mode
	err = nil
	DB_SERVER_SSL_MODE, err = readNthLine(CONFIG_PATH, 9, DB_SERVER_SSL_MODE)
	if err != nil {
		fmt.Println("No SSL config was found. Using default mode.")
	}
	fmt.Println("SSL secure connection is:", DB_SERVER_SSL_MODE)

	// Database Username
	err = nil
	DB_USERNAME, err = readNthLine(CONFIG_PATH, 5, DB_USERNAME)
	if err != nil {
		fmt.Println("No username config was found to connect database.")
		fmt.Println("This config is fatal for application and cannot be provided default for security purposes.")
		log.Fatal("Ending the program...")
	}

	// Database User password
	err = nil
	DB_PASSWORD, err = readNthLine(CONFIG_PATH, 6, DB_PASSWORD)
	if err != nil {
		fmt.Println("No password config was found to connect database.")
		fmt.Println("This config is fatal for application and cannot be provided default for security purposes.")
		log.Fatal("Ending the program...")
	}

	// Database, Table name
	err = nil
	DB_NAME, err = readNthLine(CONFIG_PATH, 7, DB_NAME)
	if err != nil {
		fmt.Println("No database table name was found to connect database")
		fmt.Println("This config is fatal for application and cannot be provided default for security purposes.")
		log.Fatal("Ending the program...")
	}

	// Starting Application
	fmt.Println("------------------------")
	fmt.Println("Server started to listening")
	fmt.Println("------------------------")

	// Reading From NATS
	message1, message2, err := getMessage(2)
	if err != nil {
		log.Fatal("An error occurred while getting the messages: ", err)
	}
	fmt.Println("Messages are successfully read:", message1, message2)
	fmt.Println("------------------------")

	//Saving to DB
	saveMessageToDB(message1, message2)
	fmt.Printf("Temperature and Pressure successfully written in DB. Temperature:%s, pressure:%s\n", message1, message2)
	fmt.Println("------------------------")
}

func getMessage(expectedAmount int) (string, string, error) {
	// Connect to a NATS server
	nc, err := nats.Connect(NATS_IP_ADRESS)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	// Use a WaitGroup to wait for messages to arrive
	var wg sync.WaitGroup
	wg.Add(expectedAmount) // Assuming you expect at least one message

	// Variables to hold the received messages
	var message1, message2 string

	// Subscribe to NATS_SUBJECT
	sub, err := nc.Subscribe(NATS_SUBJECT, func(m *nats.Msg) {
		//it reads pressure first
		if message2 == "" {
			message2 = string(m.Data)
		} else {
			message1 = string(m.Data)
		}
		wg.Done()
	})
	if err != nil {
		log.Fatal(err)
	}
	defer sub.Unsubscribe()

	// Wait for messages to come in
	wg.Wait()

	// Close the connection to ensure all messages are processed before returning
	nc.Close()
	return message1, message2, nil
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

// Creates the database if it does not exist
func createDatabaseIfNotExists(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS messages (
			id SERIAL PRIMARY KEY,
			tempature TEXT,
			pressure TEXT,
			created_at TIMESTAMPTZ DEFAULT NOW()
		);
	`)
	return err
}

// Saves a message to the PostgreSQL TimescaleDB
func saveMessageToDB(message1 string, message2 string) {
	// DB connection string, change placeholders as needed
	connStr := "user=" + DB_USERNAME + " password=" + DB_PASSWORD + " dbname=" + DB_NAME +
		" host=" + DB_SERVER_IP + " port=" + DB_SERVER_PORT + " sslmode=" + DB_SERVER_SSL_MODE

	// Connect to the database
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		log.Fatal("Failed to open a DB connection: ", err)
	}
	defer db.Close()

	// Ensure the messages table exists
	if err := createDatabaseIfNotExists(db); err != nil {
		log.Fatal("Failed to create database table: ", err)
	}

	// Insert the message into the database
	_, err = db.Exec("INSERT INTO messages(tempature, pressure) VALUES($1, $2)", message1, message2)
	if err != nil {
		log.Fatal("Failed to insert message into the database: ", err)
	}
}

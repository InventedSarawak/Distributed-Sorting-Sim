package transport

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strconv"
	"time"

	"github.com/InventedSarawak/Distributed-Sorting-Sim/pkg/types"
)

const DefaultPort = 8000

// listen maintains your original structure but starts persistent handlers
func Listen[Message any](id int, messages chan Message) error {
	port := getCurrentPort(id)
	address := ":" + strconv.Itoa(port)

	var listener net.Listener
	var err error

	for range 5 {
		listener, err = net.Listen("tcp", address)
		if err == nil {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	if err != nil {
		return fmt.Errorf("failed to initialize port after retries: %w", err)
	}

	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("failed to accept connection: %v\n", err)
			continue
		}

		// This goroutine becomes the persistent Inbox Dispatcher
		go func(c net.Conn) {
			if err := HandleConnection(c, messages); err != nil {
				fmt.Printf("connection closed: %v\n", err)
			}
		}(conn)
	}
}

// HandleConnection is now the Inbox Dispatcher: it loops until the connection closes
func HandleConnection[Message any](conn net.Conn, messages chan Message) error {
	defer conn.Close()
	decoder := json.NewDecoder(conn)

	for {
		var message Message
		// Decode waits for the next JSON object in the persistent stream
		err := decoder.Decode(&message)
		if err != nil {
			if err == io.EOF {
				return nil // Connection closed gracefully
			}
			return fmt.Errorf("failed to decode stream: %w", err)
		}

		// Push to the node's internal inbox
		messages <- message

		// Send persistent acknowledgment
		_, err = conn.Write([]byte(types.MsgAck))
		if err != nil {
			return fmt.Errorf("failed to send acknowledgment: %w", err)
		}
	}
}

// SendMessage now uses a persistent connection instead of dialing every time
func SendMessage[Payload any](conn net.Conn, message types.Message[Payload]) error {
	if conn == nil {
		return fmt.Errorf("connection not established")
	}

	// Use an encoder to write directly to the persistent stream
	encoder := json.NewEncoder(conn)
	err := encoder.Encode(message)
	if err != nil {
		return fmt.Errorf("failed to encode message: %w", err)
	}

	// Wait for acknowledgment on the same stream
	ackBuffer := make([]byte, len(types.MsgAck))
	_, err = conn.Read(ackBuffer)
	if err != nil {
		return fmt.Errorf("failed to read acknowledgment: %w", err)
	}

	if string(ackBuffer) != string(types.MsgAck) {
		return fmt.Errorf("invalid acknowledgment received: %s", string(ackBuffer))
	}

	return nil
}

// DialNeighbor is used ONCE at the start to establish the persistent manager links
func DialNeighbor(targetID int) (net.Conn, error) {
	port := getCurrentPort(targetID)
	address := "localhost:" + strconv.Itoa(port)
	return dialWithRetry(address, 10)
}

func dialWithRetry(address string, maxRetries int) (net.Conn, error) {
	var conn net.Conn
	var err error
	backoff := 100 * time.Millisecond

	for range maxRetries {
		conn, err = net.DialTimeout("tcp", address, 2*time.Second)
		if err == nil {
			return conn, nil
		}
		time.Sleep(backoff)
		backoff *= 2
		backoff += time.Duration(10) * time.Millisecond
	}
	return nil, fmt.Errorf("after %d attempts, last error: %w", maxRetries, err)
}

func getCurrentPort(id int) int {
	return DefaultPort + id
}
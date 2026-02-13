package transport

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/InventedSarawak/Distributed-Sorting-Sim/pkg/types"
)

const DefaultPort = 8000

func listen[Message any](id int, messages chan Message) error {
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

		go func(c net.Conn) {
			if err := HandleConnection(c, messages); err != nil {
				fmt.Printf("error handling connection: %v\n", err)
			}
		}(conn)
	}
}

func HandleConnection[Message any](conn net.Conn, messages chan Message) error {
	defer conn.Close()

	decoder := json.NewDecoder(conn)

	var message Message
	err := decoder.Decode(&message)
	if err != nil {
		return fmt.Errorf("failed to decode message: %w", err)
	}

	messages <- message

	_, err = conn.Write([]byte(types.MsgAck))
	if err != nil {
		return fmt.Errorf("failed to send acknowledgment: %w", err)
	}

	return nil
}

func SendMessage[Payload any](message types.Message[Payload], receiverID int) error {
	port := getCurrentPort(receiverID)
	address := "localhost:" + strconv.Itoa(port)

	conn, err := dialWithRetry(address, 10)
	if err != nil {
		return err
	}
	defer conn.Close()

	messageBytes, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	_, err = conn.Write(messageBytes)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

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

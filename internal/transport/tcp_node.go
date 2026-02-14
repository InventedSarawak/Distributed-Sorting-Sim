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

func Listen[Message any](id int, messages chan Message) error {
	port := getCurrentPort(id)
	address := ":" + strconv.Itoa(port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go HandleConnection(conn, messages)
	}
}

func HandleConnection[Message any](conn net.Conn, messages chan Message) error {

	decoder := json.NewDecoder(conn)
	for {
		var message Message
		err := decoder.Decode(&message)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("decode error: %w", err)
		}
		messages <- message
	}
}

func SendMessage[Payload any](conn net.Conn, message types.Message[Payload]) error {
	if conn == nil {
		return fmt.Errorf("connection not established")
	}

	return json.NewEncoder(conn).Encode(message)
}

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
	}
	return nil, fmt.Errorf("failed after %d attempts: %w", maxRetries, err)
}

func getCurrentPort(id int) int {
	return DefaultPort + id
}

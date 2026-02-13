package types

import (
	"net"
	"time"
)

type Direction string

const (
	Left  Direction = "LEFT"
	Right Direction = "RIGHT"
)

type Position string

const (
	Head   Position = "HEAD"
	Tail   Position = "TAIL"
	Middle Position = "MIDDLE"
)

type MessageType string

const (
	MsgInit MessageType = "INIT"
	MsgData MessageType = "DATA"
	MsgSync MessageType = "SYNC"
	MsgTerm MessageType = "TERM"
	MsgAck  MessageType = "ACK"
)

type AlgorithmType int

const (
	OddEven AlgorithmType = iota
	Sasaki
	Alternate
)

type InputType int

const (
	Random InputType = iota
	Sorted
	Reverse
)

type Config struct {
	NodeCount uint
	InputType InputType
	Algorithm AlgorithmType
}

type Message[Payload any] struct {
	Round             int         `json:"round"`
	SenderID          int         `json:"sender_id"`
	Type              MessageType `json:"type"`
	Body              Payload     `json:"body"`
	ReceiverID        int         `json:"receiver_id"`
	IncomingDirection Direction   `json:"incoming_direction"`
	Timestamp         time.Time   `json:"timestamp"`
	Sequence          int         `json:"sequence"`
}

type Node[Payload any] struct {
	ID        int
	Position  Position
	TotalNode int
	Value     Payload
	Round     int

	LeftConn  net.Conn
	RightConn net.Conn

	LeftInbox  chan Message[Payload]
	RightInbox chan Message[Payload]
}

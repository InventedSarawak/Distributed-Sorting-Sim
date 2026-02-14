package simulator

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/InventedSarawak/Distributed-Sorting-Sim/internal/transport"
	"github.com/InventedSarawak/Distributed-Sorting-Sim/pkg/types"
)

type SimulatorEngine[T any] struct {
	TotalNodes  int
	ActiveNodes int32
	Done        chan bool
	WaitGroup   sync.WaitGroup
}

func NewEngine[T any](n int) *SimulatorEngine[T] {
	return &SimulatorEngine[T]{
		TotalNodes: n,
		Done:       make(chan bool),
	}
}

func (e *SimulatorEngine[T]) IncrementClock(n *types.Node[T]) {
	n.Round++
}

func (e *SimulatorEngine[T]) InitialSetup(n *types.Node[T], config types.Config, leftBuf, rightBuf *RoundBuffer[T], debug bool) error {
	if debug {
		fmt.Printf("[Setup] Node %d: Starting Discovery Phase...\n", n.ID)
	}

	total, err := DiscoverTotalNodes(n, leftBuf, rightBuf)
	if err != nil {
		return err
	}
	e.TotalNodes = total

	if debug {
		fmt.Printf("[Setup] Node %d: Discovery Complete. Total Nodes: %d\n", n.ID, total)
	}
	return nil
}

func SetupNode[T any](n *types.Node[T], debug bool) error {

	mainInbox := make(chan types.Message[T], 500)

	go func() {
		for msg := range mainInbox {
			if msg.Type == types.MsgSync {
				continue
			}
			if msg.SenderID < n.ID {
				select {
				case n.LeftInbox <- msg:
				default:
				}
			} else if msg.SenderID > n.ID {
				select {
				case n.RightInbox <- msg:
				default:
				}
			}
		}
	}()

	addr := fmt.Sprintf(":%d", 8000+n.ID)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen error: %w", err)
	}

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				continue
			}

			go func(c net.Conn) {
				dec := json.NewDecoder(c)
				var handshake types.Message[T]

				if err := dec.Decode(&handshake); err != nil {
					c.Close()
					return
				}

				switch handshake.SenderID {
				case n.ID - 1:
					n.LeftConn = c
					if debug {
						fmt.Printf("[Net] Node %d: Accepted LeftConn from %d\n", n.ID, handshake.SenderID)
					}
				case n.ID + 1:
					n.RightConn = c
					if debug {
						fmt.Printf("[Net] Node %d: Accepted RightConn from %d\n", n.ID, handshake.SenderID)
					}
				}

				for {
					var msg types.Message[T]
					if err := dec.Decode(&msg); err != nil {
						c.Close()
						return
					}
					mainInbox <- msg
				}
			}(conn)
		}
	}()

	if n.Position != types.Tail {
		targetID := n.ID + 1
		conn, err := transport.DialNeighbor(targetID)
		if err != nil {
			return err
		}
		n.RightConn = conn
		if debug {
			fmt.Printf("[Net] Node %d: Connected to Right Neighbor %d\n", n.ID, targetID)
		}

		handshake := types.Message[T]{Type: types.MsgSync, SenderID: n.ID}
		json.NewEncoder(conn).Encode(handshake)

		go func(c net.Conn) {
			dec := json.NewDecoder(c)
			for {
				var msg types.Message[T]
				if err := dec.Decode(&msg); err != nil {
					c.Close()
					return
				}
				mainInbox <- msg
			}
		}(conn)
	}

	return nil
}

func (e *SimulatorEngine[T]) SignalStable() {
	if atomic.AddInt32(&e.ActiveNodes, -1) == 0 {
		e.Done <- true
	}
}

func (e *SimulatorEngine[T]) ResetStability() {
	atomic.StoreInt32(&e.ActiveNodes, int32(e.TotalNodes))
}

func (e *SimulatorEngine[T]) CheckTermination() bool {
	select {
	case <-e.Done:
		return true
	default:
		return false
	}
}

func DiscoverTotalNodes[T any](n *types.Node[T], leftBuf, rightBuf *RoundBuffer[T]) (int, error) {
	leftDist, rightDist := -1, -1

	time.Sleep(200 * time.Millisecond)

	sendSeed := func(conn net.Conn, dist uint64) error {
		msg := types.Message[T]{Type: types.MsgInit, SenderID: n.ID, Round: 0, Sequence: dist}
		return transport.SendMessage(conn, msg)
	}

	if n.Position == types.Head {
		leftDist = 0
		if n.RightConn == nil {
			return -1, fmt.Errorf("head rightconn nil")
		}
		if err := sendSeed(n.RightConn, uint64(leftDist)); err != nil {
			return -1, err
		}
	}
	if n.Position == types.Tail {
		rightDist = 0
		if n.LeftConn == nil {
			return -1, fmt.Errorf("tail leftconn nil")
		}
		if err := sendSeed(n.LeftConn, uint64(rightDist)); err != nil {
			return -1, err
		}
	}

	for leftDist == -1 || rightDist == -1 {
		if leftDist == -1 {
			msg := leftBuf.GetStepMessage(0)
			leftDist = int(msg.Sequence) + 1
			if n.Position != types.Tail {
				msg.Sequence = uint64(leftDist)
				transport.SendMessage(n.RightConn, msg)
			}
		}
		if rightDist == -1 {
			msg := rightBuf.GetStepMessage(0)
			rightDist = int(msg.Sequence) + 1
			if n.Position != types.Head {
				msg.Sequence = uint64(rightDist)
				transport.SendMessage(n.LeftConn, msg)
			}
		}
	}

	n.TotalNode = leftDist + rightDist + 1
	return n.TotalNode, nil
}

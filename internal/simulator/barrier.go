package simulator

import (
	"sync"

	"github.com/InventedSarawak/Distributed-Sorting-Sim/internal/transport"
	"github.com/InventedSarawak/Distributed-Sorting-Sim/pkg/types"
)

type RoundBuffer[T any] struct {
	mu         sync.Mutex
	futureMsgs map[int]types.Message[T]
	inbox      chan types.Message[T]
}

func NewRoundBuffer[T any](inbox chan types.Message[T]) *RoundBuffer[T] {
	return &RoundBuffer[T]{
		futureMsgs: make(map[int]types.Message[T]),
		inbox:      inbox,
	}
}

func (rb *RoundBuffer[T]) GetStepMessage(targetRound int) types.Message[T] {
	for {
		rb.mu.Lock()

		if msg, found := rb.futureMsgs[targetRound]; found {
			delete(rb.futureMsgs, targetRound)
			rb.mu.Unlock()
			return msg
		}
		rb.mu.Unlock()

		msg := <-rb.inbox

		if msg.Round == targetRound {
			return msg
		}

		rb.mu.Lock()
		rb.futureMsgs[msg.Round] = msg
		rb.mu.Unlock()
	}
}

func WaitForNeighbors[T any](n *types.Node[T], currentRound int, leftBuf, rightBuf *RoundBuffer[T]) (leftMsg, rightMsg *types.Message[T]) {
	var wg sync.WaitGroup

	if n.Position != types.Head {
		wg.Add(1)
		go func() {
			defer wg.Done()
			msg := leftBuf.GetStepMessage(currentRound)
			leftMsg = &msg
		}()
	}

	if n.Position != types.Tail {
		wg.Add(1)
		go func() {
			defer wg.Done()
			msg := rightBuf.GetStepMessage(currentRound)
			rightMsg = &msg
		}()
	}

	wg.Wait()
	return
}

func BroadcastTermination[T any](n *types.Node[T]) {
	termMsg := types.Message[T]{
		Type:     types.MsgTerm,
		SenderID: n.ID,
		Round:    n.Round,
	}

	if n.Position != types.Tail {

		_ = transport.SendMessage(n.RightConn, termMsg)
	}
	if n.Position != types.Head {

		_ = transport.SendMessage(n.LeftConn, termMsg)
	}
}

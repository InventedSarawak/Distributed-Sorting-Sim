package shared

import (
	"math/rand"
	"net"
	"time"

	"github.com/InventedSarawak/Distributed-Sorting-Sim/pkg/types"
)

// DiscoverTotalNodes implements the bidirectional increment method.
// It avoids circular dependencies by using functional parameters for transport and buffering.
func DiscoverTotalNodes[T any](
	n *types.Node[T],
	sendFunc func(net.Conn, types.Message[T]) error,
	getLeftMsg func(int) types.Message[T],
	getRightMsg func(int) types.Message[T],
) int {
	leftDist := -1
	rightDist := -1

	// Helper to handle seed sending and avoid repeated code blocks.
	seedBoundary := func(pos types.Position, conn net.Conn) {
		if n.Position == pos {
			val := 0
			// Fix: Cast 'val' to uint64 to match the types.Message field.
			msg := types.Message[T]{
				Type:     types.MsgInit,
				Round:    0,
				Sequence: uint64(val),
				SenderID: n.ID,
			}
			if conn != nil {
				_ = sendFunc(conn, msg)
			}
		}
	}

	// 1. Initial seeds: Head starts Left count, Tail starts Right count.
	if n.Position == types.Head {
		leftDist = 0
		seedBoundary(types.Head, n.RightConn)
	}
	if n.Position == types.Tail {
		rightDist = 0
		seedBoundary(types.Tail, n.LeftConn)
	}

	// 2. Propagation logic: Loop until both distances reach the node.
	for leftDist == -1 || rightDist == -1 {
		if leftDist == -1 && n.Position != types.Head {
			msg := getLeftMsg(0) // Wait for distance from the left neighbor
			leftDist = int(msg.Sequence) + 1
			if n.Position != types.Tail && n.RightConn != nil {
				msg.Sequence = uint64(leftDist)
				_ = sendFunc(n.RightConn, msg)
			}
		}

		if rightDist == -1 && n.Position != types.Tail {
			msg := getRightMsg(0) // Wait for distance from the right neighbor
			rightDist = int(msg.Sequence) + 1
			if n.Position != types.Head && n.LeftConn != nil {
				msg.Sequence = uint64(rightDist)
				_ = sendFunc(n.LeftConn, msg)
			}
		}
	}

	n.TotalNode = leftDist + rightDist + 1
	return n.TotalNode
}

// GenerateInitialValue centralizes element generation for all algorithms.
func GenerateInitialValue[T any](n *types.Node[T], config types.Config) {
	var val int
	switch config.InputType {
	case types.Sorted:
		val = n.ID
	case types.Reverse:
		val = int(config.NodeCount) - 1 - n.ID
	case types.Random:
		r := rand.New(rand.NewSource(time.Now().UnixNano() + int64(n.ID)))
		val = r.Intn(1000)
	}

	// Fix: Use type assertion to assign the generated integer to the generic Node.Value.
	n.Value = any(val).(T)
}

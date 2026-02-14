package shared

import (
	"math/rand"
	"net"
	"time"

	"github.com/InventedSarawak/Distributed-Sorting-Sim/pkg/types"
)

// GenerateInitialValue returns a raw integer based on the config.
func GenerateInitialValue(id int, config types.Config) int {
	var val int
	switch config.InputType {
	case types.Sorted:
		val = id
	case types.Reverse:
		val = int(config.NodeCount) - 1 - id
	case types.Random:
		r := rand.New(rand.NewSource(time.Now().UnixNano() + int64(id)))
		val = r.Intn(1000)
	}
	return val
}

// Stub for compatibility
func DiscoverTotalNodes[T any](
	n *types.Node[T],
	sendFunc func(net.Conn, types.Message[T]) error,
	getLeftMsg func(int) types.Message[T],
	getRightMsg func(int) types.Message[T],
) int {
	return 0
}

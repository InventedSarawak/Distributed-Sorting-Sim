package transport

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/InventedSarawak/Distributed-Sorting-Sim/pkg/types"
)

func TestTCPNodeCommunication(t *testing.T) {
	type TestPayload struct {
		Data string `json:"data"`
	}

	var wg sync.WaitGroup

	messages := make(chan types.Message[TestPayload], 10)

	go func() {
		err := listen(1, messages)
		if err != nil {
			t.Logf("Error starting listener: %v", err)
		}
	}()

	time.Sleep(1 * time.Second)

	testMessage := types.Message[TestPayload]{
		Round:      1,
		SenderID:   0,
		Type:       types.MsgData,
		Body:       TestPayload{Data: "Hello, Node!"},
		ReceiverID: 1,
	}

	wg.Go(func() {
		err := SendMessage(testMessage, 1)
		if err != nil {
			t.Errorf("Error sending message: %v", err)
		}
	})

	select {
	case receivedMessage := <-messages:
		if receivedMessage.SenderID != testMessage.SenderID {
			t.Errorf("Expected SenderID %d, got %d", testMessage.SenderID, receivedMessage.SenderID)
		}
		if receivedMessage.Body.Data != testMessage.Body.Data {
			t.Errorf("Expected Body %s, got %s", testMessage.Body.Data, receivedMessage.Body.Data)
		}
	case <-time.After(5 * time.Second):
		t.Error("Timed out waiting for message")
	}

	wg.Wait()

	t.Logf("Successfully sent and received message: %s", testMessage.Body.Data)
}

func TestBidirectionalTraffic(t *testing.T) {
	type TestPayload struct {
		Data string `json:"data"`
	}

	var wg sync.WaitGroup

	messages1 := make(chan types.Message[TestPayload], 100)
	messages2 := make(chan types.Message[TestPayload], 100)

	go func() {
		err := listen(1, messages1)
		if err != nil {
			t.Logf("Error starting listener 1: %v", err)
		}
	}()

	go func() {
		err := listen(2, messages2)
		if err != nil {
			t.Logf("Error starting listener 2: %v", err)
		}
	}()

	time.Sleep(1 * time.Second)

	wg.Go(func() {
		for i := range 3 {
			msg := types.Message[TestPayload]{
				SenderID:   1,
				ReceiverID: 2,
				Body:       TestPayload{Data: fmt.Sprintf("Msg from 1 to 2: %d", i)},
			}
			if err := SendMessage(msg, 2); err != nil {
				t.Errorf("Node 1 failed to send message %d: %v", i, err)
			}
		}
	})

	wg.Go(func() {
		for i := range 4 {
			msg := types.Message[TestPayload]{
				SenderID:   2,
				ReceiverID: 1,
				Body:       TestPayload{Data: fmt.Sprintf("Msg from 2 to 1: %d", i)},
			}
			if err := SendMessage(msg, 1); err != nil {
				t.Errorf("Node 2 failed to send message %d: %v", i, err)
			}
		}
	})

	wg.Wait()

	for i := range 3 {
		select {
		case <-messages2:
		case <-time.After(1 * time.Second):
			t.Errorf("Timeout waiting for message %d on Node 2", i)
		}
	}

	for i := range 4 {
		select {
		case <-messages1:
		case <-time.After(1 * time.Second):
			t.Errorf("Timeout waiting for message %d on Node 1", i)
		}
	}

	t.Logf("Successfully sent and received bidirectional messages between Node 1 and Node 2")
}

func TestHeavyLoad(t *testing.T) {
	type TestPayload struct {
		Iteration int `json:"iteration"`
	}

	const numNodes = 1000
	const messagesPerNode = 3
	var wg sync.WaitGroup

	inboxes := make([]chan types.Message[TestPayload], numNodes)
	for i := range numNodes {
		inboxes[i] = make(chan types.Message[TestPayload], messagesPerNode)
	}

	for i := range numNodes {
		nodeID := i
		go func() {
			listen(nodeID, inboxes[nodeID])
		}()
	}

	time.Sleep(2 * time.Second)

	wg.Add(numNodes)
	for i := range numNodes {
		nodeID := i
		go func() {
			defer wg.Done()

			for m := range messagesPerNode {
				if nodeID == 0 {
					msg := types.Message[TestPayload]{
						SenderID:   nodeID,
						ReceiverID: nodeID + 1,
						Round:      m,
						Body:       TestPayload{Iteration: m},
						Type:       types.MsgData,
					}
					time.Sleep(10 * time.Millisecond)
					if err := SendMessage(msg, nodeID+1); err != nil {
						t.Errorf("Head failed to send: %v", err)
					}
				} else {
					select {
					case received := <-inboxes[nodeID]:
						if nodeID < numNodes-1 {
							forwardMsg := types.Message[TestPayload]{
								SenderID:   nodeID,
								ReceiverID: nodeID + 1,
								Round:      received.Round,
								Body:       received.Body,
								Type:       types.MsgData,
							}
							if err := SendMessage(forwardMsg, nodeID+1); err != nil {
								t.Errorf("Node %d failed to forward: %v", nodeID, err)
							}
						}
					case <-time.After(10 * time.Second):
						t.Errorf("Node %d timed out waiting for iteration %d", nodeID, m)
						return
					}
				}
			}
		}()
	}

	wg.Wait()
	t.Logf("Successfully processed %d nodes with %d messages each", numNodes, messagesPerNode)
}

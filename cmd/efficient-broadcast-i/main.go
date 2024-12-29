package main

import (
	"encoding/json"
	"fmt"
	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
	"log"
	"sync"
	"time"
)

var (
	node          *maelstrom.Node
	neighbors     []chan float64
	messages      []float64
	topologyReady sync.WaitGroup
	seen          map[float64]struct{}
	lock          sync.Mutex
)

type BroadcastMessage struct {
	Type    string  `json:"type"`
	Message float64 `json:"message"`
}

func handleBroadcast(msg maelstrom.Message) error {
	var body BroadcastMessage
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	lock.Lock()
	defer lock.Unlock()

	message := body.Message
	if _, ok := seen[message]; ok {
		return nil
	}

	seen[message] = struct{}{}
	messages = append(messages, message)

	topologyReady.Wait()
	for _, c := range neighbors {
		c <- message
	}

	return node.Reply(msg, map[string]any{"type": "broadcast_ok"})
}

func handleBroadcastTo(nodeID string, c chan float64) {
	const messageType = "broadcast"

	if nodeID == node.ID() {
		return
	}

	for {
		message, ok := <-c
		if !ok {
			fmt.Printf("neighbors channel closed")
			return
		}

		payload := BroadcastMessage{
			Type:    messageType,
			Message: message,
		}

		for {
			var err error
			if err = node.RPC(nodeID, payload, func(msg maelstrom.Message) error { return nil }); err == nil {
				break
			}

			fmt.Printf("got error while sending RPC request: %s\n", err.Error())
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func handleRead(msg maelstrom.Message) error {
	return node.Reply(msg, map[string]any{
		"type":     "read_ok",
		"messages": messages,
	})
}

func handleTopology(msg maelstrom.Message) error {
	defer topologyReady.Done()

	type topologyMsg struct {
		Topology map[string][]string `json:"topology"`
	}

	var topology topologyMsg
	_ = json.Unmarshal(msg.Body, &topology)

	lock.Lock()
	defer lock.Unlock()

	for _, id := range topology.Topology[node.ID()] {
		c := make(chan float64)
		neighbors = append(neighbors, c)
		go handleBroadcastTo(id, c)
	}

	return node.Reply(msg, map[string]any{"type": "topology_ok"})
}

func main() {
	messages = make([]float64, 0)
	neighbors = make([]chan float64, 0)
	seen = make(map[float64]struct{})
	topologyReady.Add(1)
	node = maelstrom.NewNode()

	node.Handle("broadcast", handleBroadcast)
	node.Handle("read", handleRead)
	node.Handle("topology", handleTopology)

	if err := node.Run(); err != nil {
		log.Fatal(err)
	}
}

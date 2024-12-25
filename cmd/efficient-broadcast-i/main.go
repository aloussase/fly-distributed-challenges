package main

import (
	"encoding/json"
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

	if _, ok := seen[body.Message]; ok {
		return nil
	}

	seen[body.Message] = struct{}{}
	messages = append(messages, body.Message)

	topologyReady.Wait()
	for _, c := range neighbors {
		go (func(c chan float64, msg float64) {
			c <- msg
		})(c, body.Message)
	}

	response := make(map[string]any)
	response["type"] = "broadcast_ok"

	return node.Reply(msg, response)
}

func handleBroadcastTo(nodeID string, c chan float64) {
	const messageType = "broadcast"

	if nodeID == node.ID() {
		return
	}

	for {
		message := <-c

		payload := BroadcastMessage{
			Type:    messageType,
			Message: message,
		}

		for {
			var err error
			if err = node.RPC(nodeID, payload, func(msg maelstrom.Message) error { return nil }); err == nil {
				break
			}

			time.Sleep(100 * time.Millisecond)
		}
	}
}

func handleRead(msg maelstrom.Message) error {
	body := make(map[string]any, 2)
	body["type"] = "read_ok"
	body["messages"] = messages
	return node.Reply(msg, body)
}

func handleTopology(msg maelstrom.Message) error {
	type topologyMsg struct {
		Topology map[string][]string `json:"topology"`
	}

	var topology topologyMsg
	_ = json.Unmarshal(msg.Body, &topology)

	for _, id := range topology.Topology[node.ID()] {
		c := make(chan float64)
		neighbors = append(neighbors, c)
		go handleBroadcastTo(id, c)
	}

	topologyReady.Done()

	response := make(map[string]any, 1)
	response["type"] = "topology_ok"

	return node.Reply(msg, response)
}

func main() {
	messages = make([]float64, 0, 10)
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

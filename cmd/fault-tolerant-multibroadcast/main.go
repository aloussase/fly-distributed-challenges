package main

import (
	"encoding/json"
	"fmt"
	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
	"log"
	"time"
)

var (
	node      *maelstrom.Node
	neighbors map[string]chan float64
	messages  []float64
)

func handleBroadcast(msg maelstrom.Message) error {
	var body map[string]any
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	message := body["message"].(float64)
	messages = append(messages, message)

	delete(body, "message")
	body["type"] = "broadcast_ok"

	for _, c := range neighbors {
		go (func(c chan float64) {
			c <- message
		})(c)
	}

	return node.Reply(msg, body)
}

func handleBroadcastTo(nodeID string) {
	c := neighbors[nodeID]

	payload := make(map[string]any, 2)
	payload["type"] = "from_broadcast"

	for {
		message := <-c
		payload["message"] = message

		for {
			var err error
			if err = node.RPC(nodeID, payload, func(msg maelstrom.Message) error { return nil }); err == nil {
				break
			}

			fmt.Printf("error sending message to node %s: %s", nodeID, err.Error())
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func handleFromBroadcast(msg maelstrom.Message) error {
	var body map[string]any
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	message := body["message"].(float64)
	messages = append(messages, message)

	response := make(map[string]any, 1)
	response["type"] = "from_broadcast_ok"

	return node.Reply(msg, response)
}

func handleRead(msg maelstrom.Message) error {
	body := make(map[string]any, 2)
	body["type"] = "read_ok"
	body["messages"] = messages
	return node.Reply(msg, body)
}

func handleTopology(msg maelstrom.Message) error {
	for _, id := range node.NodeIDs() {
		neighbors[id] = make(chan float64)
		go handleBroadcastTo(id)
	}

	response := make(map[string]any, 1)
	response["type"] = "topology_ok"

	return node.Reply(msg, response)
}

func main() {
	messages = make([]float64, 0, 10)
	neighbors = make(map[string]chan float64)
	node = maelstrom.NewNode()

	node.Handle("broadcast", handleBroadcast)
	node.Handle("read", handleRead)
	node.Handle("topology", handleTopology)
	node.Handle("from_broadcast", handleFromBroadcast)

	if err := node.Run(); err != nil {
		log.Fatal(err)
	}
}

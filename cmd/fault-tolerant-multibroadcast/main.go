package main

import (
	"encoding/json"
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

	go (func() {
		for _, c := range neighbors {
			c <- message
		}
	})()

	return node.Reply(msg, body)
}

func handleBroadcastTo(nodeID string) {
	c := neighbors[nodeID]

	payload := make(map[string]any, 2)
	payload["type"] = "from_broadcast"

	type response struct {
		Type string `json:"type"`
	}

	retry := func(msg float64) {
		time.Sleep(100 * time.Millisecond)
		c <- msg
	}

	for {
		message := <-c
		payload["message"] = message

		err := node.RPC(nodeID, payload, func(msg maelstrom.Message) error {
			var response response
			_ = json.Unmarshal(msg.Body, &response)

			if response.Type != "from_broadcast_ok" {
				log.Printf("node %s responded bad message: %s", nodeID, response.Type)
				retry(message)
			}

			return nil
		})

		if err != nil {
			retry(message)
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
		neighbors[id] = make(chan float64, 100)
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

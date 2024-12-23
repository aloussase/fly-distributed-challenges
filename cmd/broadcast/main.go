package main

import (
	"encoding/json"
	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
	"log"
)

var (
	node     *maelstrom.Node
	messages []float64
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

	return node.Reply(msg, body)
}

func handleRead(msg maelstrom.Message) error {
	body := make(map[string]any, 2)
	body["type"] = "read_ok"
	body["messages"] = messages
	return node.Reply(msg, body)
}

func handleTopology(msg maelstrom.Message) error {
	body := make(map[string]any, 1)
	body["type"] = "topology_ok"
	return node.Reply(msg, body)
}

func main() {
	messages = make([]float64, 0, 10)
	node = maelstrom.NewNode()

	node.Handle("broadcast", handleBroadcast)
	node.Handle("read", handleRead)
	node.Handle("topology", handleTopology)

	if err := node.Run(); err != nil {
		log.Fatal(err)
	}
}

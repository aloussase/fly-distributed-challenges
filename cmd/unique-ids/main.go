package main

import (
	"encoding/json"
	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
	"log"
	"math/rand"
	"sync"
	"time"
)

type IdGenerator interface {
	GenerateId() int64
}

type SnowflakeIdGenerator struct {
	mutex   sync.Mutex
	nodeId  int64
	current int64
}

func (s *SnowflakeIdGenerator) GenerateId() int64 {
	s.mutex.Lock()
	s.current++
	s.mutex.Unlock()

	timestamp := time.Now().UnixMilli()
	var id int64

	id |= timestamp & 0x0fffffffff000000
	id |= s.nodeId & 0x0000000000ff0000
	id |= s.current & 0x000000000000ffff

	return id
}

func NewSnowflakeIdGenerator() IdGenerator {
	generator := &SnowflakeIdGenerator{
		current: 0,
		nodeId:  rand.Int63(),
	}
	return generator
}

func handler(n *maelstrom.Node, generator IdGenerator) func(message maelstrom.Message) error {
	return func(message maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(message.Body, &body); err != nil {
			return err
		}

		body["type"] = "generate_ok"
		body["id"] = generator.GenerateId()

		return n.Reply(message, body)
	}
}

func main() {
	n := maelstrom.NewNode()

	generator := NewSnowflakeIdGenerator()
	n.Handle("generate", handler(n, generator))

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}

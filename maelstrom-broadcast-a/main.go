package main

import (
	"encoding/json"
	"log"
	"log/slog"
	"sync"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func main() {
	n := maelstrom.NewNode()
	var mu sync.Mutex
	var broadcast_msg []any
	var neighbors []string
	seen := make(map[int]struct{})

	// body = {"type": str, message: int, msg_id: any ?}
	type RPCBody struct {
		MsgID    any                 `json:"msg_id,omitempty"`
		Message  int                 `json:"message"`
		Type     string              `json:"type"`
		Topology map[string][]string `json:"topology,omitempty"`
	}

	n.Handle("broadcast", func(msg maelstrom.Message) error {
		// Unmarshal the message body as an loosely-typed map.
		var body RPCBody
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}
		// bytes, _ := json.Marshal(body)
		// var myDict map[string]any
		// _ = json.Unmarshal(bytes, &myDict)

		// Update the message type to return back.
		response_map := make(map[string]string)
		response_map["type"] = "broadcast_ok"

		// Add to the list
		mu.Lock()
		_, exists := seen[body.Message]
		if !exists {
			seen[body.Message] = struct{}{}
			broadcast_msg = append(broadcast_msg, body.Message)
			for _, neighbor := range neighbors {
				n.Send(neighbor, map[string]any{
					"type": "broadcast", "message": body.Message,
				})
			}
		}
		mu.Unlock()
		// Echo the original message back with the updated message type.
		if body.MsgID != nil {
			return n.Reply(msg, response_map)
		}
		return nil
	})

	n.Handle("read", func(msg maelstrom.Message) error {
		// Unmarshal the message body as an loosely-typed map.
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		mu.Lock()
		// Update the message type to return back.
		body["type"] = "read_ok"

		// Return the list
		body["messages"] = broadcast_msg
		mu.Unlock()
		// Echo the original message back with the updated message type.
		return n.Reply(msg, body)
	})

	n.Handle("topology", func(msg maelstrom.Message) error {
		// Unmarshal the message body as an loosely-typed map.
		slog.Info("nachikets test")
		var body RPCBody
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}
		// bytes, _ := json.Marshal(body)
		// var myDict map[string]any
		// _ = json.Unmarshal(bytes, &myDict)
		mu.Lock()
		neighbors = body.Topology[n.ID()]
		// neighbors, ok = body.Topology[msg.Src].([]string)
		// if !ok {
		// 	log.Fatal("something went wrong with the topology?")
		// }
		// } else {
		// 	log.Fatal("Topology bad")
		// }

		response_body := make(map[string]string)
		// Update the message type to return back.
		response_body["type"] = "topology_ok"
		mu.Unlock()
		// Echo the original message back with the updated message type.
		return n.Reply(msg, response_body)
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}

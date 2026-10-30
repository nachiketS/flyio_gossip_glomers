package main

import (
	"encoding/json"
	"log"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func main() {
	n := maelstrom.NewNode()

	var broadcast_msg []any
	var neighbors []string

	n.Handle("broadcast", func(msg maelstrom.Message) error {
		// Unmarshal the message body as an loosely-typed map.
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		// Update the message type to return back.
		body["type"] = "broadcast_ok"

		// Add to the list

		broadcast_msg = append(broadcast_msg, body["message"])
		for _, neighbor := range neighbors {
			n.Send(neighbor, map[any]any{
				"type": "broadcast", "message": body["message"],
			})
		}
		delete(body, "message")
		// Echo the original message back with the updated message type.
		return n.Reply(msg, body)
	})

	n.Handle("read", func(msg maelstrom.Message) error {
		// Unmarshal the message body as an loosely-typed map.
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		// Update the message type to return back.
		body["type"] = "read_ok"

		// Return the list
		body["messages"] = broadcast_msg

		// Echo the original message back with the updated message type.
		return n.Reply(msg, body)
	})

	n.Handle("topology", func(msg maelstrom.Message) error {
		// Unmarshal the message body as an loosely-typed map.
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		//delete topology
		if topology, ok := body["topology"].(map[string]any); ok {
			// 2. Now you can safely index it!
			neighbors, ok = topology[msg.Src].([]string)
			if !ok {
				log.Fatal("error")
			}
		} else {
			log.Fatal("Topology bad")
		}

		delete(body, "topology")
		// Update the message type to return back.
		body["type"] = "topology_ok"

		// Echo the original message back with the updated message type.
		return n.Reply(msg, body)
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}

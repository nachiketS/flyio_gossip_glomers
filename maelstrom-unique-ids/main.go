package main

import (
	"encoding/json"
	"hash/fnv"
	"log"
	"math/rand/v2"
	"strconv"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func main() {
	n := maelstrom.NewNode()
	n.Handle("generate", func(msg maelstrom.Message) error {
		// Unmarshal the message body as an loosely-typed map.
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		// Update the message type to return back.
		body["type"] = "generate_ok"

		// generate the id
		now := time.Now()
		now_unix := strconv.FormatInt(now.Unix(), 10)
		random_num := strconv.Itoa(rand.IntN(1000000000))
		src := msg.Src
		// random_num_2 := strconv.Itoa(rand.IntN(10000000))
		id_str := random_num + src + now_unix
		hasher := fnv.New64a()
		hasher.Write([]byte(id_str))

		numericHash := hasher.Sum64()

		body["id"] = numericHash
		// Echo the original message back with the updated message type.
		return n.Reply(msg, body)
	})
	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}

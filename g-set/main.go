package main

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type addReq struct {
	Element int `json:"element"`
}

type syncReq struct {
	Elements []int `json:"elements"`
}

type gSet struct {
	n  *maelstrom.Node
	s  map[int]struct{}
	mu sync.RWMutex
}

func createGSet(n *maelstrom.Node) *gSet {
	return &gSet{
		n: n,
		s: map[int]struct{}{},
	}
}

func (gs *gSet) startSync(cd time.Duration) {
	if gs.n.NodeIDs() == nil {
		return
	}
	go func() {
		for {
			els := gs.read()
			for _, dst := range gs.n.NodeIDs() {
				gs.n.Send(dst, map[string]any{
					"type":     "sync",
					"elements": els,
				})
			}
			time.Sleep(cd)
		}
	}()
}

func (gs *gSet) add(element int) {
	gs.mu.Lock()
	gs.s[element] = struct{}{}
	gs.mu.Unlock()
}

func (gs *gSet) read() []int {
	gs.mu.RLock()
	set := make([]int, len(gs.s))
	i := 0
	for el := range gs.s {
		set[i] = el
		i++
	}
	gs.mu.RUnlock()
	return set
}

func (gs *gSet) sync(elements []int) {
	gs.mu.Lock()
	for _, el := range elements {
		gs.s[el] = struct{}{}
	}
	gs.mu.Unlock()
}

func main() {
	n := maelstrom.NewNode()
	gs := createGSet(n)

	n.Handle("init", func(msg maelstrom.Message) error {
		gs.startSync(time.Second * 5)
		return nil
	})

	n.Handle("add", func(msg maelstrom.Message) error {
		var req addReq
		err := json.Unmarshal(msg.Body, &req)
		if err != nil {
			return err
		}

		gs.add(req.Element)

		return n.Reply(msg, map[string]string{
			"type": "add_ok",
		})
	})

	n.Handle("read", func(msg maelstrom.Message) error {
		return n.Reply(msg, map[string]any{
			"type":  "read_ok",
			"value": gs.read(),
		})
	})

	n.Handle("sync", func(msg maelstrom.Message) error {
		var req syncReq
		err := json.Unmarshal(msg.Body, &req)
		if err != nil {
			return err
		}

		gs.sync(req.Elements)
		return nil
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}

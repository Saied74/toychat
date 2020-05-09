package main

import (
	"log"
	"strings"
	"time"

	nats "github.com/nats-io/nats.go"
)

func playChatHandler(msg *nats.Msg, conn *nats.Conn) {
	// runner := broker.NewRunner(msg.Reply)
	sliceValue := strings.Split(string(msg.Data), " ")
	for i, j := 0, len(sliceValue)-1; i < j; i, j = i+1, j-1 {
		sliceValue[i], sliceValue[j] = sliceValue[j], sliceValue[i]
	}
	newvalue := strings.Join(sliceValue, " ")
	conn.Publish(msg.Reply, []byte(newvalue))
}

func main() {
	var err error

	nc1, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatal("Error from onnection", err)
	}
	defer nc1.Close()

	sub, _ := nc1.SubscribeSync("forChat")
	for {
		msg, err := sub.NextMsg(10 * time.Hour)
		if err != nil {
			log.Fatal("Error from Sub Sync: ", err)
		}
		go playChatHandler(msg, nc1)
	}
}

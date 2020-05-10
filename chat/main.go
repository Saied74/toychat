package main

import (
	"log"
	"strings"
	"time"

	nats "github.com/nats-io/nats.go"
)

//I have attempted to make this safe for concurrency.  In the main goroutine
//once the message is recieved, it is handed to the playChatHandler goroutine
//with the connection so playChatHandler goroutine can complete its own
//work without needing any other data and return the results through the
//connection.  This needs to be tested under load.

func playChatHandler(msg *nats.Msg, conn *nats.Conn) {
	sliceValue := strings.Split(string(msg.Data), " ")
	for i, j := 0, len(sliceValue)-1; i < j; i, j = i+1, j-1 {
		sliceValue[i], sliceValue[j] = sliceValue[j], sliceValue[i]
	}
	newvalue := strings.Join(sliceValue, " ")
	conn.Publish(msg.Reply, []byte(newvalue))
}

func main() {
	var err error
	// TODO: perhaps create a connection pool for better efficiency.
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

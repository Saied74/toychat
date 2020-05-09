package main

import (
	"log"
	"strings"
	"time"

	nats "github.com/nats-io/nats.go"
)

func playMatHandler(msg *nats.Msg, conn *nats.Conn) {
	var matValue = ""
	sliceLen := len(strings.Split(string(msg.Data), " "))
	for i := 0; i < sliceLen; i++ {
		matValue += "mat "
	}
	conn.Publish(msg.Reply, []byte(matValue))
}

func main() {
	var err error

	nc1, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatal("Error from onnection", err)
	}
	defer nc1.Close()

	sub, _ := nc1.SubscribeSync("forMat")
	for {
		msg, err := sub.NextMsg(10 * time.Hour)
		if err != nil {
			log.Fatal("Error from Sub Sync: ", err)
		}
		go playMatHandler(msg, nc1)
	}
}

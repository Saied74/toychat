package main

import (
	"log"
	"strings"
	"time"

	nats "github.com/nats-io/nats.go"
)

func playChatHandler(chat []byte) []byte {
	sliceValue := strings.Split(string(chat), " ")
	for i, j := 0, len(sliceValue)-1; i < j; i, j = i+1, j-1 {
		sliceValue[i], sliceValue[j] = sliceValue[j], sliceValue[i]
	}
	newvalue := strings.Join(sliceValue, " ")
	return []byte(newvalue)
}

func main() {
	var err error

	nc1, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatal("Error from onnection", err)
	}
	defer nc1.Close()

	for {
		sub, _ := nc1.SubscribeSync("forChat") //, func(matMsg *nats.Msg) { //playMatHandler)
		msg, err := sub.NextMsg(10 * time.Hour)
		if err != nil {
			log.Fatal("Error from Sub Sync: ", err)
		}
		matValue := playChatHandler(msg.Data)
		nc1.Publish(msg.Reply, matValue)
		time.Sleep(1 * time.Millisecond)
	}
}

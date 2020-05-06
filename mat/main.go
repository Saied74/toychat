package main

import (
	"log"
	"strings"
	"time"

	nats "github.com/nats-io/nats.go"
)

func playMatHandler(matMsg []byte) []byte {
	var matValue = ""
	sliceLen := len(strings.Split(string(matMsg), " "))
	for i := 0; i < sliceLen; i++ {
		matValue += "mat "
	}
	return []byte(matValue)
}

func main() {
	var err error

	nc1, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatal("Error from onnection", err)
	}
	defer nc1.Close()

	for {
		sub, _ := nc1.SubscribeSync("forMat") //, func(matMsg *nats.Msg) { //playMatHandler)
		msg, err := sub.NextMsg(10 * time.Hour)
		if err != nil {
			log.Fatal("Error from Sub Sync: ", err)
		}
		matValue := playMatHandler(msg.Data)
		nc1.Publish(msg.Reply, matValue)
		time.Sleep(1 * time.Millisecond)
	}
}

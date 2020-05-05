package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	nats "github.com/nats-io/nats.go"
)

func playChatHandler(chat string) []byte {
	sliceValue := strings.Split(chat, " ")
	for i, j := 0, len(sliceValue)-1; i < j; i, j = i+1, j-1 {
		sliceValue[i], sliceValue[j] = sliceValue[j], sliceValue[i]
	}
	newvalue := strings.Join(sliceValue, " ")
	return []byte(newvalue)
}

func main() {
	var err error
	var m = &nats.Msg{}

	nc1, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatal("Error from onnection", err)
	}
	defer nc1.Close()

	nc2, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatal("Error from onnection", err)
	}
	defer nc2.Close()

	for {
		sub, err := nc1.SubscribeSync("forChat")
		if err != nil {
			log.Fatal("Error from Sub Sync: ", err)
		}
		m, err = sub.NextMsg(20 * time.Hour)
		if err != nil {
			log.Fatal("Error from next message, timed out: ", err)
		}
		matValue := playChatHandler(string(m.Data))
		fmt.Printf("Message from the far side: %s\n", string(m.Data))
		nc2.Publish("fromChat", matValue)

	}

}

//
// nc, _ := nats.Connect(nats.DefaultURL)
// defer nc.Close()
// scanner := bufio.NewScanner(os.Stdin)
//
// for scanner.Scan() {
//
//   fmt.Println(scanner.Bytes())
//
//   nc.Publish("foo", scanner.Bytes())
//
//   fmt.Printf("and the input is %s\n", scanner.Text())
// }

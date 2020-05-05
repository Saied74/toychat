package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	nats "github.com/nats-io/nats.go"
)

func playMatHandler(mat string) []byte {
	var matValue = ""
	sliceLen := len(strings.Split(mat, " "))
	for i := 0; i < sliceLen; i++ {
		matValue += "mat "
	}
	return []byte(matValue)
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
		sub, err := nc1.SubscribeSync("forMat")
		if err != nil {
			log.Fatal("Error from Sub Sync: ", err)
		}
		m, err = sub.NextMsg(20 * time.Hour)
		if err != nil {
			log.Fatal("Error from next message, timed out: ", err)
		}
		matValue := playMatHandler(string(m.Data))
		fmt.Printf("Message from the far side: %s\n", string(m.Data))
		nc2.Publish("fromMat", matValue)

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

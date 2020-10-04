package main

import (
	"fmt"
	"net/url"

	"github.com/gorilla/websocket"
)

func main() {

	u := url.URL{
		Scheme: "ws",
		Host:   "localhost:26657",
		Path:   "/websocket",
	}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		// handle error
		fmt.Println(err)
	}

	err = c.WriteMessage(websocket.TextMessage, []byte("tm.event = 'Tx' AND tx.height = 3"))
	if err != nil {
		fmt.Println(err)
	}

	_, message, err := c.ReadMessage()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(message)
	// go func() {
	// 	for e := range txs {
	// 		fmt.Println(message)
	// 	}
	// }()

}

// Copyright 2018 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Binary textdemo displays a couple of Text widgets.
// Exist when 'q' is pressed.
package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"log"
	"os"
	"os/signal"

	"github.com/sacOO7/gowebsocket"
	"github.com/tidwall/gjson"

	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/termbox"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/text"
)

// quotations are used as text that is rolled up in a text widget.
var quotations = []string{
	"When some see coincidence, I see consequence. When others see chance, I see cost.",
	"You cannot pass....I am a servant of the Secret Fire, wielder of the flame of Anor. You cannot pass. The dark fire will not avail you, flame of Udûn. Go back to the Shadow! You cannot pass.",
	"I'm going to make him an offer he can't refuse.",
	"May the Force be with you.",
	"The stuff that dreams are made of.",
	"There's no place like home.",
	"Show me the money!",
	"I want to be alone.",
	"I'll be back.",
}

// writeLines writes a line of text to the text widget every delay.
// Exits when the context expires.
func writeLines(ctx context.Context, t *text.Text, delay time.Duration) {
	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s)
	ticker := time.NewTicker(delay)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			i := r.Intn(len(quotations))
			if err := t.Write(fmt.Sprintf("%s\n", quotations[i])); err != nil {
				panic(err)
			}

		case <-ctx.Done():
			return
		}
	}
}

// writeTransactions writes the latest Transactions from the websocket.
// Exits when the context expires.
func writeTransactions(ctx context.Context, t *text.Text) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	socket := gowebsocket.New("ws://localhost:26657/websocket")

	socket.OnConnectError = func(err error, socket gowebsocket.Socket) {
		log.Fatal("Received connect error - ", err)
	}
	socket.OnTextMessage = func(message string, socket gowebsocket.Socket) {
		currentTx := gjson.Get(message, "result.data.value.TxResult")
		currentTime := time.Now()
		height := currentTx.Get("height").String()
		txType := currentTx.Get("result.events.1.type").String()
		if currentTx.String() != "" {
			if err := t.Write(fmt.Sprintf("%s\n", currentTime.Format("2006-01-02 03:04:05 PM")+"\nTransaction height "+height+" type "+txType)); err != nil {
				panic(err)
			}
		}
	}

	socket.OnDisconnected = func(err error, socket gowebsocket.Socket) {
		log.Println("Disconnected from server ")
		return
	}

	socket.Connect()

	socket.SendText("{ \"jsonrpc\": \"2.0\", \"method\": \"subscribe\", \"params\": [\"tm.event='Tx'\"], \"id\": 2 }")

	for {
		select {
		case <-ctx.Done():
			log.Println("interrupt")
			socket.Close()
			return
		}
	}
}

// writeBlocks writes the latest Block from the websocket.
// Exits when the context expires.
func writeBlocks(ctx context.Context, t *text.Text) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	socket := gowebsocket.New("ws://localhost:26657/websocket")

	socket.OnConnectError = func(err error, socket gowebsocket.Socket) {
		log.Fatal("Received connect error - ", err)
	}

	socket.OnTextMessage = func(message string, socket gowebsocket.Socket) {
		currentBlock := gjson.Get(message, "result.data.value.block.header.height")
		if currentBlock.String() != "" {
			if err := t.Write(fmt.Sprintf("%s\n", "Latest block height "+currentBlock.String())); err != nil {
				panic(err)
			}
		}

	}

	socket.OnDisconnected = func(err error, socket gowebsocket.Socket) {
		log.Println("Disconnected from server ")
		return
	}

	socket.Connect()

	socket.SendText("{ \"jsonrpc\": \"2.0\", \"method\": \"subscribe\", \"params\": [\"tm.event='NewBlock'\"], \"id\": 1 }")

	for {
		select {
		case <-ctx.Done():
			log.Println("interrupt")
			socket.Close()
			return
		}
	}
}

func main() {
	t, err := termbox.New()
	if err != nil {
		panic(err)
	}
	defer t.Close()

	ctx, cancel := context.WithCancel(context.Background())
	borderless, err := text.New()
	if err != nil {
		panic(err)
	}
	if err := borderless.Write("Text without border."); err != nil {
		panic(err)
	}

	unicode, err := text.New()
	if err != nil {
		panic(err)
	}
	if err := unicode.Write("你好，世界!"); err != nil {
		panic(err)
	}

	trimmed, err := text.New()
	if err != nil {
		panic(err)
	}
	if err := trimmed.Write("Trims lines that don't fit onto the canvas because they are too long for its width.."); err != nil {
		panic(err)
	}

	wrapped, err := text.New(text.WrapAtRunes())
	if err != nil {
		panic(err)
	}
	if err := wrapped.Write("Supports", text.WriteCellOpts(cell.FgColor(cell.ColorRed))); err != nil {
		panic(err)
	}
	if err := wrapped.Write(" colors", text.WriteCellOpts(cell.FgColor(cell.ColorBlue))); err != nil {
		panic(err)
	}
	if err := wrapped.Write(". Wraps long lines at rune boundaries if the WrapAtRunes() option is provided.\nSupports newline character to\ncreate\nnewlines\nmanually.\nTrims the content if it is too long.\n\n\n\nToo long."); err != nil {
		panic(err)
	}

	rolled, err := text.New(text.RollContent(), text.WrapAtWords())
	if err != nil {
		panic(err)
	}
	if err := rolled.Write("Rolls the content upwards if RollContent() option is provided.\nSupports keyboard and mouse scrolling.\n\n"); err != nil {
		panic(err)
	}

	rolledBlocks, err := text.New(text.RollContent(), text.WrapAtWords())
	if err != nil {
		panic(err)
	}
	if err := rolledBlocks.Write("Logs latest blocks.\nSupports keyboard and mouse scrolling.\n\n"); err != nil {
		panic(err)
	}

	rolledTx, err := text.New(text.RollContent(), text.WrapAtWords())
	if err != nil {
		panic(err)
	}
	if err := rolledTx.Write("Logs latest transactions.\nSupports keyboard and mouse scrolling.\n\n"); err != nil {
		panic(err)
	}

	go writeLines(ctx, rolled, 1*time.Second)
	go writeBlocks(ctx, rolledBlocks)
	go writeTransactions(ctx, rolledTx)

	c, err := container.New(
		t,
		container.Border(linestyle.Light),
		container.BorderTitle("PRESS Q TO QUIT"),
		container.SplitVertical(
			container.Left(
				container.Border(linestyle.Light),
				container.BorderTitle("Latest Blocks"),
				container.PlaceWidget(rolledBlocks),
			),
			container.Right(
				container.Border(linestyle.Light),
				container.BorderTitle("Latest Transactions"),
				container.PlaceWidget(rolledTx),
			),
		),
	)
	if err != nil {
		panic(err)
	}

	quitter := func(k *terminalapi.Keyboard) {
		if k.Key == 'q' || k.Key == 'Q' {
			cancel()
		}
	}

	if err := termdash.Run(ctx, t, c, termdash.KeyboardSubscriber(quitter)); err != nil {
		panic(err)
	}
}

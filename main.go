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
	"time"

	"log"
	"os"
	"os/signal"

	"gopkg.in/resty.v1"

	"github.com/jedib0t/go-pretty/table"
	"github.com/sacOO7/gowebsocket"
	"github.com/tidwall/gjson"

	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/termbox"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/gauge"
	"github.com/mum4k/termdash/widgets/text"
)

const (
	appRPC = "http://localhost:26657/"
)

func getFromRPC(endpoint string) string {
	resp, _ := resty.R().
		SetHeader("Cache-Control", "no-cache").
		SetHeader("Content-Type", "application/json").
		Get(appRPC + endpoint)

	return resp.String()
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

// writeLines writes a line of text to the text widget every delay.
// Exits when the context expires.
func writeValidators(ctx context.Context, t *text.Text, delay time.Duration) {
	ticker := time.NewTicker(delay)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			validators := gjson.Get(getFromRPC("validators"), "result.validators")
			t.Reset()

			validators.ForEach(func(key, validator gjson.Result) bool {

				ta := table.NewWriter()
				ta.AppendRow([]interface{}{key.Int(), validator.Get("address").String(), validator.Get("voting_power").String()})

				if err := t.Write(fmt.Sprintf("%s\n", ta.Render())); err != nil {
					panic(err)
				}
				return true // keep iterating
			})

		case <-ctx.Done():
			return
		}
	}
}

// playGauge continuously changes the displayed percent value on the gauge by the
// step once every delay. Exits when the context expires.
func playGauge(ctx context.Context, g *gauge.Gauge, blockHeight int64) {
	var progress int64 = 0

	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:

			maxHeight := gjson.Get(getFromRPC("dump_consensus_state"), "result.round_state.height").Int()

			progress = (blockHeight / maxHeight) * 100

			if err := g.Absolute(int(progress), 100); err != nil {
				panic(err)
			}

		case <-ctx.Done():
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

	networkStatus := gjson.Parse(getFromRPC("status"))

	ctx, cancel := context.WithCancel(context.Background())

	trimmed, err := text.New()
	if err != nil {
		panic(err)
	}
	if err := trimmed.Write("Trims lines that don't fit onto the canvas because they are too long for its width.."); err != nil {
		panic(err)
	}

	// Block parsing widget
	rolledBlocks, err := text.New(text.RollContent(), text.WrapAtWords())
	if err != nil {
		panic(err)
	}
	if err := rolledBlocks.Write("Latest block height " + networkStatus.Get("result.sync_info.latest_block_height").String() + "\n"); err != nil {
		panic(err)
	}

	// Transaction parsing widget
	rolledTx, err := text.New(text.RollContent(), text.WrapAtWords())
	if err != nil {
		panic(err)
	}
	if err := rolledTx.Write("Logs latest transactions.\nSupports keyboard and mouse scrolling.\n\n"); err != nil {
		panic(err)
	}

	rolledValidators, err := text.New(text.RollContent(), text.WrapAtWords())
	if err != nil {
		panic(err)
	}
	if err := rolledValidators.Write("Rolls the content upwards if RollContent() option is provided.\nSupports keyboard and mouse scrolling.\n\n"); err != nil {
		panic(err)
	}

	go writeValidators(ctx, rolledValidators, 10*time.Second)
	go writeBlocks(ctx, rolledBlocks)
	go writeTransactions(ctx, rolledTx)

	// blockchain download gauge
	absolute, err := gauge.New(
		gauge.Height(1),
		gauge.Color(cell.ColorBlue),
		gauge.Border(linestyle.Light),
		gauge.BorderTitle("Blockchain download %"),
	)
	if err != nil {
		panic(err)
	}

	if networkStatus.Get("result.sync_info.catching_up").String() == "false" {
		if err := absolute.Absolute(100, 100); err != nil {
			panic(err)
		}
	} else {
		go playGauge(ctx, absolute, networkStatus.Get("result.sync_info.latest_block_height").Int())
	}

	// Draw Dashboard
	c, err := container.New(
		t,
		container.Border(linestyle.Light),
		container.BorderTitle("PRESS Q TO QUIT | Network "+networkStatus.Get("result.node_info.network").String()+" Version "+networkStatus.Get("result.node_info.version").String()),
		container.BorderColor(cell.ColorNumber(2)),
		container.SplitHorizontal(
			container.Top(
				container.SplitVertical(
					container.Left(
						container.PlaceWidget(absolute),
					),
					container.Right(
						container.Border(linestyle.Light),
						container.BorderTitle("Validators"),
						container.PlaceWidget(rolledValidators),
					),
				),
			),
			container.Bottom(
				container.SplitVertical(
					container.Left(
						container.Border(linestyle.Light),
						container.BorderTitle("Latest Blocks"),
						container.PlaceWidget(rolledBlocks),
					), container.Right(
						container.Border(linestyle.Light),
						container.BorderTitle("Latest Confirmed Transactions"),
						container.PlaceWidget(rolledTx),
					),
				),
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

// Copyright 2018 Goole Inc.
// Copyright 2020 Tobias Schwarz
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

// Binary explorer demo. Displays widgets for insights of blockchain behaviour.
// Exist when 'q' or 'esc' is pressed.
package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"log"

	"gopkg.in/resty.v1"

	"github.com/google/uuid"
	"github.com/jedib0t/go-pretty/table"
	ga "github.com/ozgur-soft/google-analytics/src"
	"github.com/sacOO7/gowebsocket"
	"github.com/tidwall/gjson"

	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/termbox"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/text"
)

const (
	appRPC        = "http://localhost"
)

var givenPort = flag.String("p", "26657", "port to connect to as a string")

type Info struct {
    blocks *Blocks
}
type Blocks struct {
	amount int
	seconds_passed int
} 
func incrInfoBlocks(i Info) {
    i.blocks.amount++
}
func incrInfoSeconds(i Info) {
    i.blocks.seconds_passed++
}

func main() {
	view()

	// Init internal variables
	info := Info{}
    info.blocks = new(Blocks)

	connectionSignal := make(chan string)
	t, err := termbox.New()
	if err != nil {
		panic(err)
	}
	defer t.Close()

	flag.Parse()

	networkInfo := getFromRPC("status")
	networkStatus := gjson.Parse(networkInfo)
	if !networkStatus.Exists() {
		panic("Application not running on localhost:" + fmt.Sprintf("%s", *givenPort))
	}

	ctx, cancel := context.WithCancel(context.Background())


	// Creates the initial text for the health widget
	healthWidget, err := text.New()
	if err != nil {
		panic(err)
	}
	if err := healthWidget.Write("⌛ loading"); err != nil {
		panic(err)
	}

	// Creates the initial text for the system time widget
	timeWidget, err := text.New()
	if err != nil {
		panic(err)
	}
	currentTime := time.Now()
	if err := timeWidget.Write(fmt.Sprintf("%s\n", currentTime.Format("2006-01-02\n03:04:05 PM"))); err != nil {
		panic(err)
	}

	// Creates the initial text for the block size widget
	maxBlocksizeWidget, err := text.New()
	maxBlockSize := gjson.Get(getFromRPC("consensus_params"), "result.consensus_params.block.max_bytes").Int()
	if err != nil {
		panic(err)
	}
	if err := maxBlocksizeWidget.Write(fmt.Sprintf("%s", byteCountDecimal(maxBlockSize))); err != nil {
		panic(err)
	}

	// Creates the initial text for the peer widget
	peerWidget, err := text.New()
	if err != nil {
		panic(err)
	}
	if err := peerWidget.Write("0"); err != nil {
		panic(err)
	}

	// validator widget
	validatorWidget, err := text.New(text.RollContent(), text.WrapAtWords())
	if err != nil {
		panic(err)
	}
	if err := validatorWidget.Write("List available validators.\n\n"); err != nil {
		panic(err)
	}

	// Transaction parsing widget
	transactionWidget, err := text.New(text.RollContent(), text.WrapAtWords())
	if err != nil {
		panic(err)
	}
	if err := transactionWidget.Write("Transactions will appear as soon as they are confirmed in a block.\n\n"); err != nil {
		panic(err)
	}
	
	// Blocks parsing widget
	blocksWidget, err := text.New(text.RollContent(), text.WrapAtWords())
	if err != nil {
		panic(err)
	}
	if err := blocksWidget.Write("Latest block height " + networkStatus.Get("result.sync_info.latest_block_height").String() + "\n"); err != nil {
		panic(err)
	}

	// create seconds per block widget
	secondsPerBlockWidget, err := text.New(text.RollContent(), text.WrapAtWords())
	if err != nil {
		panic(err)
	}
	if err := secondsPerBlockWidget.Write("0"); err != nil {
		panic(err)
	}

	// create current network widget
	currentNetworkWidget, err := text.New(text.RollContent(), text.WrapAtWords())
	if err != nil {
		panic(err)
	}
	if err := currentNetworkWidget.Write(networkStatus.Get("result.node_info.network").String()); err != nil {
		panic(err)
	}

	// The functions that execute the updating widgets.

	// system powered widgets
	go writeTime(ctx, info, timeWidget, 1*time.Second)

	// rpc widgets
	go writePeers(ctx, peerWidget, 1*time.Second)
	go writeHealth(ctx, healthWidget, 500*time.Millisecond, connectionSignal)
	go writeSecondsPerBlock(ctx, info, secondsPerBlockWidget, 1*time.Second)

	// websocket powered widgets
	go writeValidators(ctx, validatorWidget, connectionSignal)
	go writeBlocks(ctx, info, blocksWidget, connectionSignal)
	go writeTransactions(ctx, transactionWidget, connectionSignal)

	// Draw Dashboard
	c, err := container.New(
		t,
		container.Border(linestyle.Light),
		container.BorderTitle("GEX: PRESS Q or ESC TO QUIT"),
		container.BorderColor(cell.ColorNumber(2)),
		container.SplitHorizontal(
			container.Top(
				container.SplitVertical(
					container.Left(
						container.SplitHorizontal(
							container.Top(
								container.SplitVertical(
									container.Left(
										container.SplitVertical(
											container.Left(
												container.Border(linestyle.Light),
												container.BorderTitle("Network"),
												container.PlaceWidget(currentNetworkWidget),
											),
											container.Right(
												container.Border(linestyle.Light),
												container.BorderTitle("Health"),
												container.PlaceWidget(healthWidget),
											),
										),
									),
									container.Right(
										container.SplitVertical(
											container.Left(
												container.Border(linestyle.Light),
												container.BorderTitle("System Time"),
												container.PlaceWidget(timeWidget),
											),
											container.Right(
												container.Border(linestyle.Light),
												container.BorderTitle("Connected Peers"),
												container.PlaceWidget(peerWidget),
											),
										),
									),
								),
							),
							container.Bottom(
								// INSERT NEW BOTTOM ROWS
								container.SplitVertical(
									container.Left(
										container.SplitVertical(
											container.Left(
												container.Border(linestyle.Light),
												container.BorderTitle("Time between blocks"),
												container.PlaceWidget(secondsPerBlockWidget),
											),
											container.Right(
												container.Border(linestyle.Light),
												container.BorderTitle("Max Block Size"),
												container.PlaceWidget(maxBlocksizeWidget),
											),
										),
									),
									container.Right(
										container.SplitVertical(
											container.Left(
												// container.Border(linestyle.Light),
												// container.BorderTitle("Max Block Size"),
												// container.PlaceWidget(maxBlocksizeWidget),
											),
											container.Right(
												// container.Border(linestyle.Light),
												// container.BorderTitle("Connected IBC Channels"),
												// container.PlaceWidget(peerWidget),
											),
										),
									),
								),
							),
						),
					),
					container.Right(
						container.Border(linestyle.Light),
						container.BorderTitle("Validators"),
						container.PlaceWidget(validatorWidget),
					),
				),
			),
			container.Bottom(
				container.SplitVertical(
					container.Left(
						container.Border(linestyle.Light),
						container.BorderTitle("Latest Blocks"),
						container.PlaceWidget(blocksWidget),
					), container.Right(
						container.Border(linestyle.Light),
						container.BorderTitle("Latest Confirmed Transactions"),
						container.PlaceWidget(transactionWidget),
					),
				),
			),
		),
	)
	if err != nil {
		panic(err)
	}

	quitter := func(k *terminalapi.Keyboard) {
		if k.Key == 'q' || k.Key == 'Q' || k.Key == keyboard.KeyEsc {
			cancel()
		}
	}

	if err := termdash.Run(ctx, t, c, termdash.KeyboardSubscriber(quitter)); err != nil {
		panic(err)
	}
}

func getFromRPC(endpoint string) string {
	port := *givenPort
	resp, _ := resty.R().
		SetHeader("Cache-Control", "no-cache").
		SetHeader("Content-Type", "application/json").
		Get(appRPC + ":" + port + "/" + endpoint)

	return resp.String()
}

// writeTime writes the current system time to the timeWidget.
// Exits when the context expires.
func writeTime(ctx context.Context, info Info, t *text.Text, delay time.Duration) {
	ticker := time.NewTicker(delay)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			currentTime := time.Now()
			t.Reset()
			if err := t.Write(fmt.Sprintf("%s\n", currentTime.Format("2006-01-02\n03:04:05 PM"))); err != nil {
				panic(err)
			}
			incrInfoSeconds(info)
		case <-ctx.Done():
			return
		}
	}
}

// writeHealth writes the status to the healthWidget.
// Exits when the context expires.
func writeHealth(ctx context.Context, t *text.Text, delay time.Duration, connectionSignal chan string) {
	reconnect := false
	health := gjson.Get(getFromRPC("health"), "result")
	t.Reset()
	if health.Exists() {
		t.Write("✔️ good")
	} else {
		t.Write("✖️ not connected")
	}

	ticker := time.NewTicker(delay)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			health := gjson.Get(getFromRPC("health"), "result")
			if health.Exists() {
				t.Reset()
				t.Write("✔️ good")
				if reconnect == true {
					connectionSignal <- "reconnect"
					connectionSignal <- "reconnect"
					connectionSignal <- "reconnect"
					reconnect = false
				}
			} else {
				t.Reset()
				t.Write("✖️ not connected")
				if reconnect == false {
					connectionSignal <- "no_connection"
					connectionSignal <- "no_connection"
					connectionSignal <- "no_connection"
					reconnect = true
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

// writeSecondsPerBlock writes the status to the Time per block.
// Exits when the context expires.
func writeSecondsPerBlock(ctx context.Context, info Info, t *text.Text, delay time.Duration) {

	t.Reset()

	ticker := time.NewTicker(delay)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			t.Reset()
			blocksPerSecond := 0.00
			if(info.blocks.seconds_passed != 0) {
				blocksPerSecond = float64(info.blocks.seconds_passed) / float64(info.blocks.amount)
			}
			
			t.Write(fmt.Sprintf("%.2f seconds", blocksPerSecond))
		case <-ctx.Done():
			return
		}
	}
}

// writePeers writes the connected Peers to the peerWidget.
// Exits when the context expires.
func writePeers(ctx context.Context, t *text.Text, delay time.Duration) {
	peers := gjson.Get(getFromRPC("net_info"), "result.n_peers").String()
	t.Reset()
	if peers != "" {
		t.Write(peers)
	}
	if err := t.Write(peers); err != nil {
		panic(err)
	}

	ticker := time.NewTicker(delay)
	t.Reset()
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			t.Reset()
			peers := gjson.Get(getFromRPC("net_info"), "result.n_peers").String()
			if peers != "" {
				t.Reset()
				t.Write(peers)
			}

		case <-ctx.Done():
			return
		}
	}
}

// writeTransactions writes the latest Transactions to the transactionsWidget.
// Exits when the context expires.
func writeTransactions(ctx context.Context, t *text.Text, connectionSignal <-chan string) {
	port := *givenPort
	socket := gowebsocket.New("ws://localhost:" + port + "/websocket")

	socket.OnTextMessage = func(message string, socket gowebsocket.Socket) {
		currentTx := gjson.Get(message, "result.data.value.TxResult.result.log")
		currentTime := time.Now()
		if currentTx.String() != "" {
			if err := t.Write(fmt.Sprintf("%s\n", currentTime.Format("2006-01-02 03:04:05 PM")+"\n"+currentTx.String())); err != nil {
				panic(err)
			}
		}
	}

	socket.Connect()

	socket.SendText("{ \"jsonrpc\": \"2.0\", \"method\": \"subscribe\", \"params\": [\"tm.event='Tx'\"], \"id\": 2 }")

	for {
		select {
		case s := <-connectionSignal:
			if s == "no_connection" {
				socket.Close()
			}
			if s == "reconnect" {
				writeTransactions(ctx, t, connectionSignal)
			}
		case <-ctx.Done():
			log.Println("interrupt")
			socket.Close()
			return
		}
	}
}

// writeBlocks writes the latest Block to the blocksWidget.
// Exits when the context expires.
func writeBlocks(ctx context.Context, info Info, t *text.Text, connectionSignal <-chan string) {

	port := *givenPort
	socket := gowebsocket.New("ws://localhost:" + port + "/websocket")

	socket.OnTextMessage = func(message string, socket gowebsocket.Socket) {
		currentBlock := gjson.Get(message, "result.data.value.block.header.height")
		if currentBlock.String() != "" {
			if err := t.Write(fmt.Sprintf("%s\n", "Latest block height "+currentBlock.String())); err != nil {
				panic(err)
			}
			incrInfoBlocks(info)
		}

	}

	socket.Connect()

	socket.SendText("{ \"jsonrpc\": \"2.0\", \"method\": \"subscribe\", \"params\": [\"tm.event='NewBlock'\"], \"id\": 1 }")

	for {
		select {
		case s := <-connectionSignal:
			if s == "no_connection" {
				socket.Close()
			}
			if s == "reconnect" {
				writeBlocks(ctx, info, t, connectionSignal)
			}
		case <-ctx.Done():
			log.Println("interrupt")
			socket.Close()
			return
		}
	}
}

// writeValidators writes the current validator set to the validatoWidget
// Exits when the context expires.
func writeValidators(ctx context.Context, t *text.Text, connectionSignal <-chan string) {
	port := *givenPort
	socket := gowebsocket.New("ws://localhost:" + port + "/websocket")

	socket.OnConnected = func(socket gowebsocket.Socket) {
		validators := gjson.Get(getFromRPC("validators"), "result.validators")
		t.Reset()
		i := 1
		validators.ForEach(func(key, validator gjson.Result) bool {

			ta := table.NewWriter()
			ta.AppendRow([]interface{}{fmt.Sprintf("%d", i), validator.Get("address").String(), validator.Get("voting_power").String()})

			if err := t.Write(fmt.Sprintf("%s\n", ta.Render())); err != nil {
				panic(err)
			}
			i++
			return true // keep iterating
		})
	}

	socket.OnTextMessage = func(message string, socket gowebsocket.Socket) {
		validators := gjson.Get(getFromRPC("validators"), "result.validators")
		t.Reset()

		i := 1
		validators.ForEach(func(key, validator gjson.Result) bool {

			ta := table.NewWriter()
			ta.AppendRow([]interface{}{fmt.Sprintf("%d", i), validator.Get("address").String(), validator.Get("voting_power").String()})

			if err := t.Write(fmt.Sprintf("%s\n", ta.Render())); err != nil {
				panic(err)
			}
			i++
			return true // keep iterating
		})
	}

	socket.Connect()

	socket.SendText("{ \"jsonrpc\": \"2.0\", \"method\": \"subscribe\", \"params\": [\"tm.event='ValidatorSetUpdates'\"], \"id\": 3 }")

	for {
		select {
		case s := <-connectionSignal:
			if s == "no_connection" {
				socket.Close()
			}
			if s == "reconnect" {
				writeValidators(ctx, t, connectionSignal)
			}
		case <-ctx.Done():
			log.Println("interrupt")
			socket.Close()
			return
		}
	}
}

// byteCountDecimal calculates bytes integer to a human readable decimal number
func byteCountDecimal(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "kMGTPE"[exp])
}

func view() {
	api := new(ga.API)
	api.ContentType = "application/x-www-form-urlencoded"

	client := new(ga.Client)
	client.ProtocolVersion = "1"
	client.ClientID = uuid.New().String()
	client.TrackingID = "UA-183957259-1"
	client.HitType = "event"
	client.DocumentLocationURL = "https://github.com/cosmos/gex"
	client.DocumentTitle = "Dashboard"
	client.DocumentEncoding = "UTF-8"
	client.EventCategory = "Start"
	client.EventAction = "Dashboard"
	client.EventLabel = "start"

	api.Send(client)
}

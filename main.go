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
	"strconv"

	"log"

	"gopkg.in/resty.v1"

	"github.com/google/uuid"
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
	"github.com/mum4k/termdash/widgets/donut"
)

const (
	// RPC requests are made to the native app running
	appRPC        = "http://localhost"
	// donut widget constants
	playTypePercent playType = iota
	playTypeAbsolute
)

// optional port variable. example: `gex -p 30057`
var givenPort = flag.String("p", "26657", "port to connect to as a string")

// Info describes a list of types with data that are used in the explorer
type Info struct {
	blocks *Blocks
	transactions *Transactions
}

// Blocks describe content that gets parsed for blocks
type Blocks struct {
	amount int
	secondsPassed int
	totalGasWanted int64
	gasWantedLatestBlock int64
	maxGasWanted int64
	lastTx int64
} 

// Transactions describe content that gets parsed for transactions
type Transactions struct {
	amount uint64
}

// playType indicates the type of the donut widget.
type playType int

func main() {
	view()

	// Init internal variables
	info := Info{}
	info.blocks = new(Blocks)
	info.transactions = new(Transactions)

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

	genesisInfo := gjson.Parse(getFromRPC("genesis"))

	ctx, cancel := context.WithCancel(context.Background())


	// START INITIALISING WIDGETS

	// Creates Network Widget
	currentNetworkWidget, err := text.New(text.RollContent(), text.WrapAtWords())
	if err != nil {
		panic(err)
	}
	if err := currentNetworkWidget.Write(networkStatus.Get("result.node_info.network").String()); err != nil {
		panic(err)
	}

	// Creates Health Widget
	healthWidget, err := text.New()
	if err != nil {
		panic(err)
	}
	if err := healthWidget.Write("⌛ loading"); err != nil {
		panic(err)
	}

	// Creates System Time Widget
	timeWidget, err := text.New()
	if err != nil {
		panic(err)
	}
	currentTime := time.Now()
	if err := timeWidget.Write(fmt.Sprintf("%s\n", currentTime.Format("2006-01-02\n03:04:05 PM"))); err != nil {
		panic(err)
	}

	// Creates Connected Peers Widget
	peerWidget, err := text.New()
	if err != nil {
		panic(err)
	}
	if err := peerWidget.Write("0"); err != nil {
		panic(err)
	}

	// Creates Seconds Between Blocks Widget
	secondsPerBlockWidget, err := text.New(text.RollContent(), text.WrapAtWords())
	if err != nil {
		panic(err)
	}
	if err := secondsPerBlockWidget.Write("0"); err != nil {
		panic(err)
	}

	// Creates Max Block Size Widget
	maxBlocksizeWidget, err := text.New()
	maxBlockSize := gjson.Get(getFromRPC("consensus_params"), "result.consensus_params.block.max_bytes").Int()
	if err != nil {
		panic(err)
	}
	if err := maxBlocksizeWidget.Write(fmt.Sprintf("%s", byteCountDecimal(maxBlockSize))); err != nil {
		panic(err)
	}

	// Creates Validators widget
	validatorWidget, err := text.New(text.RollContent(), text.WrapAtWords())
	if err != nil {
		panic(err)
	}
	if err := validatorWidget.Write("List available validators.\n\n"); err != nil {
		panic(err)
	}

	// Creates Validators widget
	gasMaxWidget, err := text.New(text.RollContent(), text.WrapAtWords())
	if err != nil {
		panic(err)
	}
	if err := gasMaxWidget.Write("How much gas.\n\n"); err != nil {
		panic(err)
	}

	gasAvgBlockWidget, err := text.New(text.RollContent(), text.WrapAtWords())
	if err != nil {
		panic(err)
	}
	if err := gasAvgBlockWidget.Write("How much gas.\n\n"); err != nil {
		panic(err)
	}

	gasAvgTransactionWidget, err := text.New(text.RollContent(), text.WrapAtWords())
	if err != nil {
		panic(err)
	}
	if err := gasAvgTransactionWidget.Write("How much gas.\n\n"); err != nil {
		panic(err)
	}

	latestGasWidget, err := text.New(text.RollContent(), text.WrapAtWords())
	if err != nil {
		panic(err)
	}
	if err := latestGasWidget.Write("How much gas.\n\n"); err != nil {
		panic(err)
	}

	// BIG WIDGETS

	// Block Status Donut widget
	green, err := donut.New(
		donut.CellOpts(cell.FgColor(cell.ColorGreen)),
		donut.Label("New Block Status", cell.FgColor(cell.ColorGreen)),
	)
	if err != nil {
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
	
	// Create Blocks parsing widget
	blocksWidget, err := text.New(text.RollContent(), text.WrapAtWords())
	if err != nil {
		panic(err)
	}
	if err := blocksWidget.Write(networkStatus.Get("result.sync_info.latest_block_height").String() + "\n"); err != nil {
		panic(err)
	}


	// END INITIALISING WIDGETS

	// The functions that execute the updating widgets.

	// system powered widgets
	go writeTime(ctx, info, timeWidget, 1*time.Second)

	// rpc widgets
	go writePeers(ctx, peerWidget, 1*time.Second)
	go writeHealth(ctx, healthWidget, 500*time.Millisecond, connectionSignal)
	go writeSecondsPerBlock(ctx, info, secondsPerBlockWidget, 1*time.Second)
	go writeAmountValidators(ctx, validatorWidget, 1000*time.Millisecond, connectionSignal)
	go writeGasWidget(ctx, info, gasMaxWidget, gasAvgBlockWidget, gasAvgTransactionWidget, latestGasWidget, 1000*time.Millisecond, connectionSignal, genesisInfo)

	// websocket powered widgets
	go writeBlocks(ctx, info, blocksWidget, connectionSignal)
	go writeTransactions(ctx, info, transactionWidget, connectionSignal)
	go writeBlockDonut(ctx, green, 0, 20, 1000*time.Millisecond, playTypePercent, connectionSignal)

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
												container.BorderTitle("Latest Block"),
												container.PlaceWidget(blocksWidget),
												
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
												container.Border(linestyle.Light),
												container.BorderTitle("s Between Blocks"),
												container.PlaceWidget(secondsPerBlockWidget),
											),
											container.Right(
												container.Border(linestyle.Light),
												container.BorderTitle("Validators"),
												container.PlaceWidget(validatorWidget),
											),
										),
									),
								),
							),
						),
					),
					container.Right(
						container.Border(linestyle.Light),
						container.BorderTitle("Current Block round"),
						container.PlaceWidget(green),
					),
				),
			),
			container.Bottom(
				container.SplitVertical(
					
					container.Left(
						container.SplitHorizontal(
							container.Top(
								container.SplitVertical(
									container.Left(
										container.SplitVertical(
											container.Left(
												container.Border(linestyle.Light),
												container.BorderTitle("Gas Max"),
												container.PlaceWidget(gasMaxWidget),
											),
											container.Right(
												container.Border(linestyle.Light),
												container.BorderTitle("Gas Ø per Block"),
												container.PlaceWidget(gasAvgBlockWidget),
											),
										),
										
									),
									container.Right(
										container.SplitVertical(
											container.Left(
												container.Border(linestyle.Light),
												container.BorderTitle("Gas Ø per Tx"),
												container.PlaceWidget(gasAvgTransactionWidget),
											),
											container.Right(
												container.Border(linestyle.Light),
												container.BorderTitle("Gas Latest Tx"),
												container.PlaceWidget(latestGasWidget),
											),
										),
										
									),
								),
							),
							container.Bottom(
								//empty
							),
						),
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
			info.blocks.secondsPassed++
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

// writeAmountValidators writes the status to the healthWidget.
// Exits when the context expires.
func writeAmountValidators(ctx context.Context, t *text.Text, delay time.Duration, connectionSignal chan string) {
	reconnect := false
	validators := gjson.Get(getFromRPC("validators"), "result")
	t.Reset()
	if validators.Exists() {
		t.Write("0")
	} else {
		t.Write("0")
	}

	ticker := time.NewTicker(delay)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			validators := gjson.Get(getFromRPC("validators"), "result")
			if validators.Exists() {
				t.Reset()
				t.Write(validators.Get("total").String())
				if reconnect == true {
					connectionSignal <- "reconnect"
					connectionSignal <- "reconnect"
					connectionSignal <- "reconnect"
					reconnect = false
				}
			} else {
				t.Reset()
				t.Write("0")
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

// writeGasWidget writes the status to the healthWidget.
// Exits when the context expires.
func writeGasWidget(ctx context.Context, info Info, tMax *text.Text, tAvgBlock *text.Text, tAvgTx *text.Text, tLatest *text.Text, delay time.Duration, connectionSignal chan string, genesisInfo gjson.Result) {
	tMax.Write("0")
	tAvgBlock.Write("0")
	tLatest.Write("0")
	tAvgTx.Write("0")

	ticker := time.NewTicker(delay)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			tMax.Reset()
			tAvgBlock.Reset()
			tAvgTx.Reset()
			tLatest.Reset()

			totalGasWanted := uint64(info.blocks.totalGasWanted)
			totalBlocks := uint64(info.blocks.amount)
			totalGasPerBlock := uint64(0)

			// don't divide by 0
			if(totalBlocks > 0) {
				totalGasPerBlock = uint64( totalGasWanted / totalBlocks )
			}


			totalTransactions := uint64(info.transactions.amount)

			// don't divide by 0
			averageGasPerTx := uint64(0)
			if(totalTransactions > 0) {
				averageGasPerTx = uint64( totalGasWanted / info.transactions.amount)
			}
			

			maxGas := genesisInfo.Get("result.genesis.consensus_params.block.max_gas").Int()

			tMax.Write(fmt.Sprintf("%v", numberWithComma(maxGas)))
			tAvgBlock.Write(fmt.Sprintf("%v", numberWithComma(int64(totalGasPerBlock))))
			tLatest.Write(fmt.Sprintf("%v", numberWithComma(info.blocks.lastTx)))
			tAvgTx.Write(fmt.Sprintf("%v", numberWithComma(int64(averageGasPerTx))))
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
			if(info.blocks.secondsPassed != 0) {
				blocksPerSecond = float64(info.blocks.secondsPassed) / float64(info.blocks.amount)
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
func writeTransactions(ctx context.Context, info Info, t *text.Text, connectionSignal <-chan string) {
	port := *givenPort
	socket := gowebsocket.New("ws://localhost:" + port + "/websocket")

	socket.OnTextMessage = func(message string, socket gowebsocket.Socket) {
		currentTx := gjson.Get(message, "result.data.value.TxResult.result.log")
		currentTime := time.Now()
		if currentTx.String() != "" {
			if err := t.Write(fmt.Sprintf("%s\n", currentTime.Format("2006-01-02 03:04:05 PM")+"\n"+currentTx.String())); err != nil {
				panic(err)
			}

			info.blocks.totalGasWanted = info.blocks.totalGasWanted + gjson.Get(message, "result.data.value.TxResult.result.gas_wanted").Int()
			info.blocks.lastTx = gjson.Get(message, "result.data.value.TxResult.result.gas_wanted").Int()
			info.transactions.amount++
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
				writeTransactions(ctx, info, t, connectionSignal)
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
			t.Reset()
			err := t.Write(fmt.Sprintf("%v", numberWithComma(int64(currentBlock.Int())))); 
			if err != nil {
				panic(err)
			}
			info.blocks.amount++
			info.blocks.maxGasWanted = gjson.Get(message, "result.data.value.result_end_block.consensus_param_updates.block.max_gas").Int()
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

// writeBlockDonut continuously changes the displayed percent value on the donut by the
// step once every delay. Exits when the context expires.
func writeBlockDonut(ctx context.Context, d *donut.Donut, start, step int, delay time.Duration, pt playType, connectionSignal <-chan string) {
	port := *givenPort
	socket := gowebsocket.New("ws://localhost:" + port + "/websocket")

	socket.OnTextMessage = func(message string, socket gowebsocket.Socket) {
		step := gjson.Get(message, "result.data.value.step")
		progress := 0

		if step.String() == "RoundStepNewHeight" {
			progress = 100
		}

		if step.String() == "RoundStepCommit" {
			progress = 80
		}

		if step.String() == "RoundStepPrecommit" {
			progress = 60
		}

		if step.String() == "RoundStepPrevote" {
			progress = 40
		}

		if step.String() == "RoundStepPropose" {
			progress = 20
		}


		if err := d.Percent(progress); err != nil {
			panic(err)
		}

	}

	socket.Connect()

	socket.SendText("{ \"jsonrpc\": \"2.0\", \"method\": \"subscribe\", \"params\": [\"tm.event='NewRoundStep'\"], \"id\": 3 }")

	for {
		select {
		case s := <-connectionSignal:
			if s == "no_connection" {
				socket.Close()
			}
			if s == "reconnect" {
				writeBlockDonut(ctx, d, start, step, delay, pt, connectionSignal)
			}
		case <-ctx.Done():
			log.Println("interrupt")
			socket.Close()
			return
		}
	}
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

// UTIL FUNCTIONS

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

func numberWithComma(n int64) string {
    in := strconv.FormatInt(n, 10)
    numOfDigits := len(in)
    if n < 0 {
        numOfDigits-- // First character is the - sign (not a digit)
    }
    numOfCommas := (numOfDigits - 1) / 3

    out := make([]byte, len(in)+numOfCommas)
    if n < 0 {
        in, out[0] = in[1:], '-'
    }

    for i, j, k := len(in)-1, len(out)-1, 0; ; i, j = i-1, j-1 {
        out[j] = in[i]
        if i == 0 {
            return string(out)
        }
        if k++; k == 3 {
            j, k = j-1, 0
            out[j] = ','
        }
    }
}
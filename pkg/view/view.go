package view

import (
	"context"
	"time"

	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/termbox"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/donut"
	"github.com/mum4k/termdash/widgets/text"
)

// Info describes a list of types with data that are used in the explorer
type Info struct {
	blocks       Blocks
	transactions Transactions
}

// Blocks describe content that gets parsed for block
type Blocks struct {
	amount               int
	secondsPassed        int
	totalGasWanted       int64
	gasWantedLatestBlock int64
	maxGasWanted         int64
	lastTx               int64
}

// Transactions describe content that gets parsed for transactions
type Transactions struct {
	amount uint64
}

// Widget widget holder.
type Widget struct {
	terminal          *termbox.Terminal
	container         *container.Container
	currentNetwork    *text.Text
	health            *text.Text
	time              *text.Text
	peer              *text.Text
	secondsPerBlock   *text.Text
	maxBlockSize      *text.Text
	validator         *text.Text
	gasMax            *text.Text
	gasAvgBlock       *text.Text
	gasAvgTransaction *text.Text
	latestGas         *text.Text
	transaction       *text.Text
	blocks            *text.Text
	green             *donut.Donut
}

func DrawView() (widget *Widget, err error) {
	// Initialize widgets
	// Creates Network Widget
	if widget.currentNetwork, err = text.New(text.RollContent(), text.WrapAtWords()); err != nil {
		return widget, err
	}
	if err := widget.currentNetwork.Write("⌛ loading"); err != nil {
		return widget, err
	}

	// Creates Health Widget
	if widget.health, err = text.New(); err != nil {
		return widget, err
	}
	if err := widget.health.Write("⌛ loading"); err != nil {
		return widget, err
	}

	// Creates System Time Widget
	if widget.time, err = text.New(); err != nil {
		return widget, err
	}
	currentTime := time.Now()
	if err := widget.time.Write(currentTime.Format("2006-01-02\n03:04:05 PM\n")); err != nil {
		return widget, err
	}

	// Creates Connected Peers Widget
	if widget.peer, err = text.New(); err != nil {
		return widget, err
	}
	if err := widget.peer.Write("0"); err != nil {
		return widget, err
	}

	// Creates Seconds Between Blocks Widget
	if widget.secondsPerBlock, err = text.New(text.RollContent(), text.WrapAtWords()); err != nil {
		return widget, err
	}
	if err := widget.secondsPerBlock.Write("0"); err != nil {
		return widget, err
	}

	// Creates Max Block Size Widget
	if widget.maxBlockSize, err = text.New(); err != nil {
		return widget, err
	}
	if err := widget.maxBlockSize.Write("0"); err != nil {
		return widget, err
	}

	// Creates Validators widget
	if widget.validator, err = text.New(text.RollContent(), text.WrapAtWords()); err != nil {
		return widget, err
	}
	if err := widget.validator.Write("List available validators.\n\n"); err != nil {
		return widget, err
	}

	// Creates Validators widget
	if widget.gasMax, err = text.New(text.RollContent(), text.WrapAtWords()); err != nil {
		return widget, err
	}
	if err := widget.gasMax.Write("How much gas.\n\n"); err != nil {
		return widget, err
	}

	// Creates Gas per Average Block Widget
	if widget.gasAvgBlock, err = text.New(text.RollContent(), text.WrapAtWords()); err != nil {
		return widget, err
	}
	if err := widget.gasAvgBlock.Write("How much gas.\n\n"); err != nil {
		return widget, err
	}

	// Creates Gas per Average Transaction Widget
	if widget.gasAvgTransaction, err = text.New(text.RollContent(), text.WrapAtWords()); err != nil {
		return widget, err
	}
	if err := widget.gasAvgTransaction.Write("How much gas.\n\n"); err != nil {
		return widget, err
	}

	// Creates Gas per Latest Transaction Widget
	if widget.latestGas, err = text.New(text.RollContent(), text.WrapAtWords()); err != nil {
		return widget, err
	}
	if err := widget.latestGas.Write("How much gas.\n\n"); err != nil {
		return widget, err
	}

	// Add big widgets

	// Block Status Donut widget
	if widget.green, err = donut.New(
		donut.CellOpts(cell.FgColor(cell.ColorGreen)),
		donut.Label("New Block Status", cell.FgColor(cell.ColorGreen)),
	); err != nil {
		return widget, err
	}

	// Transaction parsing widget
	if widget.transaction, err = text.New(text.RollContent(), text.WrapAtWords()); err != nil {
		return widget, err
	}
	if err := widget.transaction.Write("Transactions will appear as soon as they are confirmed in a block.\n\n"); err != nil {
		return widget, err
	}

	// Create Blocks parsing widget
	if widget.blocks, err = text.New(text.RollContent(), text.WrapAtWords()); err != nil {
		return widget, err
	}
	if err := widget.blocks.Write("⌛ loading"); err != nil {
		return widget, err
	}

	if widget.terminal, err = termbox.New(); err != nil {
		return widget, err
	}
	return widget, widget.drawDashboard()
}

func (w Widget) drawDashboard() (err error) {
	w.container, err = container.New(
		w.terminal,
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
												container.PlaceWidget(w.currentNetwork),
											),
											container.Right(
												container.Border(linestyle.Light),
												container.BorderTitle("Health"),
												container.PlaceWidget(w.health),
											),
										),
									),
									container.Right(
										container.SplitVertical(
											container.Left(
												container.Border(linestyle.Light),
												container.BorderTitle("System Time"),
												container.PlaceWidget(w.time),
											),
											container.Right(
												container.Border(linestyle.Light),
												container.BorderTitle("Connected Peers"),
												container.PlaceWidget(w.peer),
											),
										),
									),
								),
							),
							container.Bottom(
								// Insert new bottom rows
								container.SplitVertical(
									container.Left(
										container.SplitVertical(
											container.Left(
												container.Border(linestyle.Light),
												container.BorderTitle("Latest Block"),
												container.PlaceWidget(w.blocks),
											),
											container.Right(
												container.Border(linestyle.Light),
												container.BorderTitle("Max Block Size"),
												container.PlaceWidget(w.maxBlockSize),
											),
										),
									),
									container.Right(
										container.SplitVertical(
											container.Left(
												container.Border(linestyle.Light),
												container.BorderTitle("s Between Blocks"),
												container.PlaceWidget(w.secondsPerBlock),
											),
											container.Right(
												container.Border(linestyle.Light),
												container.BorderTitle("Validators"),
												container.PlaceWidget(w.validator),
											),
										),
									),
								),
							),
						),
					),
					container.Right(
						container.Border(linestyle.Light),
						container.BorderTitle("Current Block Round"),
						container.PlaceWidget(w.green),
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
												container.PlaceWidget(w.gasMax),
											),
											container.Right(
												container.Border(linestyle.Light),
												container.BorderTitle("Gas Ø Block"),
												container.PlaceWidget(w.gasAvgBlock),
											),
										),
									),
									container.Right(
										container.SplitVertical(
											container.Left(
												container.Border(linestyle.Light),
												container.BorderTitle("Gas Ø Tx"),
												container.PlaceWidget(w.gasAvgTransaction),
											),
											container.Right(
												container.Border(linestyle.Light),
												container.BorderTitle("Gas Latest Tx"),
												container.PlaceWidget(w.latestGas),
											),
										),
									),
								),
							),
							container.Bottom(
							// empty
							),
						),
					), container.Right(
						container.Border(linestyle.Light),
						container.BorderTitle("Latest Confirmed Transactions"),
						container.PlaceWidget(w.transaction),
					),
				),
			),
		),
	)
	return err
}

func (w Widget) Cleanup() {
	w.terminal.Close()
}

func (w Widget) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	quitter := func(k *terminalapi.Keyboard) {
		if k.Key == 'q' || k.Key == 'Q' || k.Key == keyboard.KeyEsc {
			cancel()
		}
	}

	return termdash.Run(ctx, w.terminal, w.container, termdash.KeyboardSubscriber(quitter))
}

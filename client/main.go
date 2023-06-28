package main

import (
	"fmt"
	"log"
	"net"

	"github.com/OVantsevich/faraway-test/protocol"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/OVantsevich/faraway-test/cient/internal/config"
)

var app = tview.NewApplication()
var text = tview.NewTextView().
	SetTextColor(tcell.ColorGreen).
	SetText("(g) to get random quote \n(q) to quit")
var quote = tview.NewTextView().
	SetTextColor(tcell.ColorBisque)
var flex = tview.NewFlex()

func main() {
	cfg, err := config.New()
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.Dial("tcp", fmt.Sprint(cfg.ServerHost, ":", cfg.ServerPort))
	if err != nil {
		log.Fatal(err)
	}
	client, err := protocol.NewClient(conn)
	if err != nil {
		log.Fatal(err)
	}

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 113 {
			app.Stop()
		} else if event.Rune() == 97 {
			quoteText, err := client.GetQuote()
			if err != nil {
				quote.SetText(err.Error())
			} else {
				quote.SetText(quoteText)
			}
			go app.Draw()
		}
		return event
	})
	flex.SetDirection(tview.FlexRow).AddItem(
		tview.NewFlex().AddItem(text, 0, 1, false).
			AddItem(quote, 0, 4, false), 0, 1, true)

	if err := app.SetRoot(flex, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}

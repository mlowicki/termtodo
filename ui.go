package main

import (
	"errors"
	"regexp"
	"strconv"
	"time"

	blinkt "github.com/alexellis/blinkt_go"
	"github.com/mattn/go-runewidth"
	"github.com/nsf/termbox-go"
)

type UI struct {
	cl        *CommandLine
	Scheduler *Scheduler
	todos     []Todo
	alertCh   chan struct{}
}

func NewUI(scheduler *Scheduler) *UI {
	err := termbox.Init()
	if err != nil {
		panic(err) // TODO more desciptive message
	}
	ui := UI{cl: &CommandLine{}, Scheduler: scheduler}
	go func() {
		for {
			select {
			case todos := <-scheduler.TodosCh:
				ui.todos = todos
				ui.Redraw()
			}
		}
	}()
	return &ui
}

func (ui *UI) print(x, y int, text string) {
	col := termbox.ColorDefault
	for _, r := range text {
		termbox.SetCell(x, y, r, col, col)
		x += runewidth.RuneWidth(r)
	}
}

func (ui *UI) Redraw() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	for i, todo := range ui.todos {
		ui.print(0, i, todo.Name)
	}
	if len(ui.todos) > 0 {
		if ui.alertCh == nil {
			ui.alertCh = alert()
		}
	} else {
		ui.disableAlert()
	}
	ui.cl.Redraw()
	termbox.Flush()
}
func (ui *UI) disableAlert() {
	if ui.alertCh == nil {
		return
	}
	ui.alertCh <- struct{}{}
	<-ui.alertCh
	ui.alertCh = nil
}

func (ui *UI) Close() {
	ui.disableAlert()
	termbox.Close()
}

func (ui *UI) parseTime(input []byte) (time.Duration, error) {
	re := regexp.MustCompile(`(\d+)([smhd])`)
	match := re.FindSubmatch(input)
	if match == nil {
		return 0, errors.New("invalid time format")
	}
	num, err := strconv.Atoi(string(match[1]))
	if err != nil {
		return 0, err
	}
	var unit time.Duration
	switch string(match[2]) {
	case "s":
		unit = time.Second
	case "m":
		unit = time.Minute
	case "h":
		unit = time.Hour
	case "d":
		unit = time.Hour * 24
	default:
		return 0, errors.New("invalid time unit")
	}
	return unit * time.Duration(num), nil
}

// TODO display error to user instead panicking
func (ui *UI) HandleCommand(tokens []string) {
	switch tokens[0] {
	case "add":
		if len(tokens) < 3 {
			panic("add: not enough arguments")
		}
		t, err := ui.parseTime([]byte(tokens[1]))
		if err != nil {
			trigger, err := NewTrigger(
				tokens[2],
				tokens[1],
				time.Now(),
				-1, // trigger indefinitely.
			)
			if err != nil {
				panic(err)
			}
			ui.Scheduler.AddTriggerCh <- trigger
			return
		}
		trigger, err := NewTrigger(
			tokens[2],
			"*/1 * * * * *",
			time.Now().Add(t),
			1, // one-time trigger.
		)
		if err != nil {
			panic(err)
		}
		ui.Scheduler.AddTriggerCh <- trigger
	case "done":
		if len(ui.todos) > 0 {
			ui.Scheduler.DelTodoCh <- ui.todos[0].ID
		}
	}
}

func (ui *UI) Run() {
	ui.Redraw()
	for command := range ui.cl.Run() {
		ui.HandleCommand(command)
	}
}

func alert() chan struct{} {
	ch := make(chan struct{})
	go func() {
		brightness := 0.5
		bl := blinkt.NewBlinkt(brightness)
		bl.Setup()
		r := 150
		g := 0
		b := 0
	outerloop:
		for {
			for pixel := 0; pixel < 8; pixel++ {
				select {
				case <-ch:
					break outerloop
				default:
				}
				bl.Clear()
				bl.SetPixel(pixel, r, g, b)
				bl.Show()
				blinkt.Delay(100)
			}
			for pixel := 7; pixel > 0; pixel-- {
				select {
				case <-ch:
					break outerloop
				default:
				}
				bl.Clear()
				bl.SetPixel(pixel, r, g, b)
				bl.Show()
				blinkt.Delay(100)
			}
		}
		bl.Clear()
		bl.Show()
		close(ch)
	}()
	return ch
}

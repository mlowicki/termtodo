package main

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"syscall"
	"time"

	"github.com/mattn/go-runewidth"
	"github.com/nsf/termbox-go"
)

type UI struct {
	cl        *CommandLine
	Scheduler *Scheduler
	todos     []Todo
	blinkt    *Blinkt
	err       error
	cancelErr func()
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

// showErr displays error message to user.
func (ui *UI) showErr(err error) {
	if ui.cancelErr != nil {
		ui.cancelErr()
	}
	termbox.Flush()
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		select {
		case <-time.After(time.Second * 5):
			ui.clearErr()
			cancel()
		case <-ctx.Done():
		}
	}()
	ui.cancelErr = cancel
	ui.err = err
	ui.Redraw()
}

// clearErr hides error message.
func (ui *UI) clearErr() {
	w, h := termbox.Size()
	fill(0, h-2, w, 1, termbox.Cell{Ch: ' '})
	ui.err = nil
	termbox.Flush()
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
		ui.print(0, i, fmt.Sprintf("%d. %s", i+1, todo.Name))
	}
	if len(ui.todos) > 0 {
		if ui.blinkt == nil {
			ui.blinkt = NewBlinkt()
		}
	} else {
		if ui.blinkt != nil {
			ui.blinkt.Stop()
			ui.blinkt = nil
		}
	}
	if ui.err != nil {
		_, h := termbox.Size()
		ui.print(0, h-2, ui.err.Error())
	}
	ui.cl.Redraw()
	termbox.Flush()
}

func (ui *UI) Close() {
	if ui.blinkt != nil {
		ui.blinkt.Stop()
		ui.blinkt = nil
	}
	termbox.Close()
}

var errInvalidTime = errors.New("invalid time")

func (ui *UI) parseTime(input []byte) (time.Time, error) {
	if len(input) == 0 {
		return time.Time{}, errInvalidTime
	}
	if input[0] == '+' {
		re := regexp.MustCompile(`(\d+)([smhd])`)
		match := re.FindSubmatch(input[1:])
		if match == nil {
			return time.Time{}, errInvalidTime
		}
		num, err := strconv.Atoi(string(match[1]))
		if err != nil {
			return time.Time{}, err
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
			return time.Time{}, errInvalidTime
		}
		return time.Now().Add(unit * time.Duration(num)), nil
	} else if input[0] == '@' {
		now := time.Now()
		t, err := time.ParseInLocation("15:04", string(input[1:]), now.Location())
		t = t.AddDate(now.Year(), int(now.Month())-1, now.Day()-1)
		return t, err
	}
	return time.Time{}, errInvalidTime
}

func (ui *UI) getIdx(token string) (int, error) {
	idx, err := strconv.Atoi(token)
	if err != nil {
		return -1, fmt.Errorf("invalid index: %w", err)
	}
	if idx < 1 || idx > len(ui.todos) {
		return -1, errors.New("index out of range")
	}
	return idx, nil
}

func (ui *UI) HandleCommand(tokens []string) {
	ui.clearErr()
	switch tokens[0] {
	case "add":
		if len(tokens) < 3 {
			ui.showErr(errors.New("not enough arguments"))
			return
		}
		var trigger Trigger
		t, err := ui.parseTime([]byte(tokens[1]))
		if err != nil {
			trigger, err = NewTrigger(
				tokens[2],
				tokens[1],
				time.Now(),
				-1, // trigger indefinitely.
			)
			if err != nil {
				ui.showErr(err)
				return
			}
		} else {
			trigger, err = NewTrigger(
				tokens[2],
				"*/1 * * * * *",
				t,
				1, // one-time trigger.
			)
			if err != nil {
				ui.showErr(err)
				return
			}
		}
		ui.Scheduler.AddTriggerCh <- trigger
	case "done":
		idx := 1
		if len(tokens) > 1 {
			var err error
			idx, err = ui.getIdx(tokens[1])
			if err != nil {
				ui.showErr(err)
				return
			}
		}
		ui.Scheduler.DelTodoCh <- ui.todos[idx-1].ID
	case "snooze":
		if len(tokens) < 2 {
			ui.showErr(errors.New("not enough arguments"))
			return
		}
		t, err := ui.parseTime([]byte(tokens[1]))
		if err != nil {
			ui.showErr(err)
			return
		}
		idx := 1
		if len(tokens) > 2 {
			var err error
			idx, err = ui.getIdx(tokens[2])
			if err != nil {
				ui.showErr(err)
				return
			}
		}
		todo := ui.todos[idx-1]
		trigger, err := NewTrigger(
			todo.Name,
			"*/1 * * * * *",
			t,
			1, // one-time trigger.
		)
		if err != nil {
			ui.showErr(err)
			return
		}
		ui.Scheduler.DelTodoCh <- todo.ID
		ui.Scheduler.AddTriggerCh <- trigger
	default:
		err := errors.New("unknown command: " + tokens[0])
		ui.showErr(err)
	}
}

func (ui *UI) Run() {
	ui.Redraw()
	eventsCh := make(chan termbox.Event)
	go func() {
		for {
			switch ev := termbox.PollEvent(); ev.Type {
			case termbox.EventKey:
				switch ev.Key {
				case termbox.KeyCtrlZ:
					syscall.Kill(syscall.Getpid(), syscall.SIGSTOP)
				default:
					eventsCh <- ev
				}
			case termbox.EventError:
				panic(ev.Err)
			case termbox.EventResize:
				ui.Redraw()
			}
		}
	}()
	for command := range ui.cl.Run(eventsCh) {
		ui.HandleCommand(command)
	}
}

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

type View int

const (
	TODOS View = iota
	TRIGGERS
)

type UI struct {
	cl        *CommandLine
	Scheduler *Scheduler
	todos     []Todo
	triggers  []Trigger
	blinkt    *Blinkt
	err       error
	cancelErr func()
	view      View
}

func NewUI(scheduler *Scheduler) *UI {
	err := termbox.Init()
	if err != nil {
		panic(err) // TODO more desciptive message
	}
	ui := UI{cl: &CommandLine{}, Scheduler: scheduler, view: TODOS}
	go func() {
		for {
			select {
			case todos := <-scheduler.TodosCh:
				ui.todos = todos
				ui.Redraw()
			case triggers := <-scheduler.TriggersCh:
				ui.triggers = triggers
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
	switch ui.view {
	case TODOS:
		for i, todo := range ui.todos {
			ui.print(0, i, fmt.Sprintf("%*d %s", -len(strconv.Itoa(len(ui.todos))), i+1, todo.Name))
		}
	case TRIGGERS:
		maxName := 0
		for _, trigger := range ui.triggers {
			w := len(trigger.Name)
			if w > maxName {
				maxName = w
			}
		}
		for i, trigger := range ui.triggers {
			when := trigger.Cron
			if trigger.Count != -1 {
				when = trigger.After.Format("Mon Jan 2 15:04:05")
			}
			ui.print(0, i, fmt.Sprintf("%*d %*s %s", -len(strconv.Itoa(len(ui.triggers))), i+1, -maxName, trigger.Name, when))
		}
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

func (ui *UI) getIdxs(token string) ([]int, error) {
	if token == "*" {
		length := 0
		switch ui.view {
		case TODOS:
			length = len(ui.todos)
		case TRIGGERS:
			length = len(ui.triggers)
		default:
			panic("view not supported")
		}
		idxs := make([]int, length, length)
		for i := 0; i < length; i++ {
			idxs[i] = i
		}
		return idxs, nil
	}
	idx, err := strconv.Atoi(token)
	if err != nil {
		return nil, fmt.Errorf("invalid index: %w", err)
	}
	length := 0
	switch ui.view {
	case TODOS:
		length = len(ui.todos)
	case TRIGGERS:
		length = len(ui.triggers)
	default:
		panic("view not supported")
	}
	if idx < 1 || idx > length {
		return nil, errors.New("index out of range")
	}
	return []int{idx - 1}, nil
}

func (ui *UI) HandleCommand(tokens []string) {
	ui.clearErr()
	switch tokens[0] {
	case "a", "add":
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
		ui.Scheduler.AddTriggersCh <- []Trigger{trigger}
	case "r", "rm":
		if len(tokens) == 1 {
			tokens = append(tokens, "1")
		}
		idxs, err := ui.getIdxs(tokens[1])
		if err != nil {
			ui.showErr(err)
			return
		}
		switch ui.view {
		case TODOS:
			ids := make([]string, 0, len(idxs))
			for _, idx := range idxs {
				ids = append(ids, ui.todos[idx].ID)
			}
			ui.Scheduler.DelTodosCh <- ids
		case TRIGGERS:
			ids := make([]string, 0, len(idxs))
			for _, idx := range idxs {
				ids = append(ids, ui.triggers[idx].ID)
			}
			ui.Scheduler.DelTriggersCh <- ids
		default:
			panic("view not supported")
		}
	case "s", "snooze":
		if ui.view != TODOS {
			ui.showErr(errors.New("invalid command"))
			return
		}
		if len(tokens) < 2 {
			ui.showErr(errors.New("not enough arguments"))
			return
		}
		t, err := ui.parseTime([]byte(tokens[1]))
		if err != nil {
			ui.showErr(err)
			return
		}
		if len(tokens) == 2 {
			tokens = append(tokens, "1")
		}
		idxs, err := ui.getIdxs(tokens[2])
		if err != nil {
			ui.showErr(err)
			return
		}
		todos := make([]string, 0, len(idxs))
		triggers := make([]Trigger, 0, len(idxs))
		for _, idx := range idxs {
			todo := ui.todos[idx]
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
			todos = append(todos, todo.ID)
			triggers = append(triggers, trigger)
		}
		ui.Scheduler.DelTodosCh <- todos
		ui.Scheduler.AddTriggersCh <- triggers
	case "tr", "triggers":
		ui.view = TRIGGERS
		ui.Redraw()
	case "to", "todos":
		ui.view = TODOS
		ui.Redraw()
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
				case termbox.KeyEsc, termbox.KeyCtrlC:
					syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
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
	commandsCh := ui.cl.Run(eventsCh)
	for {
		command := <-commandsCh
		if command[0] == "q" || command[0] == "quit" {
			break
		}
		ui.HandleCommand(command)
	}
}

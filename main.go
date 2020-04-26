package main

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	cron "github.com/robfig/cron/v3"
)

type Todo struct {
	Name string
	ID   string
}

type Trigger struct {
	Name  string
	Cron  string
	After time.Time
	Count int
}

func NewTrigger(name, cron string, after time.Time, count int) (Trigger, error) {
	t := Trigger{
		Name:  name,
		Cron:  cron,
		After: after,
		Count: count,
	}
	_, err := t.Schedule()
	return t, err
}

func (t *Trigger) Schedule() (cron.Schedule, error) {
	parser := cron.NewParser(cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	sched, err := parser.Parse(t.Cron)
	if err != nil {
		return nil, fmt.Errorf("invalid schedule: %w", err)
	}
	return sched, nil
}

func (t *Trigger) Next() time.Time {
	if t.Count == 0 {
		return time.Time{}
	}
	sch, err := t.Schedule()
	if err != nil {
		panic(err)
	}
	return sch.Next(t.After)
}

func (t *Trigger) Check() *Todo {
	now := time.Now()
	if t.Count == 0 || t.Next().After(now) {
		return nil
	}
	t.After = now
	if t.Count != -1 {
		t.Count--
	}
	return &Todo{Name: t.Name, ID: uuid.New().String()}
}

type Scheduler struct {
	triggers     []Trigger
	todos        []Todo
	TodosCh      chan []Todo
	AddTriggerCh chan Trigger
	DelTodoCh    chan string
	timer        *time.Timer
}

func (sch *Scheduler) delTodo(id string) {
	for i, todo := range sch.todos {
		if id == todo.ID {
			sch.todos = append(sch.todos[:i], sch.todos[i+1:]...)
		}
	}
	todos := make([]Todo, len(sch.todos))
	copy(todos, sch.todos)
	sch.TodosCh <- todos
}

func (sch *Scheduler) checkTriggers() {
	added := false
	triggers := make([]Trigger, 0)
	for _, trigger := range sch.triggers {
		if todo := trigger.Check(); todo != nil {
			sch.todos = append(sch.todos, *todo)
			added = true
		}
		if !trigger.Next().IsZero() {
			triggers = append(triggers, trigger)
		}
	}
	sch.triggers = triggers
	if added {
		todos := make([]Todo, len(sch.todos))
		copy(todos, sch.todos)
		sch.TodosCh <- todos
	}
}

func NewScheduler() *Scheduler {
	sch := Scheduler{
		triggers:     make([]Trigger, 0),
		todos:        make([]Todo, 0),
		TodosCh:      make(chan []Todo),
		AddTriggerCh: make(chan Trigger),
		DelTodoCh:    make(chan string),
		timer:        time.NewTimer(time.Millisecond),
	}
	go func() {
		for {
			timerExpired := false
			select {
			case id := <-sch.DelTodoCh:
				sch.delTodo(id)
			case trigger := <-sch.AddTriggerCh:
				sch.triggers = append(sch.triggers, trigger)
				sch.checkTriggers()
			case <-sch.timer.C:
				sch.checkTriggers()
				timerExpired = true
			}
			nextCheck := time.Now().Add(time.Hour * 24 * 7)
			for _, trigger := range sch.triggers {
				n := trigger.Next()
				if !n.IsZero() && n.Before(nextCheck) {
					nextCheck = n
				}
			}
			if !timerExpired && !sch.timer.Stop() {
				<-sch.timer.C
			}
			sch.timer.Reset(time.Until(nextCheck))
		}
	}()
	return &sch
}

func main() {
	scheduler := NewScheduler()
	ui := NewUI(scheduler)
	defer ui.Close()
	ui.Run()
}

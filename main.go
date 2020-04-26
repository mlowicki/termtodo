package main

import (
	"time"

	"github.com/google/uuid"
	cron "github.com/robfig/cron/v3"
)

type Todo struct {
	Name string
	ID   string
}

type Trigger interface {
	Check() *Todo
	Next() time.Time
}

type OneTimeTrigger struct {
	Name      string
	When      time.Time
	triggered bool
}

func (t *OneTimeTrigger) Next() time.Time {
	if t.triggered {
		return time.Time{}
	}
	return t.When
}

func (t *OneTimeTrigger) Check() *Todo {
	if t.Next().IsZero() || time.Now().Before(t.When) {
		return nil
	}
	t.triggered = true
	return &Todo{Name: t.Name, ID: uuid.New().String()}
}

type CronTrigger struct {
	Name        string
	Cron        cron.Schedule
	LastTrigger time.Time
}

func (t *CronTrigger) Next() time.Time {
	return t.Cron.Next(t.LastTrigger)
}

func (t *CronTrigger) Check() *Todo {
	now := time.Now()
	if t.Next().After(now) {
		return nil
	}
	t.LastTrigger = now
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

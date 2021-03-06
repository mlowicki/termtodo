package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/nsf/termbox-go"
	cron "github.com/robfig/cron/v3"
)

// Todo represents something to accomplish.
type Todo struct {
	Name      string
	ID        string
	CreatedAt time.Time
}

// Trigger defines when to create a Todo.
type Trigger struct {
	Name  string
	Cron  string
	After time.Time
	Count int
	ID    string
}

func NewTrigger(name, cron string, after time.Time, count int) (Trigger, error) {
	t := Trigger{
		Name:  name,
		Cron:  cron,
		After: after,
		Count: count,
		ID:    uuid.New().String(),
	}
	_, err := t.Schedule() // validate schedule
	return t, err
}

func (t *Trigger) Schedule() (cron.Schedule, error) {
	parser := cron.NewParser(cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
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
	return &Todo{Name: t.Name, ID: uuid.New().String(), CreatedAt: time.Now()}
}

type Scheduler struct {
	TodosCh       chan []Todo
	TriggersCh    chan []Trigger
	AddTriggersCh chan []Trigger
	DelTriggersCh chan []string
	DelTodosCh    chan []string
	timer         *time.Timer
	db            *DB
}

func (sch *Scheduler) checkTriggers() {
	triggers := make(map[string]Trigger)
	for _, trigger := range sch.db.Triggers {
		if todo := trigger.Check(); todo != nil {
			sch.db.Todos[(*todo).ID] = *todo
		}
		if !trigger.Next().IsZero() {
			triggers[trigger.ID] = trigger
		}
	}
	sch.db.Triggers = triggers
	err := sch.db.Write()
	if err != nil {
		panic(err)
	}
}

func (sch *Scheduler) sendTodos() {
	todos := make([]Todo, 0, len(sch.db.Todos))
	for _, todo := range sch.db.Todos {
		todos = append(todos, todo)
	}
	sch.TodosCh <- todos
}

func (sch *Scheduler) sendTriggers() {
	triggers := make([]Trigger, 0, len(sch.db.Triggers))
	for _, trigger := range sch.db.Triggers {
		triggers = append(triggers, trigger)
	}
	sch.TriggersCh <- triggers
}

func NewScheduler(db *DB) *Scheduler {
	sch := Scheduler{
		TodosCh:       make(chan []Todo),
		TriggersCh:    make(chan []Trigger),
		AddTriggersCh: make(chan []Trigger),
		DelTriggersCh: make(chan []string),
		DelTodosCh:    make(chan []string),
		timer:         time.NewTimer(time.Millisecond),
		db:            db,
	}
	go func() {
		sch.sendTodos()
		sch.sendTriggers()
		for {
			timerExpired := false
			triggersNum := len(db.Triggers)
			todosNum := len(db.Todos)
			select {
			case ids := <-sch.DelTodosCh:
				for _, id := range ids {
					delete(db.Todos, id)
				}
				err := sch.db.Write()
				if err != nil {
					panic(err)
				}
			case triggers := <-sch.AddTriggersCh:
				for _, trigger := range triggers {

					db.Triggers[trigger.ID] = trigger
				}
				sch.checkTriggers()
			case ids := <-sch.DelTriggersCh:
				for _, id := range ids {
					delete(db.Triggers, id)
				}
				err := sch.db.Write()
				if err != nil {
					panic(err)
				}
			case <-sch.timer.C:
				sch.checkTriggers()
				timerExpired = true
			}

			if len(db.Todos) != todosNum {
				sch.sendTodos()
			}
			if len(db.Triggers) != triggersNum {
				sch.sendTriggers()
			}

			nextCheck := time.Now().Add(time.Hour * 24 * 7)
			for _, trigger := range sch.db.Triggers {
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
	var dbpath = flag.String("dbpath", ".termtodo.db", "path to database")
	flag.Parse()
	db, err := NewDB(*dbpath)
	if err != nil {
		log.Fatalf("Cannot initialize database: %s", err)
	}
	scheduler := NewScheduler(db)
	ui := NewUI(scheduler)
	defer ui.Close()
	go func() {
		contCh := make(chan os.Signal, 1)
		signal.Notify(contCh, syscall.SIGCONT)
		termCh := make(chan os.Signal, 1)
		signal.Notify(termCh, syscall.SIGTERM)
		for {
			select {
			case <-contCh:
				ui.Close()
				err = termbox.Init()
				if err != nil {
					panic(err)
				}
				ui.Redraw()
			case <-termCh:
				ui.Close()
				os.Exit(0)
			}
		}
	}()
	ui.Run()
}

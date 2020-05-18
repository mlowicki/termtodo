# termtodo

Minimalistic todo app for your terminal.

*Currently supports only Raspbian*.

## Features
* One-time or recursive triggers (using cron syntax) defining when todo should be created
* Integration with [Blinkt!](https://learn.pimoroni.com/tutorial/sandyj/getting-started-with-blinkt) for notifications on Raspberry Pi
* Easy to integrate with terminal multiplexers like tmux
* Efficient command-line interface 
* All data stored locally

## Installation

### From source
1. Install [Go](https://golang.org/doc/install)
2.
```
$ git clone git@github.com:mlowicki/termtodo.git
$ cd termtodo
$ go build
$ ./termtodo
```

## Commands

### a(dd)

Examples:
```
:a +10m "call mom"
:a @9:00 "morning coffee"
:add "0 10 * * 0-5" workout
```
### r(m)
Deletes todo or trigger, depending on the active view. Accepts optional selector to specify the item to remove:
* If the selector is missing, then the first item from the top will be erased.
* If the selector is `*`, then all items will be removed.
* Otherwise selector is interpreted as a number.

Examples:
```
:r
:r 4
:rm 4
:r *
```
### s(nooze)
Postpone todo. Accepts optional selector to specify the todo to re-schedule:
* If the selector is missing, then the first item from the top will be erased.
* If the selector is `*`, then all todos will be removed.
* Otherwise selector is interpreted as a number.


Examples:
```
:s +1h
:s @11:00
:s +1m 4
:snooze +20m 2
:s +20m *
```
### to(dos)
Show things to do (default view).

Examples:
```
:to
:todos
```
### tr(iggers)
Show schedules for todos.

Examples:
```
:tr
:triggers
```
### q(uit), <ctrl-c>
Quit program.

Examples:
```
:q
:quit
:<ctrl-c>
```

## Time formats
...

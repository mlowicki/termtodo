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

Call Mom in 10 minutes:
```
:a +10m "call mom"
```

Drink coffee at 9:00:
```
:a @9:00 "morning coffee"
```

Do your workout at 10:00 every week day:
```
:add "0 10 * * 0-5" workout
```
### r(m)
Deletes todo or trigger, depending on the active view. Accepts optional selector to specify the item to remove:
* If the selector is missing, then the first item from the top will be erased.
* If the selector is `*`, then all items will be removed.
* Otherwise selector is interpreted as a number.

Examples:

Delete first todo from the list (number 1):
```
:r
```

Delete todo #4:
```
:r 4
:rm 4
```

Delete all todos:
```
:r *
```
### s(nooze)
Postpone todo. Accepts optional selector to specify the todo to re-schedule:
* If the selector is missing, then the first item from the top will be erased.
* If the selector is `*`, then all todos will be removed.
* Otherwise selector is interpreted as a number.


Examples:

Snooze first todo for 1h (trigger in 60 minutes):
```
:s +1h
```

Trigger todo again at 11:00:
```
:s @11:00
```

Re-schedule todo #4 in 4 minutes:
```
:s +1m 4
```

Snooze todo #2 for 20 minutes:
```
:snooze +20m 2
```

Snooze all todos for 20 minutes:
```
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
### q(uit) or \<ctrl-c\>
Quit program.

Examples:
```
:q
:quit
:<ctrl-c>
```

## Time formats
...

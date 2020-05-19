# termtodo [![Build Status](https://travis-ci.com/mlowicki/termtodo.svg?token=Wu2ZDxNBSqPxs4JUg6tc&branch=master)](https://travis-ci.com/mlowicki/termtodo)

Minimalistic todo app for your terminal.

![terminal session](/session.gif)

## Features
* One-time or recurring triggers (using cron syntax) defining when todo should be created
* Integration with [Blinkt!](https://learn.pimoroni.com/tutorial/sandyj/getting-started-with-blinkt) for notifications on Raspberry Pi

![Blinkt! alert](/alert.gif)

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
$ go build -tags raspbian // pass this tag only on Raspbian
$ ./termtodo
```

## Commands

### a(dd)
Create a trigger that adds todo either once or regularly.

Call Mom in 10 minutes:
```
:a +10m "call mom"
```

Drink coffee at 9:00:
```
:a @9:00 "coffee with Joe"
```

Do your workout at 10:00 every weekday:
```
:add "0 10 * * 0-5" workout
```


See [Time formats](#time-formats) for a list of all supported formats.

### r(m)
Delete todo or trigger, depending on the active view. Accepts optional selector to specify the item to remove:
* If the selector is missing, then the first item from the top will be erased.
* If the selector is `*`, then all items will be removed.
* Otherwise selector is interpreted as a number.

Delete the first todo from the list (number 1):
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

Snooze the first todo for 1h (trigger in 60 minutes):
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


See [Time formats](#time-formats) for a list of all supported formats. Snooze command supports one-time triggers only, so it doesn't support cron format.

### to(dos)
Show things to do (default view).

```
:to
:todos
```
### tr(iggers)
Show schedules for todos.

```
:tr
:triggers
```
### q(uit) or \<ctrl-c\>
Quit program.

```
:q
:quit
:<ctrl-c>
```

## Time formats

### Relative time
```
\+(\d+)[smhd]
```

Where `smhd` stands for seconds, minutes, hours and days, respectively.

In 10 seconds:
```
+10s
```

In 4 days:
```
+4d
```

### Absolute time

At a specific time of the current day:
```
@\d{2}:\d{2}
```

Examples:
```
@10:00
@23:15
@9:10
```

### Cron

Define recurring event. It's implemented by:
```
cron.NewParser(cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
```
(see https://godoc.org/github.com/robfig/cron#hdr-Alternative_Formats for more details)

At 10:00 every weekday:
```
0 10 * * 0-5
```

Every 10 seconds:
```
*/10 * * * * *
```
(adding seconds is an extension, so it requires an additional 6th field)

Every hour:
```
@hourly
@every 1h
```

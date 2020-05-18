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
Create a trigger that adds todo either once or regularly.

Examples:

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
Deletes todo or trigger, depending on the active view. Accepts optional selector to specify the item to remove:
* If the selector is missing, then the first item from the top will be erased.
* If the selector is `*`, then all items will be removed.
* Otherwise selector is interpreted as a number.

Examples:

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


Examples:

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

### Relative time
```
\+(\d+)[smhd]
```

Where `smhd` stands for seconds, minutes, hours and days, respectively.

Examples:

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

Allows to great recurring triggers. It's implemented by:
```
cron.NewParser(cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
```
see https://godoc.org/github.com/robfig/cron#hdr-Alternative_Formats for more information.

Examples:

Eevery 10 seconds:
```
*/10 * * * * *
```

Every hour:
```
@hourly
@every 1h
```

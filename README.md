# termtodo

Minimalistic todo app for your terminal.

*Currently supports only Raspbian*.

## Features
* One-time or recursive triggers (using cron syntax) 
* Integration with [Blinkt!](https://learn.pimoroni.com/tutorial/sandyj/getting-started-with-blinkt) for Raspberry Pi
* Easy to integrate with terminal multiplexers like tmux
* Efficient command-line interface 
* All data stored locally

## Commands

### a(dd)

Examples:
```
: a +10m "call mom"
: a @9:00 "morning coffee"
: add "0 10 * * 0-5" workout
```
### r(m)
Deletes todo or trigger, depending on the active view. Accepts optional number which specifies item to remove. If number is missing then first item from the top will be erased.

Examples:
```
: r
: r 4
: rm 4
```
### s(nooze)
Postpone todo. Accepts optional number which specifies todo to re-schedule. If number is missing then first todo from the top will be snoozed.

Examples:
```
: s +1h
: s @11:00
: s +1m 4
: snooze +20m 2
```
### to(dos)
Show things to do (default view).

Examples:
```
: to
: todos
```
### tr(iggers)
Show schedules for todos.

Examples:
```
: tr
: triggers
```
### q(uit)
Quit program. \<ESC\> does the same.

Examples:
```
: q
: quit
```

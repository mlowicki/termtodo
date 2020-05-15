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

### add

Examples:
```
: add +10m "call mom"
: add @9:00 "morning coffee"
: add "0 10 * * 0-5" workout
```
### rm
Deletes todo or trigger, depending on the active view. Accepts optional number which specifies item to remove. If number is missing then first item from the top will be erased.

Examples:
```
: rm
: rm 4
```
### snooze
Postpone todo. Accepts optional number which specifies todo to re-schedule. If number is missing then first todo from the top will be snoozed.

Examples:
```
: snooze +1h
: snooze @11:00
: snooze +1m 4
```
### todos
Show things to do (default view).
### triggers
Show schedules for todos.
### exit
Quit program. \<ESC\> does the same.

package main

import (
	"unicode/utf8"

	"github.com/mattn/go-runewidth"
)

func byteSliceGrow(s []byte, desiredCap int) []byte {
	if cap(s) < desiredCap {
		ns := make([]byte, len(s), desiredCap)
		copy(ns, s)
		return ns
	}
	return s
}

func byteSliceRemove(text []byte, from, to int) []byte {
	size := to - from
	copy(text[from:], text[to:])
	text = text[:len(text)-size]
	return text
}

func byteSliceInsert(text []byte, offset int, what []byte) []byte {
	n := len(text) + len(what)
	text = byteSliceGrow(text, n)
	text = text[:n]
	copy(text[offset+len(what):], text[offset:])
	copy(text[offset:], what)
	return text
}

// wcwidth returns number of columns needed to represent text.
func wcwidth(text []byte) int {
	res := 0
	for len(text) > 0 {
		r, size := utf8.DecodeRune(text)
		text = text[size:]
		res += runewidth.RuneWidth(r)
	}
	return res
}

package idfkwimd

import "time"

const (
	Invalid = iota
	NoteOn
	NoteOff
	EOT
)

const (
	NotesCount    = 128
	ChannelsLimit = 16
	MaxY          = 127
	TPS           = 150 // pos(EOT) / тривалість тестового треку, округлено до десятків
	Tick          = time.Second / TPS
)

const (
	WindowWidth  = 128 * 8
	WindowHeight = 128 * 6
)

const Notes = 128
const SoftTime = 450 * time.Millisecond

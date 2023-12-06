package idfkwimd

type MidiEvent struct {
	Type  EventType
	Track int16
	Pos   uint64
	Delta uint32
	Key /*0-127*/, Velocity/*0-127*/ uint8
}

type State struct {
	Notes [Notes]struct {
		Velocity uint8
		Pos      uint64
	}
	ActiveNotes []MidiEvent
}

type EventType int

func (ET EventType) IsValid() bool {
	switch ET {
	case NoteOn, NoteOff, EOT:
		return true
	default:
		return false
	}
}

func (me *MidiEvent) IsValid() bool {
	return me.Type.IsValid()
}

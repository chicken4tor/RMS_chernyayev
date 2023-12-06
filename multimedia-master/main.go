package main

import (
	"fmt"
	et "github.com/hajimehoshi/ebiten/v2" // Пакет для створення ігрового движка
	"gitlab.com/gomidi/midi/reader"       // Пакет для роботи з MIDI-файлами
	"log"
	plrrndr "multi/renderer" // Пакет для рендерингу
	idgafwimd "multi/utils"
	"os"
	"path/filepath"
	"sort"
)

// Структура для розпакування MIDI-подій
type MidiUnmarshaler struct {
	Events   []idgafwimd.MidiEvent // Масив подій MIDI
	DemoMode bool                  // Режим демонстрації
}

// Обробник події "Note On" MIDI
func (mun *MidiUnmarshaler) noteOn(p *reader.Position, channel, key, vel uint8) {
	if mun.DemoMode {
		fmt.Printf("Track: %v Pos: %v NoteOn\t(ch %v: Key %v Velocity: %v)\n", p.Track, p.AbsoluteTicks, channel, key, vel)
	} else {
		// Додаємо подію "Note On" до масиву подій
		mun.Events = append(mun.Events,
			idgafwimd.MidiEvent{
				Type:     idgafwimd.NoteOn,
				Track:    p.Track,
				Pos:      p.AbsoluteTicks,
				Delta:    p.DeltaTicks,
				Key:      key,
				Velocity: vel,
			},
		)
	}
}

// Обробник події "Note Off" MIDI
func (mun *MidiUnmarshaler) noteOff(p *reader.Position, channel, key, vel uint8) {
	if mun.DemoMode {
		fmt.Printf("Track: %v Pos: %v NoteOff\t(ch %v: Key %v)\n", p.Track, p.AbsoluteTicks, channel, key)
	} else {
		// Додаємо подію "Note Off" до масиву подій
		mun.Events = append(mun.Events,
			idgafwimd.MidiEvent{
				Type:  idgafwimd.NoteOff,
				Track: p.Track,
				Pos:   p.AbsoluteTicks,
				Delta: p.DeltaTicks,
				Key:   key,
			},
		)
	}
}

// Отримання відсортованих подій з MIDI-файлу
func GetSortedEvents(f string) MidiUnmarshaler {
	var mun MidiUnmarshaler
	rd := reader.New(
		reader.NoLogger(),
		reader.NoteOn(mun.noteOn),   // Обробник події "Note On"
		reader.NoteOff(mun.noteOff), // Обробник події "Note Off"
		//reader.EndOfTrack(mun.EOT),
	)

	// Зчитування MIDI-файлу
	err := reader.ReadSMFFile(rd, f)
	if err != nil {
		panic(err)
	}
	// Сортування подій за позицією у часі
	sort.Slice(mun.Events, func(i, j int) bool {
		return mun.Events[i].Pos < mun.Events[j].Pos
	})

	return mun
}

func main() {
	// Отримання поточної директорії
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	// Шлях до MIDI-файлу
	f := filepath.Join(dir, "Star-Wars-Theme.mid")
	// Отримання відсортованих подій з MIDI-файлу
	mun := GetSortedEvents(f)

	// Створення об'єкта для рендерингу візуалізації на основі подій
	rndr := plrrndr.New(mun.Events)

	// Налаштування параметрів вікна та кадрів на секунду
	et.SetWindowSize(idgafwimd.WindowWidth, idgafwimd.WindowHeight)
	et.SetTPS(500)
	// Запуск гри з використанням рендерера
	if err := et.RunGame(rndr); err != nil {
		log.Fatal(err)
	}
}

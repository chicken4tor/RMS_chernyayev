package render

import (
	"errors"
	"fmt"
	et "github.com/hajimehoshi/ebiten/v2"
	_ "github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"image/color"
	"math"
	rnd "math/rand"
	idgafwimd "multi/utils"
	"sync"
	"time"
)

type Renderer struct {
	Events    []idgafwimd.MidiEvent
	CurrState *idgafwimd.State

	t0     *time.Time
	index  int
	wx, wy int
	yScale float64
	width  float64

	OptionsPool sync.Pool
	ImagesPool  sync.Pool

	renderMu                              sync.Mutex
	rRandCoef, gRandCoef, bRandCoef       float64
	rStrelochka, gStrelochka, bStrelochka bool
	softTime                              time.Duration
}

func New(events []idgafwimd.MidiEvent) *Renderer {
	randomer := rnd.NewSource(time.Now().Unix())
	r := &Renderer{
		Events: events,
		CurrState: &idgafwimd.State{
			ActiveNotes: make([]idgafwimd.MidiEvent, 0),
		},
		index: 0,
		OptionsPool: sync.Pool{
			New: func() any {
				return make([]*et.DrawImageOptions, 0, 128)
			},
		},
		ImagesPool: sync.Pool{
			New: func() any {
				return make([]*et.Image, 0, 128)
			},
		},
		renderMu:  sync.Mutex{},
		rRandCoef: (float64(randomer.Int63())/float64(math.MaxInt64))*1.6 + 0.4,
		gRandCoef: (float64(randomer.Int63())/float64(math.MaxInt64))*1.6 + 0.4,
		bRandCoef: (float64(randomer.Int63())/float64(math.MaxInt64))*1.6 + 0.4,
	}
	fmt.Println(r.rRandCoef)
	fmt.Println(r.gRandCoef)
	fmt.Println(r.bRandCoef)

	r.wx, r.wy = r.Layout(0, 0)
	r.yScale = math.Abs(float64(r.wy) / float64(idgafwimd.MaxY))
	r.width = float64(r.wx / idgafwimd.NotesCount)
	if !r.IsValid() {
		panic("🗿")
	}

	return r
}

func (r *Renderer) IsValid() bool {
	for _, v := range r.Events {
		if !v.IsValid() {
			return false
		}
	}
	return true
}

func (r *Renderer) NoteOn(event idgafwimd.MidiEvent) {
	r.CurrState.ActiveNotes = append(r.CurrState.ActiveNotes, event)
}

func (r *Renderer) NoteOff(event idgafwimd.MidiEvent, tn time.Time) {
	key := event.Key

	noteIndex, err := r.findNote(key)
	if err != nil {
		fmt.Println(err)
		return
	}

	note := r.CurrState.ActiveNotes[noteIndex]

	if tn.After(r.t0.Add(idgafwimd.Tick*time.Duration(note.Pos) + idgafwimd.SoftTime)) {
		r.CurrState.ActiveNotes = append(r.CurrState.ActiveNotes[:noteIndex], r.CurrState.ActiveNotes[noteIndex+1:]...)
	}
}

func (r *Renderer) findNote(key uint8) (int, error) {
	for i, note := range r.CurrState.ActiveNotes {
		if note.Key == key {
			return i, nil
		}
	}
	return 0, errors.New("🗿")
}

func (r *Renderer) clearNotes() {
	for i := range r.CurrState.Notes {
		r.CurrState.Notes[i].Velocity = 0
	}
}

func (r *Renderer) calculateNotes() {
	for _, v := range r.CurrState.ActiveNotes {
		r.CurrState.Notes[v.Key].Velocity = v.Velocity
	}
}

func (r *Renderer) indexToColor(index int) color.Color {
	R, G, B := uint8(0), uint8(0), uint8(0)
	R = uint8(255 * math.Abs(math.Sin(math.Sqrt(float64(index)*r.rRandCoef))))
	G = uint8(255 * math.Abs(math.Cos(math.Sqrt(float64(index)*r.gRandCoef))))
	B = uint8(255 * math.Abs(math.Sin(math.Sqrt(float64(index)*r.bRandCoef))))
	return color.RGBA{
		R: R, G: G, B: B, A: 255,
	}
}

func (r *Renderer) Update() error {
	if r.t0 == nil {
		t := time.Now()
		r.t0 = &t
	}

	tn := time.Now()
	for {
		if r.index >= len(r.Events) {
			return errors.New("no events")
		}
		eventTimeStamp := idgafwimd.Tick * time.Duration(r.Events[r.index].Pos)
		if r.t0.Add(eventTimeStamp).After(tn) {
			break
		}

		event := r.Events[r.index]

		switch event.Type {
		case idgafwimd.NoteOn:
			r.NoteOn(event)
		case idgafwimd.NoteOff:
			r.NoteOff(event, tn)
		default:
		}

		r.index++
	}
	r.clearNotes()
	r.calculateNotes()

	return nil
}

//func (r *Renderer) PosToDuration(pos uint64) time.Duration{
//	eventTimeStamp := idgafwimd.Tick * time.Duration(pos)
//	if r.t0.Add(eventTimeStamp).After(tn) {
//		break
//	}
//
//	return r.t0.Add(idgafwimd.Tick * )
//}

func (r *Renderer) updateCoefsAndFlags() {
	if r.rRandCoef > 1.7 {
		r.rStrelochka = true
	} else if r.rRandCoef < 0.4 {
		r.rStrelochka = false
	}
	if r.rStrelochka {
		r.rRandCoef *= 0.999996
	} else {
		r.rRandCoef *= 1.0000112
	}

	if r.gRandCoef > 1.7 {
		r.gStrelochka = true
	} else if r.gRandCoef < 0.4 {
		r.gStrelochka = false
	}
	if r.gStrelochka {
		r.gRandCoef *= 0.9999866
	} else {
		r.gRandCoef *= 1.0001002
	}

	if r.bRandCoef > 1.7 {
		r.bStrelochka = true
	} else if r.bRandCoef < 0.4 {
		r.bStrelochka = false
	}
	if r.bStrelochka {
		r.bRandCoef *= 0.99996
	} else {
		r.bRandCoef *= 1.0001102
	}
}

func (r *Renderer) Draw(screen *et.Image) {
	r.renderMu.Lock()
	screen.Fill(color.Black)

	images := r.ImagesPool.Get().([]*et.Image)
	options_ := r.OptionsPool.Get().([]*et.DrawImageOptions)
	slicesMu := sync.Mutex{}
	objMu := sync.RWMutex{}
	wg := sync.WaitGroup{}

	wg.Add(idgafwimd.Notes)

	//tn := time.Now()
	for i, v := range r.CurrState.Notes {
		go func(i int, v struct {
			Velocity uint8
			Pos      uint64
		}) {
			defer wg.Done()

			objMu.Lock()

			height := float64(v.Velocity) * r.yScale

			r.updateCoefsAndFlags()
			objMu.Unlock()

			if int(height) <= 0 || int(height) > 1500 {
				return
			}

			objMu.RLock()
			img := et.NewImage(int(r.width), int(height))
			objMu.RUnlock()
			img.Fill(r.indexToColor(i))
			options := &et.DrawImageOptions{}

			options.GeoM.Translate(float64(i)*r.width, float64(idgafwimd.MaxY-int(v.Velocity))*r.yScale)

			slicesMu.Lock()
			images = append(images, img)
			options_ = append(options_, options)
			slicesMu.Unlock()
		}(i, v)
	}

	wg.Wait()
	screen.Clear()
	for i := range images {
		if i == len(options_) {
			break
		}
		screen.DrawImage(images[i], options_[i])
	}
	r.OptionsPool.Put(options_)
	r.ImagesPool.Put(images)
	r.renderMu.Unlock()
}

func (r *Renderer) Layout(outsideWidth, outsideHeight int) (int, int) {
	return idgafwimd.WindowWidth, idgafwimd.WindowHeight
}

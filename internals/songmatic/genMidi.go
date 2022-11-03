package songmatic

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	"bytes"

	"github.com/aquilax/go-perlin"
	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/gm"
	"gitlab.com/gomidi/midi/v2/smf"
)

type Mode int

const (
	Ionian     Mode = 0
	Dorian     Mode = 1
	Phrygian   Mode = 2
	Lydian     Mode = 3
	Mixolydian Mode = 4
	Aeolian    Mode = 5
	Locrian    Mode = 6
)

var key [13]string

var degree [7]string

var interval [7]string
var seventh [7]string

// Sequence : MmmMMmdMmmMMmdMmmMMmd
// I Major  : MmmMMmd                Ionian       0
// II       :  mmMMmdM               Dorian       1
// III      :   mMMmdMm              Phrygian     2
// IV       :    MMmdMmm             Lydian       3
// V        :     MmdMmmM            Mixolydian   4
// VI Minor :      mdMmmMM           Aeolian      5
// VII      :       dMmmMMm          Locrian      6

var notes [7]string

var sharps [7]string

var midiMap map[string]uint8

type Scale struct {
	Notes       [7]string
	Accidentals uint8
	UseFlats    bool
}

// BarEvent A single event that happens within a bar on a beat. For example a chord
// or a bass note, or kick drum hit.
type BarEvent struct {
	Keys     []uint8
	Length   uint32
	Velocity uint8
}

// BarEvents Several bar events (several notes). This would hold one whole bar of
// notes, chords or drum hits
type BarEvents []BarEvent

// This is a group of bar events, that may or may not happen at the same time in
// a given measure for example:
// [0] kick
// [1] snare
// [2] high hat
type BarTracks []BarEvents

// This is a group of BarTracks (which equates to one measure) - so this could be
// considered a "song"... or several measures of different snippets of ideas
// These are played in order
type SongSnippet struct {
	Instr  gm.Instr
	Tracks []BarTracks
}

// resolution: 96 ticks per quarternote 960 is also common
var ticksPerQ = 480
var clock = smf.MetricTicks(ticksPerQ)

var lastNotePerlin = 2.0
var plin *perlin.Perlin

func Alloc() {
	//                   [ sharps                     ] [ flats                          ]
	//                0    1    2    3    4    5    6     7     8     9    10   11     12
	//                0    -    -    -    -    -    -     +     +     +	    +    +     +
	key = [13]string{"C", "G", "D", "A", "E", "B", "F#", "F", "Bb", "Eb", "Ab", "Db", "Gb"}
	//                  0    1    2    3    4    5    6
	sharps = [7]string{"F", "C", "G", "D", "A", "E", "B"}
	//                  0     1     2      3    4     5     6
	degree = [7]string{"I", "II", "III", "IV", "V", "VI", "VII°"}
	//                    0    1    2    3    4    5    6
	interval = [7]string{"M", "m", "m", "M", "M", "m", "°"}
	seventh = [7]string{"∆7", "-7", "-7", "∆7", "7", "-7", "°7"}
	//                 0   1   2   3   4   5   6
	notes = [7]string{"A", "B", "C", "D", "E", "F", "G"}

	midiMap = map[string]uint8{
		"B#": 60,
		"C":  60,
		"C#": 61,
		"Db": 61,
		"D":  62,
		"D#": 63,
		"Eb": 63,
		"E":  64,
		"Fb": 64,
		"E#": 65,
		"F":  65,
		"F#": 66,
		"Gb": 66,
		"G":  67,
		"G#": 68,
		"Ab": 68,
		"A":  69,
		"A#": 70,
		"Bb": 70,
		"B":  71,
		"Cb": 71,
	}

	lastNotePerlin = RandomPerlinPos()
	rand.Seed(time.Now().UnixNano())

	alpha := 2.0
	beta := 2.0
	n := int32(3)
	const MaxInt = int(^uint(0) >> 1)
	seed := rand.Intn(MaxInt)
	plin = perlin.NewPerlin(alpha, beta, n, int64(seed))
}

// Anything after B will need an octave boost to go
// up instead of staying in the lower mid range
func Oct(note uint8, octave int8) uint8 {
	res := note + uint8(12*octave)
	if res > 127 {
		res -= 12
	}
	return res
}

// Give a tonic and a number of accidentals this returns
// the notes of a scale (or the Triad Chords depending on how you)
// look at it.
func Chords(tonic int, numAccidentals int, flats bool) [7]string {

	var inharmonic [7]string

	///////
	// Get our lookup tables for accidentals
	var accidentals []string
	if flats {
		accidentals = sharps[7-numAccidentals : 7]
	} else {
		accidentals = sharps[0:numAccidentals]
	}
	///////

	///////
	strAccidental := "#"
	if flats {
		strAccidental = "b"
	}
	///////

	///////
	i := 0
	for i < 7 {

		if Contains(accidentals, notes[tonic]) {
			inharmonic[i] = fmt.Sprintf("%s%s", notes[tonic], strAccidental)
		} else {
			inharmonic[i] = fmt.Sprintf("%v", notes[tonic])
		}

		i++

		tonic++
		if tonic == 7 {
			tonic = 0
		}
	}
	///////

	return inharmonic
}

// Looks for the needle in the haystack.
// Warning: currently linear
func Contains(haystack []string, needle string) bool {
	for _, i := range haystack {
		if needle == i {
			return true
		}
	}
	return false
}

// Generates cord flavors based on the Mode (major, minor, etc)
func ScaleDegrees(pos Mode) ([7]string, [7]string, [7]string) {
	var degrees [7]string
	var triTones [7]string
	var seventhTones [7]string

	i := 0
	for i < 7 {
		var scaleDegree string
		if interval[pos] != "M" {
			scaleDegree = strings.ToLower(degree[pos])
		} else {
			scaleDegree = degree[pos]
		}

		degrees[i] = fmt.Sprintf("%v", scaleDegree)
		triTones[i] = fmt.Sprintf("%v", interval[pos])
		seventhTones[i] = fmt.Sprintf("%v", seventh[pos])

		i++
		pos++
		if pos == 7 {
			pos = 0
		}
	}

	return degrees, triTones, seventhTones
}

const (
	BiasNone        = 0
	BiasOne         = 1
	BiasOneAndThree = 257
	BiasTwoAndFour  = 4112
	Bias4th         = 4369
	Bias8th         = 21845
)

// Beats are marked backwards
//                <---
// 0000 0000 0000 0000
//
// So for example:
// 0000 0001 0000 0001 =  257 = 1 & 3
// 0001 0000 0001 0000 = 4112 = 2 & 4
func GenerateRhythm(bias uint16) uint16 {
	max := 65535
	min := 0
	beat := uint16(rand.Intn(max-min) + min)
	return uint16(beat | bias)
}

func GenerateTempo() uint8 {
	max := 150
	min := 60
	v := rand.Intn(max-min) + min
	return uint8(v)
}

func RandomNote(scale [7]string) (string, int8) {
	notePrefs := []int8{
		0, 1, 2, 3, 4, 5, 6,
		// 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, // root
		// 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, // 3rd
		// 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, // 3rd
		// 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, // 5th
		// 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, // 5th
		// 1, 1, 1, 1, 1, 5, 5, 5, 5, 5, // 2nd & 6th
		// 3, 3, 3, 3, 3, 3, 3, 6, 6, 6, // 4th & 7th
	}

	loc := int8(math.Abs(plin.Noise1D(lastNotePerlin)*100)) % int8(len(notePrefs))
	lastNotePerlin += .01

	// this is here if we want to try to control the percentage of
	// particular notes
	notePos := notePrefs[loc]

	// * | * | * |
	// 1 2 3 4 5 6 7
	// 0 1 2 3 4 5 6
	return scale[notePos], notePos
}

func RandomPerlinPos() float64 {
	return rand.Float64()
}

func RandomFromSlice(list []int8) int8 {
	max := len(list)
	min := 0
	v := rand.Intn(max-min) + min
	return list[v]
}

func RandomOctave() int8 {
	octaves := []int8{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 2}
	return RandomFromSlice(octaves)
}

func RandomChordExtension() int8 {
	extensions := []int8{2, 4, 6, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7}
	return RandomFromSlice(extensions)
}

// RandMidiRange pick a number that can be used within a midi message
func RandMidiRange(min int, max int) uint8 {
	v := rand.Intn(max-min) + min
	return uint8(v)
}

func GenerateScale() Scale {
	max := 12
	min := 0

	v := rand.Intn(max-min) + min

	wantKey := v

	if wantKey > 12 {
		panic("Not enough notes for that")
	}

	songKey := key[wantKey]
	sharpOrFlat := 0
	numSharps := 0
	mod := 1
	if len(songKey) > 1 {
		sf := songKey[1]
		// 98 is b, it's flat
		if sf == 98 {
			mod = 6
			wantKey++
		}
		sharpOrFlat = 7 % mod
		numSharps = wantKey % 7
	} else if songKey == "F" {
		mod = 6
		sharpOrFlat = 1
		numSharps = 1
	} else {
		sharpOrFlat = 7 % mod
		numSharps = wantKey % 7
	}

	useFlats := false
	if sharpOrFlat == 1 {
		useFlats = true
	}

	noteIndex := int((songKey[0] % 64) - 1)
	scale := Chords(noteIndex, numSharps, useFlats)

	return Scale{scale, uint8(numSharps), useFlats}
}

func DisplayModes() {

	scale := GenerateScale()

	d, _, s := ScaleDegrees(Ionian)
	fmt.Printf("\nIonian\n")
	for i := 0; i < 7; i++ {
		fmt.Printf("%v%v (%v) | ", scale.Notes[i], s[i], d[i])
	}
	fmt.Printf("\n")

	d, _, s = ScaleDegrees(Dorian)
	fmt.Printf("\nDorian\n")
	for i := 0; i < 7; i++ {
		fmt.Printf("%v%v (%v) | ", scale.Notes[i], s[i], d[i])
	}
	fmt.Printf("\n")

	d, _, s = ScaleDegrees(Phrygian)
	fmt.Printf("\nPhrygian\n")
	for i := 0; i < 7; i++ {
		fmt.Printf("%v%v (%v) | ", scale.Notes[i], s[i], d[i])
	}
	fmt.Printf("\n")

	d, _, s = ScaleDegrees(Lydian)
	fmt.Printf("\nLydian\n")
	for i := 0; i < 7; i++ {
		fmt.Printf("%v%v (%v) | ", scale.Notes[i], s[i], d[i])
	}
	fmt.Printf("\n")

	d, _, s = ScaleDegrees(Mixolydian)
	fmt.Printf("\nMixolydian\n")
	for i := 0; i < 7; i++ {
		fmt.Printf("%v%v (%v) | ", scale.Notes[i], s[i], d[i])
	}
	fmt.Printf("\n")

	d, _, s = ScaleDegrees(Aeolian)
	fmt.Printf("\nAeolian\n")
	for i := 0; i < 7; i++ {
		fmt.Printf("%v%v (%v) | ", scale.Notes[i], s[i], d[i])
	}
	fmt.Printf("\n")

	d, _, s = ScaleDegrees(Locrian)
	fmt.Printf("\nLocrian\n")
	for i := 0; i < 7; i++ {
		fmt.Printf("%v%v (%v) | ", scale.Notes[i], s[i], d[i])
	}
	fmt.Printf("\n")
}

// Generate one bar of music with 16th note fidelity
func randomBarEvents(scale Scale, jazz bool) BarEvents {
	// var barEvents []BarEvent
	var notes = make([]BarEvent, 16)
	// 1 e + a 2 e + a 3 e + a 4 e + a
	// 0 1 2 3 4 5 6 7 8 9 A B C D E F
	t := GenerateRhythm(Bias4th)
	for i := 0; i < 16; i++ {
		on := (t >> i) & 1
		if on == 1 {
			rootNote, degree := RandomNote(scale.Notes)
			notes[i] = BarEvent{[]uint8{
				Oct(midiMap[rootNote], RandomOctave()),
				Oct(midiMap[scale.Notes[(degree+3)%7]], RandomOctave()),
				Oct(midiMap[scale.Notes[(degree+5)%7]], RandomOctave()),
			}, clock.Ticks16th(), RandMidiRange(50, 110)}

			if jazz {
				notes[i].Keys = append(
					notes[i].Keys,
					Oct(midiMap[scale.Notes[(degree+RandomChordExtension())%7]], RandomOctave()),
				)
			}

		} else {
			notes[i] = BarEvent{[]uint8{0}, clock.Ticks16th(), 0}
		}
	}
	return notes
}

func RandomChords(tempo float64, scale Scale, bars int, jazz bool) []byte {
	beatPerBar := 4
	channel := 0

	var snippet SongSnippet
	snippet.Tracks = make([]BarTracks, bars)

	for m := 0; m < bars; m++ {
		var track BarTracks
		track = append(track, randomBarEvents(scale, jazz))
		snippet.Tracks[m] = track
	}

	snippet.Instr = gm.Instr_ElectricGuitarJazz
	return mkSMF(uint8(channel), tempo, uint8(beatPerBar), scale, snippet)
}

func RandomBass(tempo float64, scale Scale, bars int) []byte {
	beatPerBar := 4
	channel := 0

	var snippet SongSnippet
	snippet.Tracks = make([]BarTracks, bars)

	for m := 0; m < bars; m++ {
		var track BarTracks
		var tune = make([]BarEvent, 16)
		// 1 e + a 2 e + a 3 e + a 4 e + a
		// 0 1 2 3 4 5 6 7 8 9 A B C D E F
		t := GenerateRhythm(BiasOne)
		for i := 0; i < 16; i++ {
			on := (t >> i) & 1
			if on == 1 {
				note, _ := RandomNote(scale.Notes)
				tune[i] = BarEvent{[]uint8{
					Oct(midiMap[note], -2+-RandomOctave()),
				}, clock.Ticks16th(), RandMidiRange(80, 110)}
			} else {
				tune[i] = BarEvent{[]uint8{0}, clock.Ticks16th(), 0}
			}
		}
		track = append(track, tune)

		snippet.Tracks[m] = track
	}

	snippet.Instr = gm.Instr_ElectricBassFinger
	return mkSMF(uint8(channel), tempo, uint8(beatPerBar), scale, snippet)
}

func RandomMelody(tempo float64, scale Scale, bars int) []byte {
	beatPerBar := 4
	channel := 0

	var snippet SongSnippet
	snippet.Tracks = make([]BarTracks, bars)

	for m := 0; m < bars; m++ {
		var track BarTracks
		var tune = make([]BarEvent, 16)
		// 1 e + a 2 e + a 3 e + a 4 e + a
		// 0 1 2 3 4 5 6 7 8 9 A B C D E F
		t := GenerateRhythm(Bias4th)
		for i := 0; i < 16; i++ {
			on := (t >> i) & 1
			if on == 1 {
				note, _ := RandomNote(scale.Notes)
				tune[i] = BarEvent{[]uint8{
					Oct(midiMap[note], RandomOctave()),
				}, clock.Ticks16th(), RandMidiRange(80, 110)}
			} else {
				tune[i] = BarEvent{[]uint8{0}, clock.Ticks16th(), 0}
			}
		}
		track = append(track, tune)

		snippet.Tracks[m] = track
	}

	snippet.Instr = gm.Instr_DistortionGuitar
	return mkSMF(uint8(channel), tempo, uint8(beatPerBar), scale, snippet)
}

func RandomBeat(tempo float64, scale Scale, bars int) []byte {
	beatPerBar := 4
	channel := 9 // midi defined drum track

	var snippet SongSnippet
	snippet.Tracks = make([]BarTracks, bars)

	for m := 0; m < bars; m++ {
		////////////////////////////////////////////////
		var tracks BarTracks

		var kick = make([]BarEvent, 16)
		// 1 e + a 2 e + a 3 e + a 4 e + a
		// 0 1 2 3 4 5 6 7 8 9 A B C D E F
		k := GenerateRhythm(BiasOneAndThree)
		for i := 0; i < 16; i++ {
			on := (k >> i) & 1
			if on == 1 {
				kick[i] = BarEvent{[]uint8{gm.DrumKey_AcousticBassDrum.Key()}, clock.Ticks16th(), RandMidiRange(70, 110)}
			}
		}

		var snare = make([]BarEvent, 16)
		s := GenerateRhythm(BiasTwoAndFour)
		for i := 0; i < 16; i++ {
			on := (s >> i) & 1
			if on == 1 {
				snare[i] = BarEvent{[]uint8{gm.DrumKey_AcousticSnare.Key()}, clock.Ticks16th(), RandMidiRange(0, 100)}
			}
		}

		var highhat = make([]BarEvent, 16)
		h := GenerateRhythm(Bias4th)
		for i := 0; i < 16; i++ {
			on := (h >> i) & 1
			if on == 1 {
				highhat[i] = BarEvent{[]uint8{gm.DrumKey_ClosedHiHat.Key()}, clock.Ticks16th(), RandMidiRange(0, 100)}
			}
		}

		tracks = append(tracks, kick)
		tracks = append(tracks, snare)
		tracks = append(tracks, highhat)
		////////////////////////////////////////////////
		snippet.Tracks[m] = tracks
	}

	snippet.Instr = gm.Instr_SynthDrum
	return mkSMF(uint8(channel), tempo, uint8(beatPerBar), scale, snippet)
}

// makes a SMF and returns the bytes
func mkSMF(channel uint8, tempo float64, beatPerBar uint8, scale Scale, snippet SongSnippet) []byte {
	var (
		bf bytes.Buffer
		tr smf.Track
		ch = channel
	)

	// first track must have tempo and meter information
	tr.Add(0, smf.MetaMeter(beatPerBar, 4))
	tr.Add(0, smf.MetaTempo(tempo))
	tr.Add(0, smf.MetaTimeSig(beatPerBar, 4, 0, 0))
	tr.Add(0, smf.MetaKey(0, true, scale.Accidentals, scale.UseFlats))
	tr.Add(0, smf.MetaInstrument(snippet.Instr.String()))
	tr.Add(0, midi.ProgramChange(0, snippet.Instr.Value()))

	// because delta time
	for b := 0; b < len(snippet.Tracks); b++ {
		beat := 0
		barEvents := snippet.Tracks[b]
		for i := 0; i < (ticksPerQ * int(beatPerBar)); i++ {

			// when a 16th note happens...
			if i%int(clock.Ticks16th()) == 0 {
				// loop over the array of events that could be happening in this *bar*
				for be := 0; be < len(barEvents); be++ {
					// grab one event, this may have several events in it - like a track
					events := barEvents[be]
					// a "track" could have lots of keys (notes) aka a chord
					for e := 0; e < len(events); e++ {
						if beat == e && events[e].Keys != nil {
							// Should only send the note off with a length message once
							// and all others with note off 0 offset
							// All the NoteOns on this beat...
							for k := 0; k < len(events[e].Keys); k++ {
								if events[e].Keys[k] != 0 {
									tr.Add(0, midi.NoteOn(ch, events[e].Keys[k], events[e].Velocity))
								}
							}

							// Then all the NoteOffs for this beat...
							off := false
							for k := 0; k < len(events[e].Keys); k++ {
								// if this is not the drum channel, respect note length
								// else we just note off on zero like we do for other notes
								// in this chord
								if channel != 9 && !off {
									tr.Add(events[e].Length, midi.NoteOff(ch, 0))
									off = true
								}
								tr.Add(0, midi.NoteOff(ch, events[e].Keys[k]))
							}
						}
					}
				}

				// This ensures everything sticks to a 16th note grid. It make extra events,
				// but that doesn't seem to hurt anything
				if channel == 9 {
					tr.Add(clock.Ticks16th(), midi.NoteOff(ch, 0))
				}
				beat++
			}
		}
	}
	tr.Close(0)

	// create the SMF and add the tracks
	s := smf.New()
	s.TimeFormat = clock
	s.Add(tr)
	s.WriteTo(&bf)
	return bf.Bytes()
}

// Generate midi files. The number of bars will be the length of the
// idea (number of measures of music), and then num ideas will be the
// number of midi files created. Jazz if you want to use extensions in
// in the chord generating
// func GenerateMidi(numBars int, numIdeas int, jazz bool) {
// 	tempo := float64(GenerateTempo())
// 	bars := numBars
// 	scale := GenerateScale()

// 	log.Printf("Making ideas in scale: %v\n", scale)
// 	log.Printf("With tempo: %v\n", tempo)

// 	for ideas := 0; ideas < numIdeas; ideas++ {
// 		beatSlice := RandomBeat(tempo, scale, bars)
// 		bassSlice := RandomBass(tempo, scale, bars)
// 		melodySlice := RandomMelody(tempo, scale, bars)
// 		chordSlice := RandomChords(tempo, scale, bars, jazz)

// 		beatFile, err := os.OpenFile(fmt.Sprintf("beat_%v.mid", ideas), os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 		defer beatFile.Close()

// 		bassFile, err := os.OpenFile(fmt.Sprintf("bass_%v.mid", ideas), os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 		defer bassFile.Close()

// 		melodyFile, err := os.OpenFile(fmt.Sprintf("melody_%v.mid", ideas), os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 		defer melodyFile.Close()

// 		chordFile, err := os.OpenFile(fmt.Sprintf("chords_%v.mid", ideas), os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 		defer chordFile.Close()

// 		_, err = beatFile.Write(beatSlice)
// 		if err != nil {
// 			log.Fatal(err)
// 		}

// 		_, err = bassFile.Write(bassSlice)
// 		if err != nil {
// 			log.Fatal(err)
// 		}

// 		_, err = melodyFile.Write(melodySlice)
// 		if err != nil {
// 			log.Fatal(err)
// 		}

// 		_, err = chordFile.Write(chordSlice)
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 	}
// }

// func main() {
// 	alloc()
// 	fmt.Printf("Perlin: %v\n", lastNotePerlin)

// 	// GenerateRhythm()
// 	// DisplayModes()
// 	GenerateMidi(15, 1, true)
// }

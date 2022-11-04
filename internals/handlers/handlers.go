package handlers

import (
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strconv"

	"github.com/robrohan/legendary-doodle/internals/models"
	"github.com/robrohan/legendary-doodle/internals/songmatic"
)

type pageData struct {
	Title       string
	CompanyName string
}

func TemplateInit() *template.Template {
	t, err := template.ParseGlob("./templates/*")
	if err != nil {
		log.Println("Cannot parse templates: ", err)
		os.Exit(-1)
	}

	return t
}

var routeMatch, _ = regexp.Compile(`\/(\w+)`)

func ServePage(env *models.Env, t *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pd := pageData{
			"Songmatic Home",
			"Songmatic Template",
		}

		matches := routeMatch.FindStringSubmatch(r.URL.Path)

		env.Log.Printf("Request: %v", r.URL.Path)
		env.Log.Printf("Request: %v", matches)

		if len(matches) >= 1 {
			page := matches[1] + ".html"
			if t.Lookup(page) != nil {
				w.WriteHeader(200)
				t.ExecuteTemplate(w, page, pd)
				return
			}
		} else if r.URL.Path == "/" {
			w.WriteHeader(200)
			t.ExecuteTemplate(w, "index.html", pd)
			return
		}

		w.WriteHeader(404)
		w.Write([]byte("Not Found"))
	}
}

func ServeMidiDownload(env *models.Env, t *template.Template) http.HandlerFunc {
	songmatic.Alloc()
	return func(w http.ResponseWriter, r *http.Request) {

		frmKey := r.URL.Query().Get("key")
		frmTempo := r.URL.Query().Get("tempo")
		frmType := r.URL.Query().Get("type")
		frmBars := r.URL.Query().Get("bars")

		key, err := strconv.Atoi(frmKey)
		if err != nil {
			log.Printf("Bunk key given in form: %v", frmKey)
			max := 12
			min := 0
			v := rand.Intn(max-min) + min
			key = v
		}

		tempoI, err := strconv.Atoi(frmTempo)
		if err != nil {
			log.Printf("Bunk key given in form: %v", frmTempo)
			tempoI = int(songmatic.GenerateTempo())
		}

		genType, err := strconv.Atoi(frmType)
		if err != nil {
			log.Printf("Bunk type given in form: %v", frmType)
			tempoI = int(songmatic.GenerateTempo())
		}

		bars, err := strconv.Atoi(frmBars)
		if err != nil {
			log.Printf("Bunk bars given in form: %v", frmBars)
			tempoI = int(songmatic.GenerateTempo())
		}

		tempo := float64(tempoI)
		scale := songmatic.GenerateScale(key)
		jazz := false

		var genSlice []byte
		fname := "chords"
		switch genType {
		case 0:
			genSlice = songmatic.RandomChords(tempo, scale, bars, jazz)
			fname = "chords"
		case 1:
			genSlice = songmatic.RandomBeat(tempo, scale, bars)
			fname = "drums"
		case 2:
			genSlice = songmatic.RandomBass(tempo, scale, bars)
			fname = "bass"
		case 3:
			genSlice = songmatic.RandomMelody(tempo, scale, bars)
			fname = "melody"
		}

		fileName := fmt.Sprintf("%s_%v_%s.midi", fname, tempo, scale.Notes[0])
		w.Header().Set("Content-Type", "audio/midi")
		w.Header().Set("Content-Disposition", "inline; filename="+fileName)
		w.Header().Set("Content-Length", fmt.Sprintf("%v", len(genSlice)))
		w.Write(genSlice)
	}
}

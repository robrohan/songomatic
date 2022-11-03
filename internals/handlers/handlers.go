package handlers

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"

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
			"Go Web Template Home",
			"Go Web Template",
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
		tempo := float64(songmatic.GenerateTempo())
		bars := 4 // numBars
		scale := songmatic.GenerateScale()
		jazz := false

		fileName := fmt.Sprintf("chords_%s.midi", scale.Notes[0])

		// beatSlice := songmatic.RandomBeat(tempo, scale, bars)
		// bassSlice := songmatic.RandomBass(tempo, scale, bars)
		// melodySlice := songmatic.RandomMelody(tempo, scale, bars)
		chordSlice := songmatic.RandomChords(tempo, scale, bars, jazz)

		w.Header().Set("Content-Type", "audio/midi")
		w.Header().Set("Content-Disposition", "inline; filename="+fileName)
		w.Header().Set("Content-Length", fmt.Sprintf("%v", len(chordSlice)))
		w.Write(chordSlice)
	}
}

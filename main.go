package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/jesusrmoreno/img-standin/Godeps/_workspace/src/github.com/codegangsta/negroni"
	"github.com/jesusrmoreno/img-standin/Godeps/_workspace/src/github.com/fogleman/gg"
	"github.com/jesusrmoreno/img-standin/Godeps/_workspace/src/github.com/gorilla/mux"
	"github.com/jesusrmoreno/img-standin/Godeps/_workspace/src/github.com/lucasb-eyer/go-colorful"
)

const maxWidth = 10000
const maxHeight = 10000

// Sets the color of the text on the image
const (
	defaultTextColor   = "#ffffff"
	defaultTextOutline = "#000000"
)

// Color Format Wrong, Please use Hex w/o hash symbol eg: ff0000
func imageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hex, ok := vars["color"]
	if !ok {
		http.Error(w, "Missing Color", 400)
		return
	}
	hex = "#" + hex
	width, ok := vars["width"]
	if !ok {
		http.Error(w, "Missing Width", 400)
		return
	}
	height, ok := vars["height"]
	if !ok {
		http.Error(w, "Missing Height", 400)
		return
	}

	c, err := colorful.Hex(hex)
	if err != nil {
		http.Error(w, "Invalid Color, use hex eg: #ffffff", 400)
		return
	}

	wi, err := strconv.Atoi(width)
	if err != nil {
		http.Error(w, "Width must be a number", 400)
		return
	}
	he, err := strconv.Atoi(height)
	if err != nil {
		http.Error(w, "Height must be a number", 400)
		return
	}
	if wi > maxWidth || he > maxHeight {
		msg := fmt.Sprintf("Image too big must of size less than %dx%d",
			maxWidth, maxHeight)
		http.Error(w, msg, 400)
		return
	}
	text, ok := vars["text"]
	if !ok {
		text = width + "x" + height
	} else {
		text = strings.Replace(text, "_", " ", -1)
	}

	_, withCross := vars["x"]
	w.Header().Set("Content-Type", "image/jpeg")
	if err := createImage(text, wi, he, c, withCross).EncodePNG(w); err != nil {
		http.Error(w, "Unable to create image", 500)
		return
	}
}

func createImage(text string, w, h int, bg colorful.Color, c bool) *gg.Context {
	m := gg.NewContext(w, h)
	m.SetColor(bg)
	m.DrawRectangle(0, 0, float64(w), float64(h))
	m.Fill()
	if c {
		m.DrawLine(0, 0, float64(w), float64(h))
		m.DrawLine(0, float64(h), float64(w), 0)
		if bg.R+bg.G+bg.B > 1.5 {
			m.SetRGBA(0, 0, 0, .7)
		} else {
			m.SetRGBA(1, 1, 1, .7)
		}
		m.Stroke()
	}
	m.SetHexColor(defaultTextOutline)
	strokeSize := 2
	for dy := -strokeSize; dy <= strokeSize; dy++ {
		for dx := -strokeSize; dx <= strokeSize; dx++ {
			x := float64(w/2 + dx)
			y := float64(h/2 + dy)
			m.DrawStringAnchored(text, x, y, 0.5, 0.5)
		}
	}
	m.SetHexColor(defaultTextColor)
	m.DrawStringAnchored(text, float64(w)/2, float64(h/2), 0.5, 0.5)
	m.Fill()
	return m
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/{text}/{color}-{width}-{height}-{x}.png", imageHandler).
		Methods("GET")
	r.HandleFunc("/{text}/{color}-{width}-{height}.png", imageHandler).
		Methods("GET")
	r.HandleFunc("/{color}-{width}-{height}-{x}.png", imageHandler).
		Methods("GET")
	r.HandleFunc("/{color}-{width}-{height}.png", imageHandler).
		Methods("GET")
	n := negroni.Classic()
	n.UseHandler(r)
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	n.Run(":" + port)
}

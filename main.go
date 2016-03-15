package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/codegangsta/negroni"
	"github.com/fogleman/gg"
	"github.com/gorilla/mux"
	"github.com/lucasb-eyer/go-colorful"
)

const maxWidth = 10000
const maxHeight = 10000

type err interface {
	error
	Status() int
}

// StatusError represents an error with an associated HTTP status code.
type statusError struct {
	Code int
	Err  error
}

// Allows StatusError to satisfy the error interface.
func (se statusError) Error() string {
	return se.Err.Error()
}

// Returns our HTTP status code.
func (se statusError) Status() int {
	return se.Code
}

// Color Format Wrong, Please use Hex w/o hash symbol eg: ff0000
func imageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hex, ok := vars["color"]
	if !ok {
		sErr := statusError{
			Code: 400,
			Err:  errors.New("Missing Color"),
		}
		http.Error(w, sErr.Error(), sErr.Status())
		return
	}
	hex = "#" + hex
	width, ok := vars["width"]
	if !ok {
		sErr := statusError{
			Code: 400,
			Err:  errors.New("Missing Width"),
		}
		http.Error(w, sErr.Error(), sErr.Status())
		return
	}
	height, ok := vars["height"]
	if !ok {
		sErr := statusError{
			Code: 400,
			Err:  errors.New("Missing Height"),
		}
		http.Error(w, sErr.Error(), sErr.Status())
		return
	}

	c, err := colorful.Hex(hex)
	if err != nil {
		sErr := statusError{
			Code: 400,
			Err:  errors.New("Invalid Color Format. Use HEX eg: ff0000"),
		}
		http.Error(w, sErr.Error(), sErr.Status())
		return
	}

	wi, err := strconv.Atoi(width)
	if err != nil {
		sErr := statusError{
			Code: 400,
			Err:  errors.New("Invalid Width"),
		}
		http.Error(w, sErr.Error(), sErr.Status())
		return
	}
	he, err := strconv.Atoi(height)
	if err != nil {
		sErr := statusError{
			Code: 400,
			Err:  errors.New("Invalid Height"),
		}
		http.Error(w, sErr.Error(), sErr.Status())
		return
	}
	if wi > maxWidth || he > maxHeight {
		msg := fmt.Sprintf("Image too big must of size less than %dx%d",
			maxWidth, maxHeight)
		sErr := statusError{
			Code: 400,
			Err:  errors.New(msg),
		}
		http.Error(w, sErr.Error(), sErr.Status())
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
		http.Error(w, err.Error(), 500)
		return
	}
}

const fontSize = 24

// if the r+g+b is > (256 / 2) * 3) -> lineColor = black else lineColor = white

func createImage(text string, w, h int, bg colorful.Color, cross bool) *gg.Context {
	m := gg.NewContext(w, h)
	m.SetColor(bg)
	m.DrawRectangle(0, 0, float64(w), float64(h))
	m.Fill()
	red := bg.R
	green := bg.G
	blue := bg.B
	if cross {
		m.DrawLine(0, 0, float64(w), float64(h))
		m.DrawLine(0, float64(h), float64(w), 0)
		if red+green+blue > (.5 * 3) {
			m.SetRGBA(0, 0, 0, .7)
		} else {
			m.SetRGBA(1, 1, 1, .7)
		}
		m.Stroke()
	}
	m.SetRGB(red, green, blue)
	strokeSize := 5
	for dy := -strokeSize; dy <= strokeSize; dy++ {
		for dx := -strokeSize; dx <= strokeSize; dx++ {
			x := float64(w/2 + dx)
			y := float64(h/2 + dy)
			m.DrawStringAnchored(text, x, y, 0.5, 0.5)
		}
	}
	if red+green+blue > (.5 * 3) {
		m.SetRGBA(0, 0, 0, .7)
	} else {
		m.SetRGBA(1, 1, 1, .7)
	}
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

package main

import (
	"errors"
	"net/http"
	"os"
	"strconv"

	"github.com/codegangsta/negroni"
	"github.com/fogleman/gg"
	"github.com/gorilla/mux"
	"github.com/lucasb-eyer/go-colorful"
)

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
	if wi > 10000 || he > 10000 {
		sErr := statusError{
			Code: 400,
			Err:  errors.New("Width or Height too big. Must be no more than 50000px"),
		}
		http.Error(w, sErr.Error(), sErr.Status())
		return
	}

	sizeText := width + "x" + height
	w.Header().Set("Content-Type", "image/jpeg")
	if err := createImage(sizeText, wi, he, c).EncodePNG(w); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

}

const fontSize = 24

func createImage(text string, width, height int, background colorful.Color) *gg.Context {
	m := gg.NewContext(width, height)
	m.SetColor(background)
	m.DrawRectangle(0, 0, float64(width), float64(height))
	m.Fill()
	m.SetHexColor("#000")
	strokeSize := 3
	for dy := -strokeSize; dy <= strokeSize; dy++ {
		for dx := -strokeSize; dx <= strokeSize; dx++ {
			// give it rounded corners
			if dx*dx+dy*dy >= strokeSize*strokeSize {
				continue
			}
			x := float64(width/2 + dx)
			y := float64(height - fontSize + dy)
			m.DrawStringAnchored(text, x, y, 0.5, 0.5)
		}
	}
	m.SetHexColor("#ffffff")
	m.DrawStringAnchored(text, float64(width)/2, float64(height)-fontSize, 0.5, 0.5)
	m.Fill()
	return m
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/{color}-{width}-{height}.png", imageHandler).Methods("GET")
	n := negroni.Classic()
	n.UseHandler(r)
	n.Run(":" + os.Getenv("PORT"))
}

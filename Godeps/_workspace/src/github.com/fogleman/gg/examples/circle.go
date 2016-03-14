package main

import "github.com/jesusrmoreno/img-standin/Godeps/_workspace/src/github.com/fogleman/gg"

func main() {
	dc := gg.NewContext(1000, 1000)
	dc.DrawCircle(500, 500, 400)
	dc.SetRGB(0, 0, 0)
	dc.Fill()
	dc.SavePNG("out.png")
}

package main

import (
	"fmt"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"golang.org/x/image/colornames"
)

func displayTime(win *pixelgl.Window, atlas *text.Atlas) {
	currentTime := time.Now().Format("03:04:05 PM")
	txt := text.New(pixel.V(320, 240), atlas)

	if time.Now().Hour() < 12 {
		txt.Color = colornames.Yellow
	} else {
		txt.Color = colornames.White
	}

	fmt.Fprintln(txt, currentTime)
	txt.Draw(win, pixel.IM.Scaled(txt.Orig, 4))
}

func handleTimeInput(win *pixelgl.Window) {
	// Add any specific input handling for the time display screen here
	// Currently, this function does nothing but is defined to avoid undefined errors
}

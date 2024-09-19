package main

import (
	"fmt"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"golang.org/x/image/colornames"
)

const (
	WorkDuration  = 20 * time.Minute
	BreakDuration = 5 * time.Minute
)

var (
	timerState   = "work"
	timerEnd     time.Time
	timerRunning = false
	pausedTime   = WorkDuration
)

func displayPomodoro(win *pixelgl.Window, atlas *text.Atlas) {
	now := time.Now()
	if timerRunning && now.After(timerEnd) {
		if timerState == "work" {
			timerState = "break"
			timerEnd = now.Add(BreakDuration)
			timerRunning = true
		} else {
			timerState = "work"
			timerRunning = false
			pausedTime = BreakDuration
		}
	}

	var remaining time.Duration
	if timerRunning {
		remaining = timerEnd.Sub(now)
	} else {
		remaining = pausedTime
	}

	minutes := int(remaining.Minutes())
	seconds := int(remaining.Seconds()) % 60

	txt := text.New(pixel.V(320, 240), atlas)
	if timerState == "work" {
		txt.Color = colornames.White
	} else {
		txt.Color = colornames.Blue
	}

	if timerRunning {
		fmt.Fprintf(txt, "%02d:%02d", minutes, seconds)
	} else {
		// Show "Paused" above the timer
		pausedTxt := text.New(pixel.V(320, 280), atlas)
		pausedTxt.Color = colornames.Red
		fmt.Fprintln(pausedTxt, "Paused")
		pausedTxt.Draw(win, pixel.IM.Scaled(pausedTxt.Orig, 2))

		// Blink the time when paused
		if int(now.Unix())%2 == 0 {
			fmt.Fprintf(txt, "%02d:%02d", minutes, seconds)
		}
	}
	txt.Draw(win, pixel.IM.Scaled(txt.Orig, 4))

	if !timerRunning && minutes == 0 && seconds == 0 {
		// Blink 00:00
		if int(now.Unix())%2 == 0 {
			txt.Clear()
		}
	}
}

func handlePomodoroInput(win *pixelgl.Window) {
	if win.JustPressed(pixelgl.MouseButton3) {
		toggleTimer()
	}
	if win.JustPressed(pixelgl.MouseButton4) {
		timerState = "work"
		timerEnd = time.Now().Add(WorkDuration)
		timerRunning = false
		pausedTime = WorkDuration
	}
}

func toggleTimer() {
	if timerRunning {
		pausedTime = timerEnd.Sub(time.Now())
		timerRunning = false
	} else {
		timerEnd = time.Now().Add(pausedTime)
		timerRunning = true
	}
}

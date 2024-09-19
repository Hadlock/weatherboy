package main

import (
	"log"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	bolt "go.etcd.io/bbolt"
	"golang.org/x/image/colornames"
)

var (
	currentMode = 1
	db          *bolt.DB
)

func main() {
	var err error
	db, err = bolt.Open("weather.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("WeatherData"))
		return err
	})
	if err != nil {
		log.Fatal(err)
	}

	// Load weather data on application start
	loadWeatherData()

	pixelgl.Run(run)
}

func run() {
	cfg := pixelgl.WindowConfig{
		Title:  "Desktop Clock",
		Bounds: pixel.R(0, 0, 640, 480),
		VSync:  true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	// Load fonts
	basicAtlas := LoadBasicFont()
	//	ttfAtlas, err := LoadTTF("ZillaSlab-SemiBold.ttf", 24)

	for !win.Closed() {
		if win.JustPressed(pixelgl.MouseButton1) {
			currentMode = (currentMode % 3) + 1
		}
		if win.JustPressed(pixelgl.MouseButton2) || win.JustPressed(pixelgl.MouseButton6) || win.JustPressed(pixelgl.KeyQ) || win.JustPressed(pixelgl.KeyEscape) {
			win.SetClosed(true)
		}

		win.Clear(colornames.Black)

		switch currentMode {
		case 1:
			handleTimeInput(win)
			displayTime(win, basicAtlas)
		case 2:
			handleWeatherInput(win)
			displayWeather(win, basicAtlas)
		case 3:
			handlePomodoroInput(win)
			displayPomodoro(win, basicAtlas)
		}
		win.Update()
	}
}

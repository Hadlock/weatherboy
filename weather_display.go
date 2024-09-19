package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"github.com/go-resty/resty/v2"
	bolt "go.etcd.io/bbolt"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"
)

const (
	PointsEndpoint = "https://api.weather.gov/points/40.7483,-73.9856"
	UpdateInterval = 35 * time.Minute
)

func loadWeatherData() {
	client := resty.New()
	resp, err := client.R().
		SetHeader("User-Agent", "YourAppName (your-email@example.com)").
		Get(PointsEndpoint)
	if err != nil {
		fmt.Println("Error fetching grid points:", err)
		return
	}

	var pointsData struct {
		Properties struct {
			Forecast       string `json:"forecast"`
			ForecastHourly string `json:"forecastHourly"`
		} `json:"properties"`
	}
	err = json.Unmarshal(resp.Body(), &pointsData)
	if err != nil {
		fmt.Println("Error unmarshalling grid points data:", err)
		return
	}

	forecastEndpoint := pointsData.Properties.Forecast
	resp, err = client.R().
		SetHeader("User-Agent", "YourAppName (your-email@example.com)").
		Get(forecastEndpoint)
	if err != nil {
		fmt.Println("Error fetching weather data:", err)
		return
	}

	var weatherData struct {
		Properties struct {
			Periods []struct {
				Name            string `json:"name"`
				Temperature     int    `json:"temperature"`
				TemperatureUnit string `json:"temperatureUnit"`
				ShortForecast   string `json:"shortForecast"`
			} `json:"periods"`
		} `json:"properties"`
	}
	err = json.Unmarshal(resp.Body(), &weatherData)
	if err != nil {
		fmt.Println("Error unmarshalling weather data:", err)
		return
	}

	forecastHourlyEndpoint := pointsData.Properties.ForecastHourly
	resp, err = client.R().
		SetHeader("User-Agent", "YourAppName (your-email@example.com)").
		Get(forecastHourlyEndpoint)
	if err != nil {
		fmt.Println("Error fetching hourly weather data:", err)
		return
	}

	var hourlyData struct {
		Properties struct {
			Periods []struct {
				StartTime        string `json:"startTime"`
				Temperature      int    `json:"temperature"`
				TemperatureUnit  string `json:"temperatureUnit"`
				RelativeHumidity struct {
					Value int `json:"value"`
				} `json:"relativeHumidity"`
			} `json:"periods"`
		} `json:"properties"`
	}
	err = json.Unmarshal(resp.Body(), &hourlyData)
	if err != nil {
		fmt.Println("Error unmarshalling hourly weather data:", err)
		return
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("WeatherData"))
		data, err := json.Marshal(weatherData)
		if err != nil {
			return err
		}
		err = b.Put([]byte("data"), data)
		if err != nil {
			return err
		}
		hourlyDataBytes, err := json.Marshal(hourlyData)
		if err != nil {
			return err
		}
		err = b.Put([]byte("hourlyData"), hourlyDataBytes)
		if err != nil {
			return err
		}
		return b.Put([]byte("lastUpdate"), itob(time.Now().UnixNano()))
	})
	if err != nil {
		fmt.Println("Error writing to database:", err)
	}
}

func displayWeather(win *pixelgl.Window, atlas *text.Atlas) {
	var weatherData struct {
		Properties struct {
			Periods []struct {
				Name            string `json:"name"`
				Temperature     int    `json:"temperature"`
				TemperatureUnit string `json:"temperatureUnit"`
				ShortForecast   string `json:"shortForecast"`
			} `json:"periods"`
		} `json:"properties"`
	}
	var hourlyData struct {
		Properties struct {
			Periods []struct {
				StartTime        string `json:"startTime"`
				Temperature      int    `json:"temperature"`
				TemperatureUnit  string `json:"temperatureUnit"`
				RelativeHumidity struct {
					Value int `json:"value"`
				} `json:"relativeHumidity"`
			} `json:"periods"`
		} `json:"properties"`
	}
	var lastUpdate time.Time

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("WeatherData"))
		data := b.Get([]byte("data"))
		if data != nil {
			if err := json.Unmarshal(data, &weatherData); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("no weather data available")
		}
		hourlyDataBytes := b.Get([]byte("hourlyData"))
		if hourlyDataBytes != nil {
			if err := json.Unmarshal(hourlyDataBytes, &hourlyData); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("no hourly weather data available")
		}
		lastUpdateBytes := b.Get([]byte("lastUpdate"))
		if lastUpdateBytes != nil {
			lastUpdate = time.Unix(0, int64(binary.BigEndian.Uint64(lastUpdateBytes)))
		}
		return nil
	})
	if err != nil {
		fmt.Println("Error reading from database:", err)
		displayError(win, atlas, "Failed to load weather data")
		return
	}

	// Check if the Periods slice is not empty
	if len(weatherData.Properties.Periods) == 0 || len(hourlyData.Properties.Periods) == 0 {
		displayError(win, atlas, "No weather data available")
		return
	}

	// Display the current temperature and humidity
	txt := text.New(pixel.V(10, 450), atlas)
	txt.Color = colornames.White
	currentTemp := hourlyData.Properties.Periods[0].Temperature
	tempUnit := hourlyData.Properties.Periods[0].TemperatureUnit
	currentHumidity := hourlyData.Properties.Periods[0].RelativeHumidity.Value
	fmt.Fprintf(txt, "Current Temperature: %d %s / %d%%\n", currentTemp, tempUnit, currentHumidity)
	txt.Draw(win, pixel.IM.Scaled(txt.Orig, 2))

	// Display the three-day forecast
	yPos := 400.0
	for i := 1; i <= 3 && i < len(weatherData.Properties.Periods); i++ {
		period := weatherData.Properties.Periods[i]

		// Draw the forecast text
		txt := text.New(pixel.V(10, yPos), atlas)
		txt.Color = colornames.White
		fmt.Fprintf(txt, "%s: %d %s - %s\n", period.Name, period.Temperature, period.TemperatureUnit, period.ShortForecast)
		txt.Draw(win, pixel.IM.Scaled(txt.Orig, 2))

		// Adjust yPos to create space between text and art
		yPos -= 40

		// Draw the ASCII art
		drawWeatherArt(win, period.ShortForecast, pixel.V(10, yPos))

		// Adjust yPos to create space for the next forecast
		yPos -= 100
	}

	// Draw the last fetched timestamp at the bottom right in tiny font
	smallAtlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)
	smallTxt := text.New(pixel.V(win.Bounds().Max.X-290, 10), smallAtlas)
	smallTxt.Color = colornames.White
	fmt.Fprintf(smallTxt, "Weather last fetched: %s", lastUpdate.Format("01/02/06 15:04:05"))
	smallTxt.Draw(win, pixel.IM.Scaled(smallTxt.Orig, 1))
}

func drawWeatherArt(win *pixelgl.Window, forecast string, pos pixel.Vec) {
	smallAtlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)
	smallTxt := text.New(pos, smallAtlas)
	smallTxt.Color = colornames.White

	var art string
	switch {
	case contains(forecast, "Sunny"):
		art = `
   \   /  
    .-.   
 - (   ) -
    ` + "`" + `-’   
   /   \  
`
	case contains(forecast, "Cloudy"):
		art = `
     .--.    
  .-(    ).  
 (___.__)__) 
`
	case contains(forecast, "Rain"):
		art = `
     .-.     
    (   ).   
   (___(__)  
  ‘ ‘ ‘ ‘ ‘  
 ‘ ‘ ‘ ‘ ‘  
`
	default:
		art = `
     ???     
    (   )    
   (___(__)  
`
	}
	fmt.Fprintln(smallTxt, art)
	smallTxt.Draw(win, pixel.IM.Scaled(smallTxt.Orig, 1))
}

func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

func handleWeatherInput(win *pixelgl.Window) {
	// Add any specific input handling for the weather display screen here
}

func displayError(win *pixelgl.Window, atlas *text.Atlas, message string) {
	win.Clear(colornames.Black)
	txt := text.New(pixel.V(320, 240), atlas)
	txt.Color = colornames.Red
	fmt.Fprintln(txt, message)
	txt.Draw(win, pixel.IM.Scaled(txt.Orig, 2))
	win.Update()
}

func itob(v int64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}

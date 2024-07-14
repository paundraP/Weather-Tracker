package main

import (
	"encoding/json"
	"html/template"
	"net/http"
	"os"
)

type apiconfigData struct {
	OpenWeatherMapApiKey string `json:"OpenWeatherMapApiKey"`
}

type weatherData struct {
	Name string `json:"name"`
	Main struct {
		Temperature float64 `json:"temp"`
	} `json:"main"`
}

func loadApiConfig(filename string) (apiconfigData, error) {
	bytes, err := os.ReadFile(filename)
	if err != nil {
		return apiconfigData{}, err
	}
	var c apiconfigData
	err = json.Unmarshal(bytes, &c)
	if err != nil {
		return apiconfigData{}, err
	}
	return c, nil
}

func query(city string) (weatherData, error) {
	apiConfig, err := loadApiConfig("./apiConfig.json")
	if err != nil {
		return weatherData{}, err
	}
	resp, err := http.Get("http://api.openweathermap.org/data/2.5/weather?units=metric&APPID=" + apiConfig.OpenWeatherMapApiKey + "&q=" + city)
	if err != nil {
		return weatherData{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return weatherData{}, nil
	}
	var d weatherData
	if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
		return weatherData{}, err
	}
	return d, nil
}

func weatherHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		city := r.URL.Query().Get("city")
		var data weatherData
		var err error
		errorMessage := ""

		if city != "" {
			data, err = query(city)
			if err != nil {
				http.Error(w, "Error fetching weather data: "+err.Error(), http.StatusInternalServerError)
				return
			}
			if data.Name == "" {
				errorMessage = "Nama kota yang anda masukkan tidak ditemukan."
			}
		}

		tmpl, err := template.ParseFiles("index.html")
		if err != nil {
			http.Error(w, "Error loading template: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl.Execute(w, struct {
			WeatherData  weatherData
			ErrorMessage string
		}{
			WeatherData:  data,
			ErrorMessage: errorMessage,
		})
	}
}

func main() {
	http.HandleFunc("/weather", weatherHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.ListenAndServe("localhost:8000", nil)
}

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
)

type ViaCEP struct {
	Localidade string `json:"localidade"`
	Uf         string `json:"uf"`
}

type OpenMeteoResponse struct {
	CurrentWeather struct {
		Temperature float64 `json:"temperature"`
	} `json:"current_weather"`
}

type TempResp struct {
	TempC float64 `json:"temp_C"`
	TempF float64 `json:"temp_F"`
	TempK float64 `json:"temp_K"`
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/tempo", WeatherHandler)
	fmt.Println("Servidor escutando na porta 8080...")
	http.ListenAndServe(":8080", mux)
}

func WeatherHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/tempo" {
		http.NotFound(w, r)
		return
	}

	cepParam := r.URL.Query().Get("cep")
	if !validateCEP(cepParam) {
		http.Error(w, "invalid zipcode", http.StatusUnprocessableEntity)
		return
	}

	cepData, err := BuscaCEP(cepParam)
	if err != nil || cepData.Localidade == "" {
		http.Error(w, "cannot find zipcode", http.StatusNotFound)
		return
	}

	tempC, err := fetchWeather(cepData.Localidade)
	if err != nil {
		http.Error(w, "erro ao buscar temperatura", http.StatusInternalServerError)
		return
	}

	temps := convertTemps(tempC)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(temps)
}

func validateCEP(cep string) bool {
	re := regexp.MustCompile(`^\d{8}$`)
	return re.MatchString(cep)
}

func BuscaCEP(cep string) (*ViaCEP, error) {
	resp, err := http.Get("https://viacep.com.br/ws/" + cep + "/json/")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var c ViaCEP
	err = json.Unmarshal(body, &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func fetchWeather(cidade string) (float64, error) {

	geoAPIURL := fmt.Sprintf("https://geocoding-api.open-meteo.com/v1/search?name=%s&count=1&language=pt&format=json", url.QueryEscape(cidade))
	respGeo, err := http.Get(geoAPIURL)
	if err != nil {
		return 0, fmt.Errorf("erro ao buscar coordenadas para %s: %w", cidade, err)
	}
	defer respGeo.Body.Close()

	if respGeo.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(respGeo.Body)
		return 0, fmt.Errorf("erro da API de geocodificação (%s): %s", geoAPIURL, string(body))
	}

	bodyGeo, err := io.ReadAll(respGeo.Body)
	if err != nil {
		return 0, fmt.Errorf("erro ao ler resposta de geocodificação: %w", err)
	}

	var geoResult struct {
		Results []struct {
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
		} `json:"results"`
	}

	if err := json.Unmarshal(bodyGeo, &geoResult); err != nil {
		return 0, fmt.Errorf("erro ao decodificar JSON de geocodificação: %w", err)
	}

	if len(geoResult.Results) == 0 {
		return 0, fmt.Errorf("não foi possível encontrar coordenadas para a cidade: %s", cidade)
	}

	latitude := geoResult.Results[0].Latitude
	longitude := geoResult.Results[0].Longitude

	weatherURL := fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%.6f&longitude=%.6f&current_weather=true", latitude, longitude)

	resp, err := http.Get(weatherURL)
	if err != nil {
		return 0, fmt.Errorf("erro ao buscar dados do Open-Meteo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("erro da API Open-Meteo (%s): %s", weatherURL, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("erro ao ler resposta do Open-Meteo: %w", err)
	}

	var weatherData OpenMeteoResponse
	err = json.Unmarshal(body, &weatherData)
	if err != nil {
		return 0, fmt.Errorf("erro ao decodificar JSON do Open-Meteo: %w. Resposta: %s", err, string(body))
	}

	return weatherData.CurrentWeather.Temperature, nil
}

func convertTemps(celsius float64) TempResp {
	return TempResp{
		TempC: celsius,
		TempF: celsius*1.8 + 32,
		TempK: celsius + 273,
	}
}

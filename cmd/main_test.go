package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestValidateCEP(t *testing.T) {
	valido := "12345678"
	invalido := "1234"

	if !validateCEP(valido) {
		t.Errorf("validateCEP falhou para CEP válido: %s", valido)
	}

	if validateCEP(invalido) {
		t.Errorf("validateCEP falhou para CEP inválido: %s", invalido)
	}
}

func TestWeatherHandler_InvalidCEP(t *testing.T) {
	req, err := http.NewRequest("GET", "/tempo?cep=123", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(WeatherHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusUnprocessableEntity {
		t.Errorf("esperado status %v, obteve %v", http.StatusUnprocessableEntity, status)
	}
}

func TestWeatherHandler_MissingCEP(t *testing.T) {
	req, err := http.NewRequest("GET", "/tempo", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(WeatherHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusUnprocessableEntity {
		t.Errorf("esperado status %v, obteve %v", http.StatusUnprocessableEntity, status)
	}
}

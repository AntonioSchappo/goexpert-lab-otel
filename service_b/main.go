package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

type (
	Location struct {
		CEP         string `json:"cep,omitempty"`
		Logradouro  string `json:"logradouro,omitempty"`
		Complemento string `json:"complemento,omitempty"`
		Bairro      string `json:"bairro,omitempty"`
		Location    string `json:"localidade,omitempty"`
		UF          string `json:"uf,omitempty"`
		Error       bool   `json:"erro,omitempty"`
	}

	WeatherResponse struct {
		Current Current `json:"current"`
	}

	Current struct {
		Temp_C float64 `json:"temp_c"`
		Temp_F float64 `json:"temp_f"`
	}

	ResponseDto struct {
		Temp_C float64 `json:"temp_C"`
		Temp_K float64 `json:"temp_K"`
		Temp_F float64 `json:"temp_F"`
		City   string  `json:"city"`
	}
)

const weatherApiKey = "58550850a19a4f7fb6a132328242104"

func main() {
	setTracing()
	http.HandleFunc("GET /{cep}", Handle)
	fmt.Println("Server running on port 8081")
	http.ListenAndServe(":8081", nil)
}

func Handle(w http.ResponseWriter, r *http.Request) {
	cep := r.PathValue("cep")
	ctx, span := otel.Tracer("service-b").Start(r.Context(), "2 - service-b-start")
	defer span.End()
	if valid := validCep(cep); !valid {
		http.Error(w, "invalid zipcode", http.StatusUnprocessableEntity)
		return
	}

	l, err := getLocation(ctx, cep)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if l.Error {
		http.Error(w, "can not find zipcode", http.StatusNotFound)
		return
	}

	weather, err := getWeather(ctx, l.Location)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := ResponseDto{
		Temp_C: weather.Temp_C,
		Temp_K: weather.Temp_C + 273.15,
		Temp_F: weather.Temp_F,
		City:   l.Location,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func setTracing() {
	exporter, err := zipkin.New("http://zipkin:9411/api/v2/spans")
	if err != nil {
		log.Fatalf("Fail to create Zipkin exporter: %v", err)
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("service-b"),
		)),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})
}

func getLocation(ctx context.Context, cep string) (Location, error) {
	_, span := otel.Tracer("service-b").Start(ctx, "3 - service-b-get-location")
	defer span.End()

	url := fmt.Sprintf("https://viacep.com.br/ws/%s/json", cep)
	resp, err := http.Get(url)
	if err != nil {
		return Location{}, err
	}

	defer resp.Body.Close()

	var location Location
	if err := json.NewDecoder(resp.Body).Decode(&location); err != nil {
		return Location{}, err
	}

	return location, nil
}

func getWeather(ctx context.Context, location string) (*Current, error) {
	_, span := otel.Tracer("service-b").Start(ctx, "4 - service-b-get-weather")
	defer span.End()
	url := fmt.Sprintf("http://api.weatherapi.com/v1/current.json?key=%s&q=%s&aqi=no", weatherApiKey, url.QueryEscape(location))

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var weather WeatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&weather); err != nil {
		return nil, err
	}

	return &weather.Current, nil
}

func validCep(cep string) bool {
	if len(cep) != 8 {
		return false
	}
	for _, c := range cep {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

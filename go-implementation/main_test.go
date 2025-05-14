package main

import (
	"bytes"
	"encoding/json"
	"sirbubbls.io/challenge/broadcast"
	"sirbubbls.io/challenge/handlers"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type RequestData struct {
	Date        string  `json:"date"`
	Humidity    float32 `json:"humidity"`
	Temperature float32 `json:"temperature"`
}
type RequestResponse struct {
	Date        time.Time `json:"date"`
	Humidity    float32   `json:"humidity"`
	Temperature float32   `json:"temperature"`
}

func (e *RequestResponse) Equals(val RequestData) bool {
	if val.Humidity != e.Humidity || val.Temperature != e.Temperature {
		return false
	}
	parsedTime, err := time.Parse("2006-01-02", val.Date)
	if err != nil || e.Date.Day() != parsedTime.Day() || e.Date.Month() != parsedTime.Month() || e.Date.Year() != parsedTime.Year() {
		return false
	}
	return true
}

func setupRouter() http.Handler {
	db, err := sqlx.Connect("postgres", "host=localhost user=postgres password=postgres dbname=postgres sslmode=disable")
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}

	query := `DELETE FROM  measurements`
	db.Exec(query)

	app := &handlers.AppContext{DB: db, Updates: broadcast.NewBroadcastHub()}
	mux := http.NewServeMux()
	mux.HandleFunc("/health", handlers.Health)
	mux.HandleFunc("/submit", app.ReceiveWheaterData)
	mux.HandleFunc("/day", app.GetWeatherForDay)
	mux.HandleFunc("/range", app.GetWeatherForRange)
	mux.HandleFunc("/updates", app.WsHandler)
	return mux
}

func createNewDataRequest(data RequestData) *http.Request {
	jsonBytes, _ := json.Marshal(data)
	req := httptest.NewRequest("POST", "/submit", bytes.NewReader(jsonBytes))
	req.Header.Set("Content-Type", "application/json")
	return req

}

// this test does a basic test to the /health api
func TestHealth(t *testing.T) {
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	// Use your router/handler
	handler := setupRouter()
	handler.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if status := resp.StatusCode; status != http.StatusOK {
		t.Errorf("expected status 200, got %d", status)
	}
}

// submits a new measurement to the /submit api and validates that the returned data equals the data sent in the request
// - submits data via /submit
// - queries the data from the same day via /day api
// - validates if submitted and queried objects are equal
func TestNewMeasurement(t *testing.T) {
	payload := RequestData{
		Date:        "2021-01-01",
		Humidity:    2.1,
		Temperature: 3.3,
	}

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		t.Errorf("Unable to serialize data")
	}
	req := httptest.NewRequest("POST", "/submit", bytes.NewReader(jsonBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Use your router/handler
	handler := setupRouter()
	handler.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if status := resp.StatusCode; status != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("expected status 200, got %s %s", resp.Status, string(body))
	}

	// check if data is correctly stored
	req = httptest.NewRequest("GET", "/day?day=2021-01-01", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	resp = w.Result()
	defer resp.Body.Close()

	if status := resp.StatusCode; status != http.StatusOK {
		t.Errorf("expected status 200, got %s", resp.Status)
	}
	decoder := json.NewDecoder(resp.Body)
	decoder.DisallowUnknownFields()

	var data RequestResponse
	if err := decoder.Decode(&data); err != nil {
		t.Errorf("unable to deserialize response")
		return
	}

	if !data.Equals(payload) {
		t.Errorf("Unexpected response value %+v %+v", data, payload)
		return
	}
}

// submits a new measurement to the /submit api and validates that the returned data from the /range endpoint
// - submits data via /submit
// - queries the data from the same day via /range api
// - validates if the correct date range containing the correct data is returned
func TestRange(t *testing.T) {
	payloads := []RequestData{
		{Date: "2021-01-01", Humidity: 2.1, Temperature: 3.3},
		{Date: "2021-01-02", Humidity: 3.1, Temperature: 4.3},
		{Date: "2021-01-03", Humidity: 4.1, Temperature: 5.3},
		{Date: "2021-01-04", Humidity: 5.1, Temperature: 6.3},
		{Date: "2021-01-05", Humidity: 6.1, Temperature: 7.3},
	}
	handler := setupRouter()
	for _, payload := range payloads {
		req := createNewDataRequest(payload)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}

	w := httptest.NewRecorder()

	// check if data is correctly stored
	req := httptest.NewRequest("GET", "/range?start=2021-01-02&end=2021-01-04", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	resp := w.Result()
	defer resp.Body.Close()

	if status := resp.StatusCode; status != http.StatusOK {
		t.Errorf("expected status 200, got %s", resp.Status)
		return
	}
	decoder := json.NewDecoder(resp.Body)
	decoder.DisallowUnknownFields()

	var data []RequestResponse
	if err := decoder.Decode(&data); err != nil {
		t.Errorf("unable to deserialize response")
		return
	}

	if len(data) != 3 {
		t.Errorf("expected 3 days in the response payload got %d", len(data))
		return
	}
	for i, response := range data {
		if !response.Equals(payloads[i+1]) {
			t.Errorf("Expected %+v got %+v", payloads[i+1], response)
			return
		}
	}
}

// tests the update stream exposed by the /updates websocket api
// - creates two websocket clients, each having their own connection
// - submits data via /submit
// - reads a single payload from each websocket connections
// - checks if both connections got a response and if the response is correct
func TestUpdateStream(t *testing.T) {
	db, err := sqlx.Connect("postgres", "host=localhost user=postgres password=postgres dbname=postgres sslmode=disable")
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}

	query := `DELETE FROM  measurements`
	db.Exec(query)

	app := &handlers.AppContext{DB: db, Updates: broadcast.NewBroadcastHub()}

	server := httptest.NewServer(http.HandlerFunc(app.WsHandler))
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn1, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}
	defer conn1.Close()

	conn2, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}
	defer conn2.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/submit", app.ReceiveWheaterData)

	payload := RequestData{
		Date:        "2021-01-01",
		Humidity:    2.1,
		Temperature: 3.3,
	}
	req := createNewDataRequest(payload)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	_, resp1, err := conn1.ReadMessage()
	_, resp2, err := conn2.ReadMessage()

	var response1 RequestResponse
	err = json.Unmarshal(resp1, &response1)

	if err != nil {
		t.Fatalf("Unable to deserialize WS response: %v", err)
		return
	}

	var response2 RequestResponse
	err = json.Unmarshal(resp2, &response2)

	if err != nil {
		t.Fatalf("Unable to deserialize WS response: %v", err)
		return
	}

	if response1.Date != response2.Date || response1.Temperature != response2.Temperature || response1.Humidity != response2.Humidity {

		t.Fatalf("Both ws connections got two different values %v %v", response1, response2)
		return
	}

	if !response1.Equals(payload) {
		t.Fatalf("Returned value is incorrect. Expected: %v, got: %v", payload, response1)
		return
	}
}

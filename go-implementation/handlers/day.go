package handlers

import (
	"encoding/json"
	"sirbubbls.io/challenge/models"
	_ "github.com/lib/pq" // PostgreSQL driver
	"time"
	"net/http"
)

func (app *AppContext) GetWeatherForDay(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var data models.Weather

	// extract query parameter
	queryParams := r.URL.Query()
	day := queryParams.Get("day")
	parsedTime, err := time.Parse("2006-01-02", day)
	if err != nil {
		http.Error(w, "Invalid date format"+err.Error(), http.StatusBadRequest)
		return
	}


	// submit temperature to database
	query := `SELECT * FROM measurements WHERE time=$1 LIMIT 1`
	err = app.DB.Get(&data, query, parsedTime)
	if err != nil {
		http.Error(w, "Failed to insert temperature measurement: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

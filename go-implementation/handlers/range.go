package handlers

import (
	"encoding/json"
	"sirbubbls.io/challenge/models"
	_ "github.com/lib/pq" // PostgreSQL driver
	"net/http"
	"time"
)

func (app *AppContext) GetWeatherForRange(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var data []models.Weather

	// extract query parameter
	queryParams := r.URL.Query()
	startDay := queryParams.Get("start")
	startTime, errStart := time.Parse("2006-01-02", startDay)
	endDay := queryParams.Get("end")
	endTime, errEnd := time.Parse("2006-01-02", endDay)
	if errStart != nil || errEnd != nil {
		http.Error(w, "Invalid date format", http.StatusBadRequest)
		return
	}

	// submit temperature to database
	query := `SELECT * FROM measurements WHERE time >= $1 AND time <= $2 ORDER BY time ASC`
	dbErr := app.DB.Select(&data, query, startTime, endTime)
	if dbErr != nil {
		http.Error(w, "Failed to insert temperature measurement: "+dbErr.Error(), http.StatusInternalServerError)
		return
	}


	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

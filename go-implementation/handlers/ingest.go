package handlers

import (
	"encoding/json"
	"sirbubbls.io/challenge/models"
	_ "github.com/lib/pq" // PostgreSQL driver
	"net/http"
)

// request handler for handling data ingest requests
// incoming JSON data will be stored in the database
func (app *AppContext) ReceiveWheaterData(w http.ResponseWriter, r *http.Request) {
	var data models.Weather

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := decoder.Decode(&data); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if !data.IsValid() {
		http.Error(w, "Invalid Values for Payload", http.StatusBadRequest)
		return
	}


	// submit temperature to database
	query := `INSERT INTO measurements (temperature, humidity, time) VALUES (:temperature, :humidity, :time) ON CONFLICT (time) DO UPDATE SET humidity=excluded.humidity, temperature=excluded.temperature`
	_, err := app.DB.NamedExec(query, &data)
	if err != nil {
		http.Error(w, "Failed to insert temperature measurement: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// submit temperature to the update channel
	app.Updates.Broadcast(data)

	w.WriteHeader(200)
}

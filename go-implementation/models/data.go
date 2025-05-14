package models

import (
    "encoding/json"
    "fmt"
    "time"
)

type Weather struct {
	Date time.Time`json:"date" db:"time"`
	Temperature float32 `json:"temperature" db:"temperature"`
	Humidity float32 `json:"humidity" db:"humidity"`
}

// das ist ja wild
const customLayout = "2006-01-02"

// Custom unmarshal function for `time.Time` in Event struct
func (e *Weather) UnmarshalJSON(data []byte) error {
    type Alias Weather

    // Unmarshal the basic fields
    var aux struct {
        Time string `json:"date"`
        Alias
    }

    if err := json.Unmarshal(data, &aux); err != nil {
        return err
    }

    // Parse the time field from custom format
    parsedTime, err := time.Parse(customLayout, aux.Time)
    if err != nil {
        return fmt.Errorf("failed to parse time: %w", err)
    }

    // Assign the parsed time to the struct's Time field
    e.Date = parsedTime
    e.Temperature = aux.Temperature // Transfer the other fields
    e.Humidity = aux.Humidity // Transfer the other fields

    return nil
}

func (e *Weather) IsValid() bool {
    if(e.Humidity < 0 || e.Humidity > 100) {
        return false
    }
    return true
}

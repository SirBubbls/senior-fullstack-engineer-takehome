package main

import (
	"sirbubbls.io/challenge/database"
	"sirbubbls.io/challenge/handlers"
	"sirbubbls.io/challenge/broadcast"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"log"
	"net/http"
)

func main() {
	//
	// initialize database connection
	//
	db, err := sqlx.Connect("postgres", "host=localhost user=postgres password=postgres dbname=postgres sslmode=disable")
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	defer db.Close()

	//
	// apply database migrations
	//
	if err := database.RunMigrations(db); err != nil {
		if err == migrate.ErrNoChange {
			log.Println("No new migrations to apply.")
		} else {
			log.Fatalf("Migration failed: %v", err)
		}
	} else {
		log.Println("Migrations applied successfully.")
	}

	//
	// construct application context
	//
	updates := broadcast.NewBroadcastHub() 
	app := &handlers.AppContext{DB: db, Updates: updates}

	//
	// run web-server and serve endpoints
	//
	http.HandleFunc("/health", handlers.Health)
	http.HandleFunc("/submit", app.ReceiveWheaterData)
	http.HandleFunc("/day", app.GetWeatherForDay)
	http.HandleFunc("/range", app.GetWeatherForRange)
	http.HandleFunc("/updates", app.WsHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

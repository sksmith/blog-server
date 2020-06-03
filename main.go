package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
)

// Comment represents a single comment on a blog post
type Comment struct {
	Day     string
	Month   string
	Year    string
	Author  string
	Comment string
	Created time.Time
}

const port = "8082"

// following gorilla/mux's recommendations on best practices for
// app server startup
func main() {
	var wait time.Duration
	flag.DurationVar(&wait, "graceful-timeout", time.Second*15, "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	flag.Parse()

	log.Println("configuring mux...")
	r := mux.NewRouter()
	r.HandleFunc("/posts/{year}/{month}/{day}/comments", get).Methods("GET")
	r.HandleFunc("/posts/{year}/{month}/{day}/comments", create).Methods("POST")
	http.Handle("/", r)

	srv := &http.Server{
		Addr:         "0.0.0.0:" + port,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r,
	}

	log.Printf("listening for requests at %s", port)

	go func() {
		if err := srv.ListenAndServeTLS(os.Getenv("BLOG_SERVER_CERT"), os.Getenv("BLOG_SERVER_KEY")); err != nil {
			log.Println(err)
		}
	}()

	c := make(chan os.Signal, 1)

	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	srv.Shutdown(ctx)
	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	log.Println("shutting down")
	os.Exit(0)
}

func getPathDate(vars map[string]string) (year, month, day string) {
	year = vars["year"]
	if len(year) == 2 {
		year = "20" + year
	}

	month = vars["month"]
	if len(month) == 1 {
		month = "0" + month
	}

	day = vars["day"]
	if len(day) == 1 {
		day = "0" + day
	}

	return
}

func get(w http.ResponseWriter, r *http.Request) {
	year, month, day := getPathDate(mux.Vars(r))

	comment := &Comment{Year: year, Month: month, Day: day}

	js, err := json.Marshal(comment)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(js)
}

func create(w http.ResponseWriter, r *http.Request) {
	year, month, day := getPathDate(mux.Vars(r))

	comment := &Comment{Year: year, Month: month, Day: day}

	js, err := json.Marshal(comment)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(js)
}

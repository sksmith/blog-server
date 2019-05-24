package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	yaml "gopkg.in/yaml.v2"

	"github.com/gorilla/mux"
	"github.com/russross/blackfriday"
)

// Post is a single blog post
type Post struct {
	Title          string
	Subtitle       string
	Author         string
	Path           string
	Created        time.Time
	Edited         time.Time
	Tags           []string
	ContentPreview string
	Content        string
}

// Tags contains all tags associated with blog posts
var Tags []string

// Posts contains all blog posts
var Posts []*Post

// PostMap contains all posts in 2006-01-02 format
var PostMap map[string]*Post

const port = "8081"

// following gorilla/mux's recommendations on best practices for
// app server startup
func main() {
	var wait time.Duration
	flag.DurationVar(&wait, "graceful-timeout", time.Second*15, "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	flag.Parse()

	if os.Getenv("BLOGPATH") == "" {
		log.Fatal("environment variable BLOGPATH must be set")
	}

	log.Println("loading files...")
	loadFiles()

	log.Println("configuring mux...")
	r := mux.NewRouter()
	r.HandleFunc("/tags", getTags)
	r.HandleFunc("/posts", getPosts)
	r.HandleFunc("/posts/{year}/{month}/{day}", getPost)
	r.HandleFunc("/refresh", refresh)
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
		if err := srv.ListenAndServe(); err != nil {
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

func getPost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	year := vars["year"]
	if len(year) == 2 {
		year = "20" + year
	}

	month := vars["month"]
	if len(month) == 1 {
		month = "0" + month
	}

	day := vars["day"]
	if len(day) == 1 {
		day = "0" + day
	}

	key := fmt.Sprintf("%s-%s-%s", year, month, day)
	log.Printf("requested: %s", key)

	js, err := json.Marshal(PostMap[key])

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(js)
}

func getPosts(w http.ResponseWriter, r *http.Request) {
	js, err := json.Marshal(Posts)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(js)
}

func getTags(w http.ResponseWriter, r *http.Request) {
	js, err := json.Marshal(Tags)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(js)
}

func refresh(w http.ResponseWriter, r *http.Request) {
	loadFiles()
}

func loadFiles() error {
	files, err := ioutil.ReadDir(os.Getenv("BLOGPATH"))
	if err != nil {
		return err
	}

	Posts = make([]*Post, 0)
	PostMap = make(map[string]*Post)
	TagMap := make(map[string]bool)

	// Must close the encoder when finished to flush any partial blocks.
	// If you comment out the following line, the last partial block "r"
	// won't be encoded.

	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".md") {
			continue
		}

		data, err := ioutil.ReadFile(os.Getenv("BLOGPATH") + "/" + f.Name())
		if err != nil {
			log.Printf("unable to read file: %s; error: %v\n", f.Name(), err)
			continue
		}

		metaData, bodyPreview, body, err := splitMetadata(data)
		if err != nil || len(metaData) == 0 {
			log.Printf("unable to split metadata for file: %s; error: %v\n", f.Name(), err)
			continue
		}

		post := &Post{}
		err = yaml.Unmarshal(metaData, post)
		if err != nil {
			log.Printf("unable to parse metadata for file: %s; error: %v\n", f.Name(), err)
			continue
		}

		for _, tag := range post.Tags {
			TagMap[tag] = true
		}

		if len(bodyPreview) > 0 {
			bodyPreview = blackfriday.MarkdownCommon(bodyPreview)
			post.ContentPreview = base64.StdEncoding.EncodeToString(bodyPreview)
		}

		body = blackfriday.MarkdownCommon(body)
		post.Path = post.Created.Format("/posts/2006/01/02")
		post.Content = base64.StdEncoding.EncodeToString(body)

		Posts = append(Posts, post)
		PostMap[post.Created.Format("2006-01-02")] = post
	}

	Tags = make([]string, len(TagMap))
	i := 0
	for tag := range TagMap {
		Tags[i] = tag
		i++
	}

	Posts = sortposts(Posts)

	return nil
}

func sortposts(posts []*Post) []*Post {
	if len(posts) < 2 {
		return posts
	}

	left, right := 0, len(posts)-1

	pivot := rand.Int() % len(posts)

	posts[pivot], posts[right] = posts[right], posts[pivot]

	for i := range posts {
		if posts[i].Created.After(posts[right].Created) {
			posts[left], posts[i] = posts[i], posts[left]
			left++
		}
	}

	posts[left], posts[right] = posts[right], posts[left]

	sortposts(posts[:left])
	sortposts(posts[left+1:])

	return posts
}

const metaStartTag = "<!--META--"
const metaEndTag = "--END-->"
const breakTag = "<!--BREAK-->"

func splitMetadata(data []byte) ([]byte, []byte, []byte, error) {
	if data == nil || len(data) == 0 {
		return nil, nil, nil, errors.New("no content in post")
	}
	text := string(data)
	metaEndIdx := strings.Index(text, metaEndTag)
	breakTagIdx := strings.Index(text, breakTag)

	meta := ""
	body := text
	var bodyPreview string
	if strings.HasPrefix(text, metaStartTag) && metaEndIdx >= 0 {
		meta = text[len(metaStartTag):metaEndIdx]
		body = text[metaEndIdx+len(metaEndTag):]

		if breakTagIdx > -1 {
			bodyPreview = text[metaEndIdx+len(metaEndTag) : breakTagIdx]
		}

		meta = strings.TrimSpace(meta)
		body = strings.TrimSpace(body)
	}
	return []byte(meta), []byte(bodyPreview), []byte(body), nil
}

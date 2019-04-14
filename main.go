package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	yaml "gopkg.in/yaml.v2"

	"github.com/gorilla/mux"
	"github.com/russross/blackfriday"
)

// Post is a single blog post
type Post struct {
	Metadata *Metadata
	Content  string
}

// Metadata is the data about the post stored within it
type Metadata struct {
	Title    string
	Subtitle string
	Author   string
	Created  time.Time
	Edited   time.Time
	Tags     []string
}

// Posts contains all blog posts
var Posts []*Post

const port = "8080"

func main() {
	log.Println("loading files...")
	loadFiles()

	log.Println("configuring mux...")
	r := mux.NewRouter()
	r.HandleFunc("/", getPosts)
	r.HandleFunc("/refresh", refresh)
	http.Handle("/", r)

	log.Printf("listening for requests at %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
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

func refresh(w http.ResponseWriter, r *http.Request) {
	loadFiles()
}

func loadFiles() error {
	files, err := ioutil.ReadDir("./")
	if err != nil {
		return err
	}

	Posts = make([]*Post, 0)

	// Must close the encoder when finished to flush any partial blocks.
	// If you comment out the following line, the last partial block "r"
	// won't be encoded.

	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".md") {
			continue
		}

		data, err := ioutil.ReadFile(f.Name())
		if err != nil {
			log.Printf("unable to read post: %s; error: %v\n", f.Name(), err)
			continue
		}

		metaData, body, err := splitMetadata(data)
		if err != nil || len(metaData) == 0 {
			log.Printf("unable to split metadata for post: %s; error: %v\n", f.Name(), err)
			continue
		}

		meta := &Metadata{}
		err = yaml.Unmarshal(metaData, meta)
		if err != nil {
			log.Printf("unable to parse metadata for post: %s; error: %v\n", f.Name(), err)
			continue
		}

		body = blackfriday.MarkdownCommon(body)

		// TODO Still need to build this stuff
		Posts = append(Posts, &Post{
			Metadata: meta,
			Content:  base64.StdEncoding.EncodeToString(body),
		})
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
		if posts[i].Metadata.Created.After(posts[right].Metadata.Created) {
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

func splitMetadata(data []byte) ([]byte, []byte, error) {
	if data == nil || len(data) == 0 {
		return nil, nil, errors.New("no content in post")
	}
	text := string(data)
	metaEndIdx := strings.Index(text, metaEndTag)

	meta := ""
	body := text
	if strings.HasPrefix(text, metaStartTag) && metaEndIdx >= 0 {
		meta = text[len(metaStartTag):metaEndIdx]
		body = text[metaEndIdx+len(metaEndTag):]

		meta = strings.TrimSpace(meta)
		body = strings.TrimSpace(body)
	}
	return []byte(meta), []byte(body), nil
}

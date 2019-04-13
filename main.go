package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/russross/blackfriday"
)

// Post is a single blog post
type Post struct {
	metadata Metadata
	content  []byte
}

// Metadata is the data about the post stored within it
type Metadata struct {
	created time.Time
	author  string
}

type postResponse struct {
	posts []string
}

// Posts contains all blog posts in chronological order
var Posts []*Post

const port = "8080"

func main() {
	log.Println("loading files...")
	loadFiles()

	log.Println("configuring mux")
	r := mux.NewRouter()
	r.HandleFunc("/", getPosts)
	r.HandleFunc("/refresh", refresh)
	http.Handle("/", r)

	log.Printf("listening for requests: %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func getPostsResponse() *postResponse {
	postArr := make([]string, 0)

	for _, p := range Posts {
		postArr = append(postArr, string(p.content))
	}

	response := new(postResponse)
	response.posts = postArr

	return response
}

func getPosts(w http.ResponseWriter, r *http.Request) {
	response := getPostsResponse()

	js, err := json.Marshal(response)
	log.Println(string(js))

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func refresh(w http.ResponseWriter, r *http.Request) {
	loadFiles()
}

func loadFiles() {
	files, err := ioutil.ReadDir("./")
	if err != nil {
		log.Panic(err)
	}

	Posts = make([]*Post, 0)

	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".md") {
			data, err := ioutil.ReadFile(f.Name())
			if err != nil {
				log.Printf("unable to read post: %s\n", f.Name())
			} else {
				// TODO This needs to load the converted data
				blackfriday.MarkdownCommon(data)
				// TODO Still need to build this stuff
				// Posts = append(Posts, &Post{data})
			}
		}
	}
}

const metaStartTag = "==> META"
const metaEndTag = "<== END"

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
	}
	return []byte(meta), []byte(body), nil
}

// TODO parse the array of bytes, splitting results into a key value map
func parseMetadata(data []byte) (*Post, error) {
	if data == nil || len(data) == 0 {
		return nil, errors.New("no content in post")
	}
	startMetaData := -1

	for i, _ := range string(data) {
		// if the file doesn't start with a meta tag, assume
		// no meta data
		if i < len(metaEndTag) {
			if data[i] != metaStartTag[i] {
				break
			}
		} else if startMetaData == -1 {
			startMetaData = i
		}
	}
	//post := &Post{author: "", content: nil, created: time.Now()}

	return nil, nil
}

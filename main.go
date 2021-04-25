package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	yaml "gopkg.in/yaml.v2"

	"github.com/russross/blackfriday/v2"
	"github.com/sksmith/blog-server/views"
)

type RepoFile struct {
	Name        string `json:"name"`
	DownloadUrl string `json:"download_url"`
}

type PostFile struct {
	Name    string
	Content []byte
}

type IndexModel struct {
	PreviousPage int
	NextPage     int
	Posts        []*Post
	Tags         []string
}

// Post is a single blog post
type Post struct {
	Title          string
	Subtitle       string
	Author         string
	Path           string
	Created        time.Time
	Edited         time.Time
	Tags           []string
	ContentPreview template.HTML
	Content        template.HTML
}

// Tags contains all tags associated with blog posts
var Tags []string

// Posts contains all blog posts
var Posts []*Post

// PostMap contains all posts in 2006-01-02 format
var PostMap map[string]*Post

const port = "8080"

var contact *views.View
var index *views.View
var post *views.View

// following gorilla/mux's recommendations on best practices for
// app server startup
func main() {
	var timeout time.Duration
	flag.DurationVar(&timeout, "graceful-timeout", time.Second*15, "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	flag.Parse()

	blogUrl := os.Getenv("BLOG_REPO_URL")
	if blogUrl == "" {
		panic("please set BLOG_REPO_URL")
	}

	loadTemplates()

	repoFiles, err := getRepoFileList(blogUrl)
	if err != nil {
		panic(err)
	}

	files, err := downloadFiles(repoFiles)
	if err != nil {
		panic(err)
	}

	parseFiles(files)
	sortPosts(Posts)
	configureRoutes()
	startServer(timeout)
}

func configureRoutes() {
	log.Println("configuring routes...")

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/", getHome)
	http.HandleFunc("/refresh", refresh)
	http.HandleFunc("/posts/", getPost)
}

func getPost(w http.ResponseWriter, r *http.Request) {

	id := strings.TrimPrefix(r.URL.Path, "/posts/")
	vars := strings.Split(id, "/")

	key := fmt.Sprintf("%s-%s-%s", vars[0], vars[1], vars[2])
	log.Printf("requested: %s\n", key)

	err := post.Render(w, PostMap[key])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

const postsPerPage = 5

func getHome(w http.ResponseWriter, r *http.Request) {
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	start := page * postsPerPage
	end := start + postsPerPage

	if start > len(Posts) {
		start = len(Posts) - 5
	}
	if end > len(Posts) {
		end = len(Posts)
	}

	previousPage := page - 1
	if previousPage < 0 {
		previousPage = 0
	}

	nextPage := page + 1

	err = index.Render(w, IndexModel{
		PreviousPage: previousPage,
		NextPage:     nextPage,
		Posts:        Posts[start:end],
		Tags:         Tags,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func getContact(w http.ResponseWriter, r *http.Request) {
	err := contact.Render(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func loadTemplates() {
	log.Println("loading templates...")

	index = views.NewView("bootstrap", "views/index.gohtml")
	contact = views.NewView("bootstrap", "views/contact.gohtml")
	post = views.NewView("bootstrap", "views/post.gohtml")
}

func refresh(w http.ResponseWriter, r *http.Request) {
	repoFiles, err := getRepoFileList(os.Getenv("BLOG_REPO_URL"))
	if err != nil {
		panic(err)
	}

	files, err := downloadFiles(repoFiles)
	if err != nil {
		panic(err)
	}

	parseFiles(files)
	sortPosts(Posts)
}

func getRepoFileList(url string) ([]RepoFile, error) {
	log.Printf("downloading posts from %s...", url)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	repoContents := make([]RepoFile, 0)
	if err := json.Unmarshal(body, &repoContents); err != nil {
		return nil, err
	}

	return repoContents, nil
}

func downloadFiles(repoFiles []RepoFile) ([]PostFile, error) {
	files := make([]PostFile, len(repoFiles))
	for i, f := range repoFiles {
		resp, err := http.Get(f.DownloadUrl)
		if err != nil {
			log.Printf("failed to download file=[%s], err=[%v]\n", f.Name, err)
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("failed to read file=[%s], err=[%v]\n", f.Name, err)
		}
		files[i] = PostFile{Name: f.Name, Content: body}
	}
	return files, nil
}

func parseFiles(files []PostFile) error {

	Posts = make([]*Post, 0)
	PostMap = make(map[string]*Post)
	TagMap := make(map[string]bool)

	// Must close the encoder when finished to flush any partial blocks.
	// If you comment out the following line, the last partial block "r"
	// won't be encoded.

	for _, f := range files {
		metaData, bodyPreview, body, err := splitMetadata(f.Content)
		if err != nil {
			log.Printf("unable to split metadata file=[%s] error=[%v]\n", f.Name, err)
			continue
		}
		if len(metaData) == 0 {
			log.Printf("no metadata found, skipping file=[%s]\n", f.Name)
			continue
		}

		post := &Post{}
		err = yaml.Unmarshal(metaData, post)
		if err != nil {
			log.Printf("unable to parse metadata for file=[%s] error=[%v]\n", f.Name, err)
			continue
		}

		for _, tag := range post.Tags {
			TagMap[tag] = true
		}

		if len(bodyPreview) > 0 {
			post.ContentPreview = template.HTML((blackfriday.Run(bodyPreview)))
		}

		post.Path = post.Created.Format("/posts/2006/01/02")
		post.Content = template.HTML(blackfriday.Run(body))

		Posts = append(Posts, post)
		PostMap[post.Created.Format("2006-01-02")] = post
	}

	Tags = make([]string, len(TagMap))
	i := 0
	for tag := range TagMap {
		Tags[i] = tag
		i++
	}

	return nil
}

func sortPosts(posts []*Post) {
	log.Println("sorting posts...")
	sort(posts)
}

func sort(posts []*Post) []*Post {
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

	sort(posts[:left])
	sort(posts[left+1:])

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

func startServer(timeout time.Duration) {
	srv := &http.Server{
		Addr:         "0.0.0.0:" + port,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	c := make(chan os.Signal, 1)

	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	log.Printf("listening for requests at %s", port)

	// Block until we receive our signal.
	<-c

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
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

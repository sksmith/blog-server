package views

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strconv"
	"time"
)

const LayoutDir string = "views/layouts"
const RFC3339 = "2006-01-02T15:04:05Z07:00"

func NewView(layout string, files ...string) *View {
	files = append(layoutFiles(), files...)

	t, err := template.New("bootstrap").
		Funcs(template.FuncMap{
			"postlink": postLink,
			"jsdate":   jsDate,
		}).
		ParseFiles(files...)
	if err != nil {
		panic(err)
	}

	return &View{
		Template: t,
		Layout:   layout,
	}
}

type View struct {
	Template *template.Template
	Layout   string
}

func (v *View) Render(w http.ResponseWriter, data interface{}) error {
	return v.Template.ExecuteTemplate(w, v.Layout, data)
}

func layoutFiles() []string {
	files, err := filepath.Glob(LayoutDir + "/*.gohtml")
	if err != nil {
		panic(err)
	}
	return files
}

func postLink(t time.Time) string {
	year := t.Year()

	month := strconv.Itoa(int(t.Month()))
	if len(month) == 1 {
		month = "0" + month
	}

	day := strconv.Itoa(t.Day())
	if len(day) == 1 {
		day = "0" + day
	}

	return fmt.Sprintf("/posts/%d/%s/%s", year, month, day)
}

func jsDate(t time.Time) string {
	return t.Format(RFC3339)
}

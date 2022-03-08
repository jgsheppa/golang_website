package views

import (
	"bytes"
	"errors"
	"html/template"
	"io"
	"net/http"
	"path/filepath"

	"github.com/gorilla/csrf"
	"github.com/jgsheppa/golang_website/context"
)

var (
	LayoutDir string = "views/layouts/"
	TemplateDir string = "views/"
	TemplateExt string = ".gohtml"
)

func NewView(layout string, files ...string) *View {
	// Prepend and append file paths with "views"
	// and ".gohtml"
	addTemplatePath(files)
	addTemplateExtension(files)

	files = append(files, layoutFiles()...)

	t, err := template.New("").Funcs(template.FuncMap{
		"csrfField": func() (template.HTML, error) {
			return "", errors.New("CSRF field is not implemented")
		},
	}).ParseFiles(files...)
	if err != nil {
		panic(err)
	}

	return &View{
		Template: t,
		Layout: layout,
	}
}

type View struct {
	Template *template.Template
	Layout string
}

func (v *View) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	v.Render(w, r, nil)
}

func (v *View) Render(w http.ResponseWriter, r *http.Request, data interface{}) {
	w.Header().Set("Content-Type", "text/html")
	var vd Data
	switch d := data.(type) {
	case Data:
		vd = d
		// do nothing
	default:
		vd = Data{
			Yield: data,
		}
	}
	
	if alert := getAlert(r); alert != nil && vd.Alert == nil {
		vd.Alert = alert
		clearAlert(w)
	}
	vd.User = context.User(r.Context())
	var buf bytes.Buffer

	csrfField := csrf.TemplateField(r)
	template := v.Template.Funcs(template.FuncMap{
		"csrfField": func() template.HTML {
			return csrfField
		},
	})

	if err := template.ExecuteTemplate(&buf, v.Layout, vd); err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	io.Copy(w, &buf)
}

// Returns a slice of strings 
// representing the layout files
func layoutFiles() []string {
	files, err := filepath.Glob(LayoutDir + "*" + TemplateExt)
	if err != nil {
		panic(err)
	}
	return files
}

// This function takes in a slice of strings
// representing files path for templates
// and it prepends the TemplateDir to each string
// in the slice.
//
// Ex.: "home" would result in "views/home/" if
// TemplateDir == "views/"
func addTemplatePath(files []string) {
	for i, f := range files {
		files[i] = TemplateDir + f
	}
}

func addTemplateExtension(files []string) {
	for i, f := range files {
		files[i] = f + TemplateExt
	}
}
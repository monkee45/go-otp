package views

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

var (
	LayoutDir   string = "views/layouts/"
	TemplateDir string = "views/"
)

type View struct {
	Template *template.Template
}

// Combines the layout files, with any passed as parameters and parses the files
func NewView(files ...string) *View {
	addTemplatePath(files)
	files = append(layoutFiles(), files...)
	t, err := template.ParseFiles(files...)
	if err != nil {
		log.Fatalf("loading template %s failed: %v", files, err)
	}

	return &View{
		Template: t,
	}
}

// addTemplatePath takes in a slice of strings representing filepaths for
// templates and prepends the TemplateDir directory to each string in the slice.
// e.g. input {"home"} would result in the output {"views/home"} if
// the TemplateDir is "views/"
func addTemplatePath(files []string) {
	for i, f := range files {
		files[i] = TemplateDir + f
	}
}

// layoutFiles returns a slice of strings representing the layout files
func layoutFiles() []string {
	files, err := filepath.Glob(LayoutDir + "*")
	if err != nil {
		panic(err)
	}
	return files
}

func (v *View) Render(w http.ResponseWriter, r *http.Request, data any) {
	var vd Data
	switch d := data.(type) {
	case Data:
		vd = d
		// do nothing
	default:
		vd = Data{
			Yield: data, // we assume that whatever they passed in is for the Yield template
		}
	}

	err := v.Template.ExecuteTemplate(w, "base", vd)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("template error: %v", err)
	}
}

func (v *View) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	v.Render(w, r, nil)
}

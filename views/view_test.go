package views

import (
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

// Tests the addTemplatePath function
func TestAddTemplatePlath(t *testing.T) {
	old := TemplateDir
	TemplateDir = "testviews/"
	defer func() { TemplateDir = old }()

	files := []string{"home.gohtml", "about.gohtml"}

	addTemplatePath(files)

	expected := []string{
		"testviews/home.gohtml",
		"testviews/about.gohtml",
	}

	if !reflect.DeepEqual(files, expected) {
		t.Fatalf("expected %v, got %v", expected, files)
	}
}

func TestLayoutFiles(t *testing.T) {
	old := LayoutDir
	LayoutDir = "testdata/layouts/"
	defer func() { LayoutDir = old }()

	files := layoutFiles()
	expected := []string{"testdata/layouts/base.gohtml"}
	if !reflect.DeepEqual(files, expected) {
		t.Fatalf("expected %v, got %v", expected, files)
	}
}

func TestNewView(t *testing.T) {
	oldLayout := LayoutDir
	oldTemplate := TemplateDir

	LayoutDir = "testdata/layouts/"
	TemplateDir = "testdata/"

	defer func() {
		LayoutDir = oldLayout
		TemplateDir = oldTemplate
	}()

	v := NewView("home.gohtml")

	if v == nil {
		t.Fatalf("expected view")
	}

	if v.Template == nil {
		t.Fatalf("expected a parsed template")
	}
}

func TestRender_WithPlainData(t *testing.T) {
	// setupTestTemplates()
	oldLayout := LayoutDir
	oldTemplate := TemplateDir

	LayoutDir = "testdata/layouts/"
	TemplateDir = "testdata/"

	defer func() {
		LayoutDir = oldLayout
		TemplateDir = oldTemplate
	}()

	v := NewView("home.gohtml")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	v.Render(rr, req, "World")

	body := rr.Body.String()

	if !strings.Contains(body, "Hello World") {
		t.Fatalf("unexpected body: %s", body)
	}
}

func TestRender_WithData(t *testing.T) {
	// setupTestTemplates()
	oldLayout := LayoutDir
	oldTemplate := TemplateDir

	LayoutDir = "testdata/layouts/"
	TemplateDir = "testdata/"

	defer func() {
		LayoutDir = oldLayout
		TemplateDir = oldTemplate
	}()

	v := NewView("home.gohtml")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	v.Render(rr, req, Data{
		Yield: "Custom",
	})

	body := rr.Body.String()

	if !strings.Contains(body, "Hello Custom") {
		t.Fatalf("unexpected body: %s", body)
	}
}

func TestRender_TemplateError(t *testing.T) {
	// setupTestTemplates()
	oldLayout := LayoutDir
	oldTemplate := TemplateDir

	LayoutDir = "testdata/layouts/"
	TemplateDir = "testdata/"

	defer func() {
		LayoutDir = oldLayout
		TemplateDir = oldTemplate
	}()

	v := NewView("broken.gohtml")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	v.Render(rr, req, "data")

	log.Printf("\nrr.Code = %v\n", rr.Code)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 got %d", rr.Code)
	}
}

func TestServeHTTP(t *testing.T) {
	// setupTestTemplates()
	oldLayout := LayoutDir
	oldTemplate := TemplateDir

	LayoutDir = "testdata/layouts/"
	TemplateDir = "testdata/"

	defer func() {
		LayoutDir = oldLayout
		TemplateDir = oldTemplate
	}()

	v := NewView("home.gohtml")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	v.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rr.Code)
	}
}

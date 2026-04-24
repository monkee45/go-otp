package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"photos.com/controllers"
)

func TestContactPage(t *testing.T) {
	wr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/contact", nil)
	staticC := controllers.NewStaticController()

	staticC.Contact.Render(wr, req, nil)
	if wr.Code != http.StatusOK {
		t.Errorf("got HTTP status code %d, expected 200", wr.Code)
	}

	if !strings.Contains(wr.Body.String(), "Mike Walsh") {
		t.Errorf(
			"Expected `Mike Walsh`, got %v", wr.Body.String(),
		)
	}
}

func TestHomePage(t *testing.T) {
	wr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/home", nil)
	staticC := controllers.NewStaticController()

	staticC.Home.Render(wr, req, nil)
	if wr.Code != http.StatusOK {
		t.Errorf("got HTTP status code %d, expected 200", wr.Code)
	}
	if !strings.Contains(wr.Body.String(), "Create and share Stories") {
		t.Errorf(
			"Expected `Create and share Stories`, got %v", wr.Body.String(),
		)
	}
}

func TestAboutPage(t *testing.T) {
	wr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/about", nil)
	staticC := controllers.NewStaticController()

	staticC.About.Render(wr, req, nil)
	if wr.Code != http.StatusOK {
		t.Errorf("got HTTP status code %d, expected 200", wr.Code)
	}

	if !strings.Contains(wr.Body.String(), "Relive your favourite memories") {
		t.Errorf(
			"Expected `Relive your favourite memories`, got %v", wr.Body.String(),
		)
	}
}

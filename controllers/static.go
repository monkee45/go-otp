package controllers

import "photos.com/views"

type Static struct {
	Home    *views.View
	About   *views.View
	Contact *views.View
}

func NewStaticController() *Static {
	return &Static{
		Home:    views.NewView("static/home.html"),
		About:   views.NewView("static/about.html"),
		Contact: views.NewView("static/contact.html"),
	}
}

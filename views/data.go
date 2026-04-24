package views

import "photos.com/models"

const (
	AlertLvlError   = "danger"
	AlertLvlWarning = "warning"
	AlertLvlInfo    = "info"
	AlertLvlSuccess = "success"
)

// Alert is used to render bootstrap alert messages in templates
type Alert struct {
	Level   string
	Message string
}

// Data is the top level structure that views expect data to come from
type Data struct {
	Alert *Alert
	User  *models.User
	Yield any
}

package mailer

import "embed"

const (
	fromName           = "GopherSocial"
	maxRetries         = 3
	UserInviteTemplate = "user_inivatation.tmpl"
)

//go:embed "templates"
var FS embed.FS

type Client interface {
	Send(templateFile, username, email string, data any, isSendbox bool) error
}

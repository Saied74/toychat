package main

import (
	"os"
	"path/filepath"
)

//constants used in the handlers in place of the strings to avoid typing mistakes
const (
	home                = "home"
	login               = "login"
	signup              = "signup"
	chat                = "chat"
	mat                 = "mat"
	table               = "table"
	authenticatedUserID = "authenticatedUserID"
	GET                 = "GET"
	POST                = "POST"
	superHome           = "/super/home"
	superLogin          = "/super/login"
	superLogout         = "/super/logout"
	adminHome           = "/admin/home"
	adminLogin          = "/admin/login"
	adminLogout         = "/admin/logout"
	agentHome           = "/agent/home"
	agentLogin          = "/agent/login"
	agentLogout         = "/agent/logout"
	admins              = "admins"
	msg                 = "Please log in"
	addAdmin            = "/super/addAdmin"
	activateAdmin       = "/super/activateAdmin"
	deactivateAdmin     = "/super/deactivateAdmin"
	addAgent            = "/admin/addAgent"
	activateAgent       = "/admin/activateAgent"
	deactivateAgent     = "/admin/deactivateAgent"
)

var allTmplFiles = map[string][]string{
	"home": []string{
		filepath.Join(os.Getenv("GOPATH"), "src/toychat/backend/backendviews/base.tmpl"),
		filepath.Join(os.Getenv("GOPATH"), "src/toychat/backend/backendviews/home.tmpl"),
	},
	"login": []string{
		filepath.Join(os.Getenv("GOPATH"), "src/toychat/backend/backendviews/base.tmpl"),
		filepath.Join(os.Getenv("GOPATH"), "src/toychat/backend/backendviews/login.tmpl"),
	},
	"signup": []string{
		filepath.Join(os.Getenv("GOPATH"), "src/toychat/backend/backendviews/base.tmpl"),
		filepath.Join(os.Getenv("GOPATH"), "src/toychat/backend/backendviews/signup.tmpl"),
	},
	"chat": []string{
		filepath.Join(os.Getenv("GOPATH"), "src/toychat/backend/backendviews/base.tmpl"),
		filepath.Join(os.Getenv("GOPATH"), "src/toychat/backend/backendviews/chat.tmpl"),
		filepath.Join(os.Getenv("GOPATH"), "src/toychat/backend/backendviews/play.tmpl"),
	},
	"mat": []string{
		filepath.Join(os.Getenv("GOPATH"), "src/toychat/backend/backendviews/base.tmpl"),
		filepath.Join(os.Getenv("GOPATH"), "src/toychat/backend/backendviews/mat.tmpl"),
		filepath.Join(os.Getenv("GOPATH"), "src/toychat/backend/backendviews/playmat.tmpl"),
	},
	"table": []string{
		filepath.Join(os.Getenv("GOPATH"), "src/toychat/backend/backendviews/base.tmpl"),
		filepath.Join(os.Getenv("GOPATH"), "src/toychat/backend/backendviews/table.tmpl"),
	},
}

//Self signed keys.  Works on Safari on Mac, Chrome constantly complains
var serverKey = filepath.Join(os.Getenv("GOPATH"), "src/toychat/certs/https-server.key")
var serverCrt = filepath.Join(os.Getenv("GOPATH"), "src/toychat/certs/https-server.crt")

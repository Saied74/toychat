package main

import (
	"os"
	"path/filepath"
)

const (
	home                = "home"
	login               = "login"
	signup              = "signup"
	chat                = "chat"
	mat                 = "mat"
	authenticatedUserID = "authenticatedUserID"
)

var allTmplFiles = map[string][]string{
	"home": []string{
		filepath.Join(os.Getenv("GOPATH"), "src/toychat/frontend/views/base.tmpl"),
		filepath.Join(os.Getenv("GOPATH"), "src/toychat/frontend/views/home.tmpl"),
	},
	"login": []string{
		filepath.Join(os.Getenv("GOPATH"), "src/toychat/frontend/views/base.tmpl"),
		filepath.Join(os.Getenv("GOPATH"), "src/toychat/frontend/views/login.tmpl"),
	},
	"signup": []string{
		filepath.Join(os.Getenv("GOPATH"), "src/toychat/frontend/views/base.tmpl"),
		filepath.Join(os.Getenv("GOPATH"), "src/toychat/frontend/views/signup.tmpl"),
	},
	"chat": []string{
		filepath.Join(os.Getenv("GOPATH"), "src/toychat/frontend/views/base.tmpl"),
		filepath.Join(os.Getenv("GOPATH"), "src/toychat/frontend/views/chat.tmpl"),
		filepath.Join(os.Getenv("GOPATH"), "src/toychat/frontend/views/play.tmpl"),
	},
	"mat": []string{
		filepath.Join(os.Getenv("GOPATH"), "src/toychat/frontend/views/base.tmpl"),
		filepath.Join(os.Getenv("GOPATH"), "src/toychat/frontend/views/mat.tmpl"),
		filepath.Join(os.Getenv("GOPATH"), "src/toychat/frontend/views/playmat.tmpl"),
	},
}

var serverKey = filepath.Join(os.Getenv("GOPATH"), "src/toychat/certs/https-server.key")
var serverCrt = filepath.Join(os.Getenv("GOPATH"), "src/toychat/certs/https-server.crt")

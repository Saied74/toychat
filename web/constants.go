package main

const (
	home                = "home"
	login               = "login"
	signup              = "signup"
	chat                = "chat"
	mat                 = "mat"
	authenticatedUserID = "authenticatedUserID"
)

var templateFiles = []string{"../views/base.tmpl", "../views/home.tmpl"}
var homeFiles = []string{"../views/base.tmpl", "../views/home.tmpl"}
var loginFiles = []string{"../views/base.tmpl", "../views/login.tmpl"}
var signupFiles = []string{"../views/base.tmpl", "../views/signup.tmpl"}
var chatFiles = []string{"../views/base.tmpl", "../views/chat.tmpl",
	"../views/play.tmpl"}
var matFiles = []string{"../views/base.tmpl", "../views/mat.tmpl",
	"../views/playmat.tmpl"}

var allTmplFiles = map[string][]string{
	"home":   []string{"../views/base.tmpl", "../views/home.tmpl"},
	"login":  []string{"../views/base.tmpl", "../views/login.tmpl"},
	"signup": []string{"../views/base.tmpl", "../views/signup.tmpl"},
	"chat":   []string{"../views/base.tmpl", "../views/chat.tmpl", "../views/play.tmpl"},
	"mat":    []string{"../views/base.tmpl", "../views/mat.tmpl", "../views/playmat.tmpl"},
}

var serverKey = "../certs/https-server.key"
var serverCrt = "../certs/https-server.crt"

type errMsg int

const (
	noErr         errMsg = iota //no error
	errZero                     //simple error
	noRecord                    //errNoRecord
	invalidCreds                //errInvalidCredentials
	duplicateMail               //errDuplicateEmail
)

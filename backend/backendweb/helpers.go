package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"runtime/debug"
	"strings"
	"time"

	"github.com/justinas/nosurf"
	nats "github.com/nats-io/nats.go"
	"github.com/saied74/toychat/pkg/broker"
	"github.com/saied74/toychat/pkg/forms"
)

var getToken = nosurf.Token
var getUser = broker.GetUserR
var isAuth = isAuthenticated

//These two loggers are written so one can pass other writer opbjects to them
//for testing (to write to a buffer) and also for logging to file at some point.
//these loggers will have to be moved to a package file.
func getInfoLogger(out io.Writer) func() *log.Logger {
	infoLog := log.New(out, "INFO\t", log.Ldate|log.Ltime)
	return func() *log.Logger {
		return infoLog
	}
}

func getErrorLogger(out io.Writer) func() *log.Logger {
	errorLog := log.New(out, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	return func() *log.Logger {
		return errorLog
	}
}

//consolidated screen error reporting.  Once the chat manager application
//is written, these html error sceen handlers need to be moved to a package file.
// TODO: make the html error message pages nice, they are ugly.
func (app *App) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.errorLog.Println(trace)

	http.Error(w, http.StatusText(http.StatusInternalServerError),
		http.StatusInternalServerError)
}

func (app *App) clientError(w http.ResponseWriter, status int, err error) {
	app.errorLog.Printf("client error %v", err)

	http.Error(w, http.StatusText(status), status)
}

func (app *App) pickPath(w http.ResponseWriter, r *http.Request) error {
	app.initTD()
	path := strings.Split(r.URL.Path, "/")
	if len(path) < 3 {
		return fmt.Errorf("bad path %s, short string", r.URL.Path)
	}
	switch path[1] {
	case "super":
		app.buildSuper()
		switch path[2] {
		case "activateAdmin":
			app.td.Active = true
		case "deactivateAdmin":
			app.td.Active = false
		case "changePassword":
			app.td.Msg = pwdMsg
		}
	case "admin":
		app.buildAdmin()
		switch path[2] {
		case "activateAgent":
			app.td.Active = true
		case "deactivateAgent":
			app.td.Active = false
		case "changePassword":
			app.td.Msg = pwdMsg
		}
	case agent:
		app.buildAgent()
	default:
		return fmt.Errorf("bad path %s", r.URL.Path)
	}
	if strings.HasPrefix(path[2], "add") {
		app.td.Msg = addMsg
	}
	return nil
}

func (app *App) buildSuper() {
	app.table = admins
	app.role = "superadmin"
	app.nextRole = admin
	app.redirect = superHome
	app.td.Scope = "Super User"
	app.td.Home = superHome
	app.td.Login = superLogin
	app.td.Logout = superLogout
	app.td.ChgPwd = ""
	app.td.SideLink1 = addAdmin
	app.td.SideLink2 = activateAdmin
	app.td.SideLink3 = deactivateAdmin
	app.td.Super = true
	app.td.Admin = false
	app.td.Agent = false
	app.td.Msg = loginMsg
}

func (app *App) buildAdmin() {
	app.table = admins
	app.role = admin
	app.nextRole = agent
	app.redirect = adminHome
	app.td.Scope = "Admin User"
	app.td.Home = adminHome
	app.td.Login = adminLogin
	app.td.Logout = adminLogout
	app.td.ChgPwd = adminChgPwd
	app.td.SideLink1 = addAgent
	app.td.SideLink2 = activateAgent
	app.td.SideLink3 = deactivateAgent
	app.td.Super = false
	app.td.Admin = true
	app.td.Agent = false
	app.td.Msg = loginMsg
}

func (app *App) buildAgent() {
	app.table = admins
	app.role = agent
	app.nextRole = ""
	app.redirect = agentHome
	app.td.Scope = "Agent"
	app.td.Home = agentHome
	app.td.Login = agentLogin
	app.td.Logout = agentLogout
	app.td.ChgPwd = agentChgPwd
	app.td.SideLink1 = agentOnline
	app.td.SideLink2 = agentOffline
	app.td.SideLink3 = ""
	app.td.Super = false
	app.td.Admin = false
	app.td.Agent = true
	app.td.Msg = loginMsg
}

func (app *App) initTD() {
	app.td = &templateData{
		Form: &forms.FormData{
			Fields: url.Values{},
			Errors: forms.ErrOrs{},
		},
	}
}

type tmData map[string][]string

type tmDataer interface {
	tmpData() (*tmData, error)
}

func (inTM tmData) tmpData() (*tmData, error) {
	in := inTM
	var out = tmData{}
	for key, files := range in {
		strList := []string{}
		for _, file := range files {
			tmStr, err := ioutil.ReadFile(file)
			if err != nil {
				return &tmData{}, err
			}
			strList = append(strList, string(tmStr))
		}
		out[key] = strList
	}
	return &out, nil
}

//templates are cashed by name to avoid repeated disk access.
func newTemplateCache(tmpl tmDataer) map[string]*template.Template {
	tc := map[string]*template.Template{}
	var t *template.Template
	tmptr, err := tmpl.tmpData()
	if err != nil {
		log.Fatal(err)
	}
	tm := *tmptr
	for key, data := range tm {
		t = template.New(key)
		for _, datum := range data {
			t = template.Must(t.Parse(datum))
		}
		tc[key] = t
	}
	return tc
}

//used by the middleware wrapping handlers that need authentication.  In the
//current situation, the chat (but not the mat) application only.
func isAuthenticated(r *http.Request) bool {
	isAuthenticated, ok := r.Context().Value(contextKeyIsAuthenticated).(bool)
	if !ok {
		return false
	}
	return isAuthenticated
}

//adds authenicated users name, authenicated flag, and the csrf token to the form.
func addDefaultData(td *templateData, r *http.Request,
	app *App) (*templateData, error) {

	if td == nil {
		td = &templateData{}
	}
	td.Flash = app.sessionManager.PopString(r.Context(), "flash")
	td.LoggedIn = isAuth(r)
	id := app.sessionManager.GetInt(r.Context(), authenticatedUserID)
	if td.LoggedIn {
		usr, err := getUser("admins", id) //broker.GetUserR("admins", id)
		if err != nil {
			return nil, err
		}
		td.UserName = string(usr.Name)
	}
	td.CSRFToken = getToken(r) //nosurf.Token(r)
	return td, nil
}

//writes the form to a buffer to check for error prior to writing the response.
func (app *App) render(w http.ResponseWriter, r *http.Request, name string) {
	t := app.cache[name]
	buf := new(bytes.Buffer)
	tData, err := addDefaultData(app.td, r, app)
	if err != nil {
		app.serverError(w, err)
	}
	err = t.Execute(buf, tData)
	if err != nil {
		app.serverError(w, err)
		return
	}
	buf.WriteTo(w)
}

//sends string data to the far end, waits for the response and returns.
//for chat and mat, the data is string.  For dbmgr, the data is a struct.
//which is gob encoded before it is sent.  Gob encoder is in the broker pkg.
// TODO: find a way to build a nats connecton pool like the MySQL connection
//pool to speed up transactions.
func (app *App) chatConnection(matValue, forCM, fromCM string) []byte {
	var err error

	nc1, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		app.errorLog.Printf("in chatConnection connecting error %v", err)
	}
	defer nc1.Close()
	msg, err := nc1.Request(forCM, []byte(matValue), 2*time.Second)
	if err != nil {
		app.errorLog.Printf("in chatConnection %s request did not complete %v",
			forCM, err)
		return []byte{}
	}
	return msg.Data
}

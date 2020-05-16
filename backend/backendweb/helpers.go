package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/justinas/nosurf"
	nats "github.com/nats-io/nats.go"
	"github.com/saied74/toychat/pkg/models"
)

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

func (app *App) buildSuper() {
	app.td.table = admins
	app.td.role = "superadmin"
	app.td.nextRole = "admin"
	app.td.Scope = "Super User"
	app.td.Home = "/super/home"
	app.td.Login = "/super/login"
	app.td.Logout = "/super/logout"
	app.td.SideLink1 = addAdmin
	app.td.SideLink2 = activateAdmin
	app.td.SideLink3 = deactivateAdmin
	app.td.ChgPWD = ""
	app.td.Super = true
	app.td.Admin = false
	app.td.Agent = false
	app.td.Msg = msg
}

func (app *App) buildAdmin() {
	app.td.table = admins
	app.td.role = "admin"
	app.td.nextRole = "agent"
	app.td.Scope = "Admin User"
	app.td.Home = "/admin/home"
	app.td.Login = "/admin/login"
	app.td.Logout = "/admin/logout"
	app.td.SideLink1 = addAgent
	app.td.SideLink2 = activateAgent
	app.td.SideLink3 = deactivateAgent
	app.td.ChgPWD = "/admin/logout"
	app.td.Super = false
	app.td.Admin = true
	app.td.Agent = false
	app.td.Msg = msg
}

func (app *App) buildAgent() {
	app.td.table = admins
	app.td.role = ""
	app.td.Scope = "Agent"
	app.td.Home = "/agent/home"
	app.td.Login = "/agent/login"
	app.td.Logout = "/agent/logout"
	app.td.SideLink1 = ""
	app.td.SideLink2 = ""
	app.td.SideLink3 = ""
	app.td.ChgPWD = "/agent/logout"
	app.td.Super = false
	app.td.Admin = false
	app.td.Agent = true
	app.td.Msg = msg
}

//templates are cashed by name to avoid repeated disk access.
func newTemplateCache(tmpls map[string][]string) map[string]*template.Template {
	tc := map[string]*template.Template{}
	for key, files := range tmpls {
		t := template.Must(template.ParseFiles(files...))
		tc[key] = t
	}
	return tc
}

//used by the middleware wrapping handlers that need authentication.  In the
//current situation, the chat (but not the mat) application only.
func (app *App) isAuthenticated(r *http.Request) bool {
	isAuthenticated, ok := r.Context().Value(contextKeyIsAuthenticated).(bool)
	if !ok {
		return false
	}
	return isAuthenticated
}

//adds authenicated users name, authenicated flag, and the csrf token to the form.
func (app *App) addDefaultData(td *templateData, r *http.Request) (*templateData, error) {

	if td == nil {
		td = &templateData{}
	}
	// td.Home = "home"
	td.Flash = app.sessionManager.PopString(r.Context(), "flash")
	td.LoggedIn = app.isAuthenticated(r)
	id := app.sessionManager.GetInt(r.Context(), authenticatedUserID)
	if td.LoggedIn {
		usr, err := models.GetUserR("admins", id)
		if err != nil {
			return nil, err
		}
		td.UserName = string(usr.Name)
	}
	td.CSRFToken = nosurf.Token(r)
	return td, nil
}

//writes the form to a buffer to check for error prior to writing the response.
func (app *App) render(w http.ResponseWriter, r *http.Request, name string) {
	t := app.cache[name]
	buf := new(bytes.Buffer)
	tData, err := app.addDefaultData(app.td, r)
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

package main

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"runtime/debug"
	"time"

	"github.com/justinas/nosurf"
	"github.com/nats-io/nats.go"
	"github.com/saied74/toychat/pkg/broker"
	"github.com/saied74/toychat/pkg/centerr"
	"github.com/saied74/toychat/pkg/forms"
)

//consolidated screen error reporting.  Once the chat manager application
//is written, these html error sceen handlers need to be moved to a package file.
// TODO: make the html error message pages nice, they are ugly.
func (st *sT) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	centerr.ErrorLog.Println(trace)

	http.Error(w, http.StatusText(http.StatusInternalServerError),
		http.StatusInternalServerError)
}

func (st *sT) clientError(w http.ResponseWriter, status int, err error) {
	centerr.ErrorLog.Printf("client error %v", err)

	http.Error(w, http.StatusText(status), status)
}

func (st *sT) initTD() {
	st.td = &templateData{
		Form: &forms.FormData{
			Fields: url.Values{},
			Errors: forms.ErrOrs{},
		},
	}
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
func (st *sT) isAuthenticated(r *http.Request) bool {
	isAuthenticated, ok := r.Context().Value(contextKeyIsAuthenticated).(bool)
	if !ok {
		return false
	}
	return isAuthenticated
}

//adds authenicated users name, authenicated flag, and the csrf token to the form.
func (st *sT) addDefaultData(td *templateData, r *http.Request) (*templateData, error) {

	if td == nil {
		td = &templateData{}
	}
	td.Flash = st.sessionManager.PopString(r.Context(), "flash")
	td.LoggedIn = st.isAuthenticated(r)
	id := st.sessionManager.GetInt(r.Context(), authenticatedUserID)
	if td.LoggedIn {
		usr, err := broker.GetEUR("users", id)
		if err != nil {
			return nil, err
		}
		td.UserName = string(usr.Name)
	}
	td.CSRFToken = nosurf.Token(r)
	return td, nil
}

//writes the form to a buffer to check for error prior to writing the response.
func (st *sT) render(w http.ResponseWriter, r *http.Request, name string) {
	t := st.cache[name]
	buf := new(bytes.Buffer)
	tData, err := st.addDefaultData(st.td, r)
	if err != nil {
		st.serverError(w, err)
	}
	err = t.Execute(buf, tData)
	if err != nil {
		st.serverError(w, err)
		return
	}
	buf.WriteTo(w)
}

//sends string data to the far end, waits for the response and returns.
//for chat and mat, the data is string.  For dbmgr, the data is a struct.
//which is gob encoded before it is sent.  Gob encoder is in the broker pkg.
// TODO: find a way to build a nats connecton pool like the MySQL connection
//pool to speed up transactions.
func (st *sT) chatConnection(matValue, forCM, fromCM string) []byte {
	var err error

	nc1, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		centerr.ErrorLog.Printf("in chatConnection connecting error %v", err)
	}
	defer nc1.Close()
	msg, err := nc1.Request(forCM, []byte(matValue), 2*time.Second)
	if err != nil {
		centerr.ErrorLog.Printf("in chatConnection %s request did not complete %v",
			forCM, err)
		return []byte{}
	}
	return msg.Data
}

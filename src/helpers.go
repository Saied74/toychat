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
)

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

func (st *sT) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	st.errorLog.Println(trace)

	http.Error(w, http.StatusText(http.StatusInternalServerError),
		http.StatusInternalServerError)
}

func (st *sT) clientError(w http.ResponseWriter, status int, err error) {
	st.errorLog.Printf("client error %v", err)

	http.Error(w, http.StatusText(status), status)
}

func newTemplateCache(tmpls map[string][]string) map[string]*template.Template {

	tc := map[string]*template.Template{}
	for key, files := range tmpls {
		t := template.Must(template.ParseFiles(files...))
		tc[key] = t
	}
	return tc
}

func (st *sT) isAuthenticated(r *http.Request) bool {
	isAuthenticated, ok := r.Context().Value(contextKeyIsAuthenticated).(bool)
	if !ok {
		return false
	}
	return isAuthenticated
}

func (st *sT) addDefaultData(td *templateData, r *http.Request) (*templateData, error) {

	if td == nil {
		td = &templateData{}
	}
	td.Flash = st.sessionManager.PopString(r.Context(), "flash")
	td.LoggedIn = st.isAuthenticated(r)
	id := st.sessionManager.GetInt(r.Context(), authenticatedUserID)
	if td.LoggedIn {
		usr, err := st.users.getUser(id)
		if err != nil {
			return nil, err
		}
		td.UserName = string(usr.Name)
	}
	td.CSRFToken = nosurf.Token(r)
	return td, nil
}

func (st *sT) render(w http.ResponseWriter, r *http.Request, name string) {
	// t := template.Must(template.ParseFiles(files...))
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

func (st *sT) matConnect(matValue string) []byte {
	var err error
	var m = &nats.Msg{}

	nc1, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatal("Error from onnection", err)
	}
	defer nc1.Close()

	nc2, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatal("Error from onnection", err)
	}
	defer nc2.Close()

	nc2.Publish("forMat", matValue)
	sub, err := nc1.SubscribeSync("fromMat")
	if err != nil {
		log.Fatal("Error from Sub Sync: ", err)
	}
	m, err = sub.NextMsg(20 * time.Hour)
	if err != nil {
		log.Fatal("Error from next message, timed out: ", err)
	}
	matValue := playMatHandler(string(m.Data))
	fmt.Printf("Message from the far side: %s\n", string(m.Data))
	return m.Data

}

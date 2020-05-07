package main

import (
	"errors"
	"net/http"
)

// TODO: in general, these handlers are weak with respect to transport error
func (st *sT) homeHandler(w http.ResponseWriter, r *http.Request) {
	st.render(w, r, home)
}

func (st *sT) loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/login" {
		st.errorLog.Printf("bad path %s", r.URL.Path)
		http.NotFound(w, r)
		return
	}
	switch r.Method {

	case "GET":
		st.render(w, r, login)

	case "POST":
		err := r.ParseForm()
		if err != nil {
			st.clientError(w, http.StatusBadRequest, err)
			return
		}
		st.td.Form = st.newForm(r.PostForm, r)
		id, err := st.authenticateUserR(st.td.Form.getField("email"),
			st.td.Form.getField("password"))
		if err != nil {
			if errors.Is(err, errInvalidCredentials) { //|| errors.Is(err, errNoRecord)
				st.td.Form.Errors.addError("generic", "Email or Password is incorrect")
				st.render(w, r, login)
			} else {
				st.serverError(w, err)
			}
			return
		}

		st.sessionManager.RenewToken(r.Context())
		st.sessionManager.Put(r.Context(), authenticatedUserID, id)
		http.Redirect(w, r, home, http.StatusSeeOther)

	default:
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte(http.StatusText(http.StatusNotImplemented)))
	}
}

func (st *sT) signupHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/signup" {
		st.errorLog.Printf("bad path %s", r.URL.Path)
		http.NotFound(w, r)
		return
	}

	switch r.Method {

	case "GET":
		st.infoLog.Printf("got to get %s", r.Method)
		st.render(w, r, signup)

	case "POST":
		err := r.ParseForm()
		if err != nil {
			st.clientError(w, http.StatusBadRequest, err)
		}
		Form := st.newForm(r.PostForm, r)
		Form.fieldRequired("name", "email", "password")
		Form.maxLength("name", 255)
		Form.maxLength("email", 255)
		Form.matchPattern("email", emailRX)
		Form.minLength("password", 10)

		if !Form.valid() {
			st.render(w, r, signup)
			return
		}
		err = st.insertUserR(Form.getField("name"), Form.getField("email"),
			Form.getField("password"))
		if err != nil {
			st.errorLog.Printf("Fatal Error %v", err)
			if errors.Is(err, errDuplicateEmail) {
				Form.Errors.addError("email", "Address is already in use")
				st.render(w, r, signup)
			} else {
				st.serverError(w, err)
			}
			return
		}
		st.sessionManager.RenewToken(r.Context())
		st.sessionManager.Put(r.Context(), "flash", "Your signup was successful, pleaselogin")
		http.Redirect(w, r, "/login", http.StatusSeeOther)

	default:
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte(http.StatusText(http.StatusNotImplemented)))
	}
}
func (st *sT) logoutHandler(w http.ResponseWriter, r *http.Request) {
	st.sessionManager.RenewToken(r.Context())
	st.sessionManager.Remove(r.Context(), authenticatedUserID)
	http.Redirect(w, r, "/home", http.StatusSeeOther)
}

func (st *sT) chatHandler(w http.ResponseWriter, r *http.Request) {
	st.render(w, r, chat)
}

func (st *sT) matHandler(w http.ResponseWriter, r *http.Request) {
	st.render(w, r, mat)
}

func (st *sT) playHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm() //parse request, handle error
	if err != nil {
		st.errorLog.Println(err)
	}
	value, ok := r.Form["value"]
	if ok {
		chatValue := st.chatConnection(value[0], "forChat", "fromChat")
		w.Write(chatValue)
	}
}

func (st *sT) playMatHandler(w http.ResponseWriter, r *http.Request) {
	// var matValue = ""
	err := r.ParseForm() //parse request, handle error
	if err != nil {
		st.errorLog.Println(err)
	}
	value, ok := r.Form["value"]
	if ok {
		matValue := st.chatConnection(value[0], "forMat", "fromMat")
		w.Write(matValue)
	}
}

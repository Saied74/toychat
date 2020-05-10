package main

import (
	"errors"
	"net/http"

	//broker pkg contains the code that is used on both sides of the nats connectoin.
	"github.com/saied74/toychat/pkg/broker"
)

// TODO: these handlers might be weak with respect to nats transport error

//much of the commmon work is done in render, addDefaultData and middlewares
func (st *sT) homeHandler(w http.ResponseWriter, r *http.Request) {
	st.render(w, r, home)
}

//using the golang standard library, so need to check for correct path and method.
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

		//authenticateUserR R stands for remote sends the data to the dbmgr over
		//the nats connectoin to be validated.
		id, err := st.authenticateUserR(st.td.Form.getField("email"),
			st.td.Form.getField("password"))
		if err != nil {
			if errors.Is(err, broker.ErrInvalidCredentials) {
				st.td.Form.Errors.addError("generic", "Email or Password is incorrect")
				st.render(w, r, login)
			} else {
				st.serverError(w, err)
			}
			return
		}
		//RenewToken is used for security purpose for each state change.
		st.sessionManager.RenewToken(r.Context())
		st.sessionManager.Put(r.Context(), authenticatedUserID, id)
		http.Redirect(w, r, home, http.StatusSeeOther)

	default:
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte(http.StatusText(http.StatusNotImplemented)))
	}
}

//The same here as login with path and method checking.
func (st *sT) signupHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/signup" {
		st.errorLog.Printf("bad path %s", r.URL.Path)
		http.NotFound(w, r)
		return
	}

	switch r.Method {

	case "GET":
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
		//once the form is validated (above), it is sent to the dbmgr over nats
		//to be inserted into the database.
		err = st.insertUserR(Form.getField("name"), Form.getField("email"),
			Form.getField("password"))
		if err != nil {
			st.errorLog.Printf("Fatal Error %v", err)
			if errors.Is(err, broker.ErrDuplicateEmail) {
				Form.Errors.addError("email", "Address is already in use")
				st.render(w, r, signup)
			} else {
				st.serverError(w, err)
			}
			return
		}
		//RenewToken is used for security purpose for each state change.
		st.sessionManager.RenewToken(r.Context())
		st.sessionManager.Put(r.Context(), "flash", "Your signup was successful, pleaselogin")
		http.Redirect(w, r, "/login", http.StatusSeeOther)

	default:
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte(http.StatusText(http.StatusNotImplemented)))
	}
}
func (st *sT) logoutHandler(w http.ResponseWriter, r *http.Request) {
	//RenewToken is used for security purpose for each state change.
	st.sessionManager.RenewToken(r.Context())
	st.sessionManager.Remove(r.Context(), authenticatedUserID)
	http.Redirect(w, r, "/home", http.StatusSeeOther)
}

//for chatValue and matHandler, the work is done in thier Ajax handlers
//below wch are playHandler (for chatHandler) and playMatHandler for matHandler
func (st *sT) chatHandler(w http.ResponseWriter, r *http.Request) {
	st.render(w, r, chat)
}

func (st *sT) matHandler(w http.ResponseWriter, r *http.Request) {
	st.render(w, r, mat)
}

//This is the Ajax end point for the chat.
func (st *sT) playHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm() //parse request, handle error
	if err != nil {
		st.errorLog.Println(err)
	}
	value, ok := r.Form["value"]
	if ok {
		//it sends the input to the chat application over the nats connection
		//and synchroously waits for the response to be delivered to its mailbox.
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

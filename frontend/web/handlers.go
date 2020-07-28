package main

import (
	"errors"
	"log"
	"net/http"

	//broker pkg contains the code that is used on both sides of the nats connectoin.
	"github.com/saied74/toychat/pkg/broker"
	"github.com/saied74/toychat/pkg/centerr"
	"github.com/saied74/toychat/pkg/forms"
	"golang.org/x/crypto/bcrypt"
)

// TODO: these handlers might be weak with respect to nats transport error

//=============================== Home ======================================

//much of the commmon work is done in render, addDefaultData and middlewares
func (st *sT) homeHandler(w http.ResponseWriter, r *http.Request) {
	st.render(w, r, home)
}

//=============================== Login ======================================
//using the golang standard library, so need to check for correct path and method.
func (st *sT) loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/login" {
		http.NotFound(w, r)
		return
	}
	switch r.Method {

	case "GET":
		st.render(w, r, login)

	case "POST":
		log.Printf("got to post")
		err := r.ParseForm()
		if err != nil {
			st.clientError(w, http.StatusBadRequest, err)
			return
		}
		Form := forms.NewForm(r.PostForm)
		st.initTD()
		//authenticateUserR R stands for remote sends the data to the dbmgr over
		//the nats connectoin to be validated.
		person, err := broker.AuthenticateEUR("users", Form.GetField("email"))
		log.Printf("AuthEUR: %v", person)
		if err != nil {
			if errors.Is(err, broker.ErrNoRecord) {
				st.td.Form.Errors.AddError("generic", "Email or Password is incorrect")
				st.render(w, r, login)
			} else {
				st.serverError(w, err)
			}
			return
		}
		hashedPassword := person.HashedPassword
		if len(hashedPassword) != 60 {
			st.td.Form.Errors.AddError("generic", "No such a record was found")
			st.render(w, r, login)
			return
		}
		err = bcrypt.CompareHashAndPassword([]byte(hashedPassword),
			[]byte(Form.GetField("password")))
		if err != nil {
			if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
				st.td.Form.Errors.AddError("generic", "Email or Password is incorrect")
				st.render(w, r, login)
			} else {
				st.serverError(w, err)
			}
			return
		}
		// st.td.Form = Form
		//RenewToken is used for security purpose for each state change.
		st.sessionManager.RenewToken(r.Context())
		st.sessionManager.Put(r.Context(), authenticatedUserID, person.ID)
		http.Redirect(w, r, home, http.StatusSeeOther)

	default:
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte(http.StatusText(http.StatusNotImplemented)))
	}
}

//=============================== Sign up =====================================

//The same here as login with path and method checking.
func (st *sT) signupHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/signup" {
		centerr.ErrorLog.Printf("bad path %s", r.URL.Path)
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
		Form := forms.NewForm(r.PostForm)
		Form.FieldRequired("name", "email", "password")
		Form.MaxLength("name", 255)
		Form.MaxLength("email", 255)
		Form.MatchPattern("email", forms.EmailRX)
		Form.MinLength("password", 10)

		if !Form.Valid() {
			st.render(w, r, signup)
			return
		}
		//once the form is validated (above), it is sent to the dbmgr over nats
		//to be inserted into the database.
		password := Form.GetField("password")
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
		if err != nil {
			st.serverError(w, err)
			return //note we are not returning any words so we can check for the error
		}
		err = broker.InsertEUR("users", Form.GetField("name"),
			Form.GetField("email"), string(hashedPassword)) //Form.GetField("password"))
		if err != nil {
			centerr.ErrorLog.Printf("Fatal Error %v", err)
			if errors.Is(err, broker.ErrDuplicateEmail) {
				Form.Errors.AddError("email", "Address is already in use")
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

//============================== Log Out ======================================

func (st *sT) logoutHandler(w http.ResponseWriter, r *http.Request) {
	//RenewToken is used for security purpose for each state change.
	st.sessionManager.RenewToken(r.Context())
	st.sessionManager.Remove(r.Context(), authenticatedUserID)
	http.Redirect(w, r, "/home", http.StatusSeeOther)
}

//=============================== Chat ======================================

//for chatValue and matHandler, the work is done in thier Ajax handlers
//below wch are playHandler (for chatHandler) and playMatHandler for matHandler
func (st *sT) chatHandler(w http.ResponseWriter, r *http.Request) {
	st.render(w, r, chat)
}

//=============================== Mat ======================================

func (st *sT) matHandler(w http.ResponseWriter, r *http.Request) {
	st.render(w, r, mat)
}

//============================= Play (chat) ===================================

//This is the Ajax end point for the chat.
func (st *sT) playHandler(w http.ResponseWriter, r *http.Request) {
	var dialogID, agentID int
	var msg string
	err := r.ParseForm() //parse request, handle error
	if err != nil {
		centerr.ErrorLog.Println(err)
	}
	message, ok := r.Form["value"]
	if ok {
		msg = message[0]

		//<------------------ Get User ID --------------------------->
		id := st.sessionManager.GetInt(r.Context(), authenticatedUserID)

		//<----------------- Get Dialog Record ----------------------->
		dialog, err := broker.GetDialog("dialogs", id) //check to see if ongoing dialog
		if err != nil {

			//<---------- If no dialog record, get agent ---------------->
			if errors.Is(err, broker.ErrNoRecord) { //&& dialog.AgentID != 0 {
				agentID, err2 := broker.SelectAgent() // TODO: no record found is not cared for
				if err2 != nil {
					st.serverError(w, err2)
					return
				}
				// <------------ with agentID and user ID, make dialog ----------->
				err2 = broker.MakeDialog("dialogs", id, agentID)
				if err2 != nil {
					st.serverError(w, err2)
					return
				}
				dialog, err := broker.GetDialog("dialogs", id)
				dialogID = dialog.DialogID
				if err != nil {
					st.serverError(w, err)
				}
			} else {
				st.serverError(w, err)
				return
			}
		} else {
			dialogID = dialog.DialogID
			agentID = dialog.AgentID
		}
		err = broker.EnterMsg("messages", dialogID, msg)
		if err != nil {
			st.serverError(w, err)
		}
		reply, err := broker.MessageAgent(agentID, id, msg)
		if err != nil {
			st.serverError(w, err)
		}
		w.Write([]byte(reply))
		return
	}
	return
}

//============================= Play (mat) ====================================

func (st *sT) playMatHandler(w http.ResponseWriter, r *http.Request) {
	// var matValue = ""
	err := r.ParseForm() //parse request, handle error
	if err != nil {
		centerr.ErrorLog.Println(err)
	}
	value, ok := r.Form["value"]
	if ok {
		matValue := st.chatConnection(value[0], "forMat", "fromMat")
		w.Write(matValue)
	}
}

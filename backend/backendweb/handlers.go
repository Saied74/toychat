package main

import (
	"encoding/gob"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	//broker pkg contains the code that is used on both sides of the nats connectoin.
	"github.com/saied74/toychat/pkg/broker"
	"github.com/saied74/toychat/pkg/centerr"
	"github.com/saied74/toychat/pkg/forms"
	"golang.org/x/crypto/bcrypt"
)

// TODO: these handlers might be weak with respect to nats transport error

//much of the commmon work is done in render, addDefaultData and middlewares

//============================ Home ================================
func (app *App) homeHandler(w http.ResponseWriter, r *http.Request) {
	err := app.pickPath(w, r)
	if err != nil {
		centerr.ErrorLog.Printf("bad path %s", r.URL.Path)
		http.NotFound(w, r)
		return
	}
	app.render(w, r, home)
}

//============================ Login ================================
func (app *App) loginHandler(w http.ResponseWriter, r *http.Request) {
	err := app.pickPath(w, r)
	if err != nil {
		centerr.ErrorLog.Printf("bad path %s", r.URL.Path)
		http.NotFound(w, r)
		return
	}
	switch r.Method {
	case GET:
		app.render(w, r, login)
		return
	case POST:
		gob.Register(broker.TableRow{})
		err := r.ParseForm()
		if err != nil {
			app.clientError(w, http.StatusBadRequest, err)
		}
		Form := forms.NewForm(r.PostForm)
		person, err := broker.AuthenticateXR(app.table, app.role,
			Form.GetField("email"))
		if err != nil {
			if errors.Is(err, broker.ErrNoRecord) {
				app.td.Form.Errors.AddError("generic", "No such a record was found")
				app.render(w, r, login)
			} else {
				app.serverError(w, err)
			}
			return
		}
		hashedPassword := person.HashedPassword
		if len(hashedPassword) != 60 {
			app.td.Form.Errors.AddError("generic", "No such a record was found")
			app.render(w, r, login)
			return
		}
		err = bcrypt.CompareHashAndPassword([]byte(hashedPassword),
			[]byte(Form.GetField("password")))
		if err != nil {
			if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
				app.td.Form.Errors.AddError("generic", "Email or Password is incorrect")
				app.render(w, r, login)
			} else {
				app.serverError(w, err)
			}
			return
		}
		app.sessionManager.RenewToken(r.Context())
		app.sessionManager.Put(r.Context(), authenticatedUserID, person.ID)
		http.Redirect(w, r, home, http.StatusSeeOther)

	default:
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte(http.StatusText(http.StatusNotImplemented)))
	}
}

//============================ Logout ================================
func (app *App) logoutHandler(w http.ResponseWriter, r *http.Request) {
	err := app.pickPath(w, r)
	if err != nil {
		centerr.ErrorLog.Printf("bad path %s", r.URL.Path)
		http.NotFound(w, r)
		return
	}
	//RenewToken is used for security purpose for each state change.
	app.sessionManager.RenewToken(r.Context())
	app.sessionManager.Remove(r.Context(), authenticatedUserID)
	http.Redirect(w, r, app.redirect, http.StatusSeeOther)
}

//======================== Add (Admin or Agent) ===============================
func (app *App) addHandler(w http.ResponseWriter, r *http.Request) {
	err := app.pickPath(w, r)
	if err != nil {
		centerr.ErrorLog.Printf("bad path %s", r.URL.Path)
		http.NotFound(w, r)
		return
	}
	switch r.Method {
	case GET:
		app.render(w, r, signup)
	case POST:
		err := r.ParseForm()
		if err != nil {
			app.clientError(w, http.StatusBadRequest, err)
		}
		app.td.Form = forms.NewForm(r.PostForm)
		app.td.Form.FieldRequired("name", "email", "password")
		app.td.Form.MaxLength("name", 256)
		app.td.Form.MaxLength("email", 256)
		app.td.Form.MatchPattern("email", forms.EmailRX)
		app.td.Form.MinLength("password", 10)
		if !app.td.Form.Valid() {
			app.render(w, r, signup)
			return
		}
		password := app.td.Form.GetField("password")
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
		if err != nil {
			app.serverError(w, err)
			return //note we are not returning any words so we can check for the error
		}
		err = broker.InsertXR(app.table, app.nextRole, app.td.Form.GetField("name"),
			app.td.Form.GetField("email"), string(hashedPassword))
		if err != nil {
			if errors.Is(err, broker.ErrDuplicateEmail) {
				app.td.Form.Errors.AddError("email", "Address is already in use")
				app.render(w, r, signup)
			} else {
				app.serverError(w, err)
			}
			return
		}
		app.sessionManager.RenewToken(r.Context())
		app.sessionManager.Put(r.Context(), "flash", "Your signup was successful, pleaselogin")
		http.Redirect(w, r, app.redirect, http.StatusSeeOther)

	default:
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte(http.StatusText(http.StatusNotImplemented)))
	}
}

//======================= Activation (admin or agent) ==========================
func (app *App) activationHandler(w http.ResponseWriter, r *http.Request) {
	err := app.pickPath(w, r)
	if err != nil {
		centerr.ErrorLog.Printf("bad path %s", r.URL.Path)
		http.NotFound(w, r)
		return
	}
	switch r.Method {
	case GET:
		//gwt admins from the admins table with active status as false
		people, err := broker.GetByStatusR("admins", app.nextRole, !app.td.Active)
		centerr.InfoLog.Printf("in activation get err %v", err)
		if err != nil {
			if errors.Is(err, broker.ErrNoRecord) {
				app.td.Table = &broker.TableRows{}
			} else {
				app.serverError(w, err)
				return
			}
		}
		// persons := *people
		if len(people) == 1 { // len(*people) == 1 {
			// person := *people
			if len(people[0].HashedPassword) != 60 {
				app.td.Table = &broker.TableRows{}
			} else {
				app.td.setPeople(&people)
			}
		} else {
			app.td.setPeople(&people)
			// app.td.setTable(people)
		}
		app.render(w, r, table)

	case POST:
		err := r.ParseForm()
		if err != nil {
			app.clientError(w, http.StatusBadRequest, err)
		}
		people, err := broker.GetByStatusR("admins", app.nextRole, !app.td.Active)
		if err != nil {
			centerr.ErrorLog.Printf("Fatal Error %v", err)
			app.serverError(w, err)
		}
		newPeople := broker.TableRows{}
		for i, person := range people { //Short because that is how the api responds
			candidate := "stateCheck" + strconv.Itoa(i)
			for key := range r.Form {
				if key == candidate {
					person.Active = app.td.Active
					centerr.InfoLog.Println("person.Active", person.Active)
					newPeople = append(newPeople, person)
				}
			}
		}
		err = broker.ActivationR("admins", app.nextRole, &newPeople)
		if err != nil {
			centerr.InfoLog.Printf("Fatal Error %v", err)
			app.serverError(w, err)
		}
		// centerr.ErrorLog.Printf("Activation: %v", newPeople)
		app.sessionManager.RenewToken(r.Context())
		http.Redirect(w, r, app.td.Home, http.StatusSeeOther)
	default:
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte(http.StatusText(http.StatusNotImplemented)))
	}
}

//====================== Change Password (Admin or Agent) ======================
func (app *App) changePasswordHandler(w http.ResponseWriter, r *http.Request) {
	err := app.pickPath(w, r)
	if err != nil {
		centerr.ErrorLog.Printf("bad path %s", r.URL.Path)
		http.NotFound(w, r)
		return
	}
	switch r.Method {
	case "GET":
		app.render(w, r, "chgPwd")
		return
	case "POST":
		err := r.ParseForm()
		if err != nil {
			app.clientError(w, http.StatusBadRequest, err)
		}
		app.td.Form = forms.NewForm(r.PostForm)
		app.td.Form.FieldRequired("email", "passwordOld", "passwordNew")
		app.td.Form.MaxLength("email", 256)
		app.td.Form.MatchPattern("email", forms.EmailRX)
		app.td.Form.MinLength("passwordNew", 10)
		app.td.Form.MinLength("passwordOld", 10)
		if !app.td.Form.Valid() {
			app.render(w, r, "chgPwd")
			return
		}
		email := app.td.Form.GetField("email")
		pwd := app.td.Form.GetField("passwordOld")
		centerr.ErrorLog.Printf("table: %s, role: %s, email: %s", app.table, app.role, email)
		person, err := broker.AuthenticateXR(app.table, app.role, email)
		if err != nil {
			app.serverError(w, err)
			return
		}
		hashedPassword := person.HashedPassword
		centerr.InfoLog.Printf("hashed password: %s", hashedPassword)
		if len(hashedPassword) != 60 {
			app.td.Form.Errors.AddError("generic", "No such a record was found")
			app.render(w, r, "chgPwd")
			return
		}
		err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(pwd))
		if err != nil {
			centerr.ErrorLog.Printf("Bcrypt err: %v", err)
			if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
				app.td.Form.Errors.AddError("generic", "Email or Password is incorrect")
				centerr.ErrorLog.Printf("Error from bcypt match 1 %v", err)
				app.render(w, r, "chgPwd")

			} else {
				centerr.ErrorLog.Printf("Error from bcypt match 2 %v", err)
				app.serverError(w, err)

			}
			return
		}
		password := app.td.Form.GetField("passwordNew")
		hashedNewPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
		if err != nil {
			app.serverError(w, err)
			return //note we are not returning any words so we can check for the error
		}
		//once the form is validated (above), it is sent to the dbmgr over nats
		//to be inserted into the database.
		err = broker.ChgPwdR(app.table, app.role, email, string(hashedNewPassword))
		if err != nil {
			centerr.InfoLog.Printf("error from change pwd: %v", err)
			app.td.Form.Errors.AddError("generic", "Change password fail, try again")
			app.render(w, r, home)
			return
		}
		//RenewToken is used for security purpose for each state change.
		app.sessionManager.RenewToken(r.Context())
		app.sessionManager.Put(r.Context(), "flash", "Your password was changed, pleaselogin")
		app.td.Msg = "You changed your password, please re-login"
		app.render(w, r, login)

		// http.Redirect(w, r, app.redirect, http.StatusSeeOther)
	default:
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte(http.StatusText(http.StatusNotImplemented)))
	}
}

//============================== Agent online ================================
func (app *App) agentOnlineHandler(w http.ResponseWriter, r *http.Request) {
	err := app.pickPath(w, r)
	if err != nil {
		centerr.ErrorLog.Printf("bad path %s", r.URL.Path)
		http.NotFound(w, r)
		return
	}
	switch r.Method {
	case GET:
		app.td.Online = true
		id := app.sessionManager.GetInt(r.Context(), authenticatedUserID)
		if id == 0 {
			app.serverError(w, fmt.Errorf("no session id"))
		}
		err := broker.PutLine(app.table, app.role, id, true)
		if err != nil {
			app.serverError(w, err)
		}
		app.render(w, r, chat)
		return
	default:
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte(http.StatusText(http.StatusNotImplemented)))
	}
}

//============================== Agent onffline ================================
func (app *App) agentOfflineHandler(w http.ResponseWriter, r *http.Request) {
	err := app.pickPath(w, r)
	if err != nil {
		centerr.ErrorLog.Printf("bad path %s", r.URL.Path)
		http.NotFound(w, r)
		return
	}
	switch r.Method {
	case GET:
		app.td.Online = false
		id := app.sessionManager.GetInt(r.Context(), authenticatedUserID)
		if id == 0 {
			app.serverError(w, fmt.Errorf("no session id"))
		}
		err := broker.PutLine(app.table, app.role, id, false)
		if err != nil {
			app.serverError(w, err)
		}
		app.render(w, r, home)
		return
	default:
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte(http.StatusText(http.StatusNotImplemented)))
	}
}

//============================== Agent onffline ================================
func (app *App) agentChatHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm() //parse request, handle error
	if err != nil {
		centerr.ErrorLog.Println(err)
	}
	value, ok := r.Form["action"]
	if !ok {
		centerr.ErrorLog.Println(err)
	}
	centerr.InfoLog.Printf("got answer from the browser %s", value)

}

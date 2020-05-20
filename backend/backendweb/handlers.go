package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	//broker pkg contains the code that is used on both sides of the nats connectoin.
	"github.com/saied74/toychat/pkg/broker"
	"github.com/saied74/toychat/pkg/centerr"
	"github.com/saied74/toychat/pkg/forms"
	"github.com/saied74/toychat/pkg/models"
	"golang.org/x/crypto/bcrypt"
)

// TODO: these handlers might be weak with respect to nats transport error

//much of the commmon work is done in render, addDefaultData and middlewares
func (app *App) homeHandler(w http.ResponseWriter, r *http.Request) {
	err := app.pickPath(w, r)
	if err != nil {
		app.errorLog.Printf("bad path %s", r.URL.Path)
		http.NotFound(w, r)
		return
	}
	app.render(w, r, home)
}

func (app *App) loginHandler(w http.ResponseWriter, r *http.Request) {
	err := app.pickPath(w, r)
	if err != nil {
		app.errorLog.Printf("bad path %s", r.URL.Path)
		http.NotFound(w, r)
		return
	}
	switch r.Method {
	case GET:
		app.render(w, r, login)
		return
	case POST:
		err := r.ParseForm()
		if err != nil {
			app.clientError(w, http.StatusBadRequest, err)
		}
		Form := forms.NewForm(r.PostForm)
		person, err := models.AuthenticateUserR(app.table, app.role,
			Form.GetField("email"))
		if err != nil {
			app.serverError(w, err)
			return
		}
		app.infoLog.Printf("Person back in handler %v", person)
		hashedPassword := person.HashedPassword
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

func (app *App) logoutHandler(w http.ResponseWriter, r *http.Request) {
	err := app.pickPath(w, r)
	if err != nil {
		app.errorLog.Printf("bad path %s", r.URL.Path)
		http.NotFound(w, r)
		return
	}
	//RenewToken is used for security purpose for each state change.
	app.sessionManager.RenewToken(r.Context())
	app.sessionManager.Remove(r.Context(), authenticatedUserID)
	http.Redirect(w, r, app.redirect, http.StatusSeeOther)
}

func (app *App) addHandler(w http.ResponseWriter, r *http.Request) {
	err := app.pickPath(w, r)
	if err != nil {
		app.errorLog.Printf("bad path %s", r.URL.Path)
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
		Form := forms.NewForm(r.PostForm)
		Form.FieldRequired("name", "email", "password")
		Form.MaxLength("name", 255)
		Form.MaxLength("email", 255)
		Form.MatchPattern("email", forms.EmailRX)
		Form.MinLength("password", 10)
		if !Form.Valid() {
			app.render(w, r, signup)
			return
		}
		password := Form.GetField("password")
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
		if err != nil {
			app.serverError(w, err)
			return //note we are not returning any words so we can check for the error
		}

		err = models.InsertAdminR(app.table, app.nextRole, Form.GetField("name"),
			Form.GetField("email"), string(hashedPassword))
		if err != nil {
			centerr.ErrorLog.Printf("Fatal Error %v", err)
			if errors.Is(err, broker.ErrDuplicateEmail) {
				Form.Errors.AddError("email", "Address is already in use")
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

func (app *App) activationHandler(w http.ResponseWriter, r *http.Request) {
	err := app.pickPath(w, r)
	if err != nil {
		app.errorLog.Printf("bad path %s", r.URL.Path)
		http.NotFound(w, r)
		return
	}
	switch r.Method {
	case GET:
		//gwt admins from the admins table with active status as false
		people, err := models.GetByStatusR("admins", app.nextRole, !app.td.Active)
		if err != nil {
			centerr.ErrorLog.Printf("Fatal Error %v", err)
			app.serverError(w, err)
		}
		app.td.Table = people
		app.render(w, r, table)
	case POST:
		err := r.ParseForm()
		if err != nil {
			app.clientError(w, http.StatusBadRequest, err)
		}
		people, err := models.GetByStatusR("admins", app.nextRole, !app.td.Active)
		if err != nil {
			centerr.ErrorLog.Printf("Fatal Error %v", err)
			app.serverError(w, err)
		}
		newPeople := []broker.Person{}
		for i, person := range *people { //Short because that is how the api responds
			candidate := "stateCheck" + strconv.Itoa(i)
			for key := range r.Form {
				if key == candidate {
					person.Active = app.td.Active
					centerr.InfoLog.Println("person.Active", person.Active)
					newPeople = append(newPeople, person)
				}
			}
		}
		err = models.ActivationR("admins", app.nextRole, &newPeople)
		if err != nil {
			centerr.ErrorLog.Printf("Fatal Error %v", err)
			app.serverError(w, err)
		}
		app.sessionManager.RenewToken(r.Context())
		http.Redirect(w, r, app.td.Home, http.StatusSeeOther)
	default:
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte(http.StatusText(http.StatusNotImplemented)))
	}
}

func (app *App) changePasswordHandler(w http.ResponseWriter, r *http.Request) {
	err := app.pickPath(w, r)
	if err != nil {
		app.errorLog.Printf("bad path %s", r.URL.Path)
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
		centerr.InfoLog.Println("post form", r.PostForm)
		Form := forms.NewForm(r.PostForm)
		Form.FieldRequired("email", "passwordOld", "passwordNew")
		Form.MaxLength("email", 255)
		Form.MatchPattern("email", forms.EmailRX)
		Form.MinLength("passwordOld", 10)
		Form.MinLength("passwordNew", 10)
		if !Form.Valid() {
			app.render(w, r, signup)
			return
		}
		email := Form.GetField("email")
		pwd := Form.GetField("passwordOld")
		person, err := models.AuthenticateUserR(app.table, app.role, email)
		if err != nil {
			app.serverError(w, err)
			return
		}
		hashedPassword := person.HashedPassword
		err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(pwd))
		if err != nil {
			app.errorLog.Printf("Bcrypt err: %v", err)
			if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
				app.td.Form.Errors.AddError("generic", "Email or Password is incorrect")
				app.render(w, r, login)
			} else {
				app.serverError(w, err)
			}
			return
		}
		password := Form.GetField("passwordNew")
		hashedNewPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
		if err != nil {
			app.serverError(w, err)
			return //note we are not returning any words so we can check for the error
		}
		//once the form is validated (above), it is sent to the dbmgr over nats
		//to be inserted into the database.
		err = models.ChgPwdR(app.table, app.role, email, string(hashedNewPassword))
		if err != nil {
			app.infoLog.Printf("error from change pwd: %v", err)
			app.render(w, r, "chgPwd")
			// app.serverError(w, err)

			return
		}
		//RenewToken is used for security purpose for each state change.
		app.sessionManager.RenewToken(r.Context())
		app.sessionManager.Put(r.Context(), "flash", "Your password was changed, pleaselogin")
		http.Redirect(w, r, app.redirect, http.StatusSeeOther)
	default:
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte(http.StatusText(http.StatusNotImplemented)))
	}
}

func (app *App) agentOnlineHandler(w http.ResponseWriter, r *http.Request) {
	err := app.pickPath(w, r)
	if err != nil {
		app.errorLog.Printf("bad path %s", r.URL.Path)
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
		err := models.PutLine(app.table, app.role, id, true)
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

func (app *App) agentOfflineHandler(w http.ResponseWriter, r *http.Request) {
	err := app.pickPath(w, r)
	if err != nil {
		app.errorLog.Printf("bad path %s", r.URL.Path)
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
		err := models.PutLine(app.table, app.role, id, false)
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

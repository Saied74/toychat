package main

import (
	"errors"
	"net/http"
	"strconv"

	//broker pkg contains the code that is used on both sides of the nats connectoin.
	"github.com/saied74/toychat/pkg/broker"
	"github.com/saied74/toychat/pkg/centerr"
	"github.com/saied74/toychat/pkg/forms"
	"github.com/saied74/toychat/pkg/models"
)

// TODO: these handlers might be weak with respect to nats transport error

//much of the commmon work is done in render, addDefaultData and middlewares
func (app *App) homeHandler(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case superHome:
		app.buildSuper()
	case adminHome:
		app.buildAdmin()
	case agentHome:
		app.buildAgent()
	default:
		app.errorLog.Printf("bad path %s", r.URL.Path)
		http.NotFound(w, r)
		return
	}
	app.render(w, r, home)
}

func (app *App) loginHandler(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case superLogin:
		app.buildSuper()
	case adminLogin:
		app.buildAdmin()
	case agentLogin:
		app.buildAgent()
	default:
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
		app.td.Form = forms.NewForm(r.PostForm)

		//authenticateUserR R stands for remote sends the data to the dbmgr over
		//the nats connectoin to be validated.
		table := app.td.table
		role := app.td.role
		email := app.td.Form.GetField("email")
		pwd := app.td.Form.GetField("password")
		id, err := models.AuthenticateUserR(table, role, email, pwd)
		// app.td.table,
		// app.td.Form.GetField("email"), app.td.Form.GetField("password"))
		if err != nil {
			if errors.Is(err, broker.ErrInvalidCredentials) {
				app.td.Form.Errors.AddError("generic", "Email or Password is incorrect")
				app.render(w, r, login)
			} else {
				app.serverError(w, err)
			}
			return
		}
		//RenewToken is used for security purpose for each state change.
		app.sessionManager.RenewToken(r.Context())
		app.sessionManager.Put(r.Context(), authenticatedUserID, id)
		http.Redirect(w, r, home, http.StatusSeeOther)

	default:
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte(http.StatusText(http.StatusNotImplemented)))
	}
}

func (app *App) logoutHandler(w http.ResponseWriter, r *http.Request) {
	var redirect string
	switch r.URL.Path {
	case superLogout:
		app.buildSuper()
		redirect = superHome
	case adminLogout:
		app.buildAdmin()
		redirect = adminHome
	case agentLogout:
		app.buildAgent()
		redirect = agentHome
	default:
		app.errorLog.Printf("bad path %s", r.URL.Path)
		http.NotFound(w, r)
		return
	}
	//RenewToken is used for security purpose for each state change.
	app.sessionManager.RenewToken(r.Context())
	app.sessionManager.Remove(r.Context(), authenticatedUserID)
	http.Redirect(w, r, redirect, http.StatusSeeOther)
}

func (app *App) addHandler(w http.ResponseWriter, r *http.Request) {
	var redirect string
	switch r.URL.Path {
	case "/super/addAdmin":
		app.buildSuper() //only super users can add admins
		redirect = superHome
	case "/admin/addAgent":
		app.buildAdmin()
		redirect = adminHome
	default:
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
		//once the form is validated (above), it is sent to the dbmgr over nats
		//to be inserted into the database.
		err = models.InsertAdminR(app.td.table, app.td.nextRole, Form.GetField("name"),
			Form.GetField("email"), Form.GetField("password"))
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
		//RenewToken is used for security purpose for each state change.
		app.sessionManager.RenewToken(r.Context())
		app.sessionManager.Put(r.Context(), "flash", "Your signup was successful, pleaselogin")
		http.Redirect(w, r, redirect, http.StatusSeeOther)

	default:
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte(http.StatusText(http.StatusNotImplemented)))
	}
}

func (app *App) activationHandler(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case activateAdmin:
		app.buildSuper()
		app.td.Active = true
	case deactivateAdmin:
		app.buildSuper()
		app.td.Active = false
	case activateAgent:
		app.buildAdmin()
		app.td.Active = true
	case deactivateAgent:
		app.buildAdmin()
		app.td.Active = false
	default:
		app.errorLog.Printf("bad path %s", r.URL.Path)
		http.NotFound(w, r)
		return
	}
	switch r.Method {
	case GET:
		//gwt admins from the admins table with active status as false
		people, err := models.GetByStatusR("admins", app.td.nextRole, !app.td.Active)
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
		people, err := models.GetByStatusR("admins", app.td.nextRole, !app.td.Active)
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
					newPeople = append(newPeople, person)
				}
			}
		}
		err = models.ActivationR("admins", app.td.nextRole, &newPeople)
		if err != nil {
			centerr.ErrorLog.Printf("Fatal Error %v", err)
			app.serverError(w, err)
		}
		http.Redirect(w, r, app.td.Home, http.StatusSeeOther)
	default:
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte(http.StatusText(http.StatusNotImplemented)))
	}

}

func (app *App) adminChangePasswordHandler(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, home)
}

func (app *App) adminAddAgentHandler(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, home)
}

func (app *App) agentChangePasswordHandler(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, home)
}

func (app *App) agentOnlineHandler(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, home)
}

func (app *App) agentOfflineHandler(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, home)
}

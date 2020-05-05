package main

import (
	"errors"
	"net/http"
)

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
		id, err := st.users.authenticateUser(st.td.Form.getField("email"),
			st.td.Form.getField("password"))
		if err != nil {
			if errors.Is(err, errInvalidCredentials) {
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
		st.infoLog.Printf("\nname %s \nemail %s \npassword %s \n",
			Form.getField("name"), Form.getField("email"), Form.getField("password"))

		if !Form.valid() {
			st.render(w, r, signup)
			return
		}
		err = st.users.insertUser(Form.getField("name"), Form.getField("email"),
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
		chatValue := st.chatConnection(value[0])
		// sliceValue := strings.Split(value[0], " ")
		// for i, j := 0, len(sliceValue)-1; i < j; i, j = i+1, j-1 {
		// 	sliceValue[i], sliceValue[j] = sliceValue[j], sliceValue[i]
		// }
		// newvalue := strings.Join(sliceValue, " ")
		// w.Write([]byte(newvalue))
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
		matValue := st.matConnection(value[0])
		// sliceLen := len(strings.Split(value[0], " "))
		// for i := 0; i < sliceLen; i++ {
		// matValue += "mat "
		// }
		// w.Write([]byte(matValue))
		w.Write(matValue)
	}
}

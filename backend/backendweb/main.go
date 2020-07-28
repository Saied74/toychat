package main

import (
	"crypto/tls"
	"database/sql"
	"flag"
	"html/template"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/v2"
	_ "github.com/go-sql-driver/mysql"
	"github.com/saied74/toychat/pkg/broker"
	"github.com/saied74/toychat/pkg/centerr"
	"github.com/saied74/toychat/pkg/forms"
)

//so if the string is used in new packages, it remains privat for this app.
type contextKey string

const contextKeyIsAuthenticated = contextKey("isAuthenticated")

//UserModel wraps the sql.DB connections
type UserModel struct {
	DB *sql.DB
}

//App struct is for injecting data into the handlers and supporting methods.
type App struct {
	cache          map[string]*template.Template
	sessionManager *scs.SessionManager
	users          *UserModel
	td             *templateData
	table          string
	role           string
	nextRole       string
	redirect       string
}

type templateData struct {
	Scope     string //scope pharase on the navbar
	Home      string //home address link (e.g. /super/home or /admin/home)
	Login     string //login link (e.g. /super/login or /admin/login)
	Logout    string //same with logout.
	ChgPwd    string
	Msg       string         //login, add admin or add agent message.
	SideLink1 string         //addAgent or addAdmin
	SideLink2 string         //activateAgent or activateAdmin
	SideLink3 string         //deactivateAgent or deactivateAdmin
	Super     bool           //role super = true
	Admin     bool           //role admin = true
	Agent     bool           // role agent= true
	Active    bool           //active or not
	Online    bool           //Agent online or offline
	Table     *broker.People //[]broker.Person
	Form      *forms.FormData
	UserName  string
	LoggedIn  bool
	Flash     string
	CSRFToken string
}

func (t *templateData) Length() int {
	return 1
}

//ReturnFirst is to handle indexing into TableProxy concrte type People
func (t *templateData) ReturnFirst() *broker.Person {
	return &broker.Person{}
}

func (t *templateData) setPeople(p *broker.People) {
	t.Table = p
}

func main() {
	var err error
	//the pw flag is mandatory.
	ipAddress := flag.String("ipa", ":8000", "server ip address")
	dsn := flag.String("dsn", "toy:password@/toychat?parseTime=true",
		"MySQL data source name")
	pw := flag.String("pw", "password", "database password is always required")
	flag.Parse()
	dbAddress := strings.Replace(*dsn, "password", *pw, 1)

	db, err := openDB(dbAddress)
	if err != nil {
		centerr.ErrorLog.Fatal(err)
	}
	defer db.Close()

	// var allTmplFiles tmDataer

	app := &App{
		sessionManager: scs.New(),
		users:          &UserModel{DB: db},
		td: &templateData{
			Form: &forms.FormData{
				Fields: url.Values{},
				Errors: forms.ErrOrs{},
			},
		},
		cache: newTemplateCache(allTmplFiles),
	}
	//at some point when different applicaitons are running on different servers
	//the database for each applicaiton needs to be seperated.
	app.sessionManager.Store = mysqlstore.New(db)
	app.sessionManager.Lifetime = 72 * time.Hour
	app.sessionManager.Cookie.Name = "sessionTwo"

	tlsConfig := &tls.Config{
		PreferServerCipherSuites: true,
		CurvePreferences:         []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	mux := app.routes()
	srv := &http.Server{
		Addr:         *ipAddress,
		ErrorLog:     centerr.ErrorLog,
		Handler:      app.dynamicRoutes(mux), //see the middlware file.
		TLSConfig:    tlsConfig,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	centerr.InfoLog.Printf("Starting server on %s", *ipAddress)
	err = srv.ListenAndServeTLS(serverCrt, serverKey)
	centerr.ErrorLog.Fatal(err)
}

// The openDB() function wraps sql.Open() and returns a sql.DB connection pool
// for a given DSN.
func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

//chat and play, in addition to the dynamicRoutes middlware are wrapped with
//requireAuthentication so they are not accessible to non authenicated users.
func (app *App) routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc(superHome, app.homeHandler)
	mux.HandleFunc(superLogin, app.loginHandler)
	mux.HandleFunc(superLogout, app.logoutHandler)
	mux.HandleFunc(addAdmin, app.requireAuthentication(app.addHandler))
	mux.HandleFunc(activateAdmin, app.requireAuthentication(app.activationHandler))
	mux.HandleFunc(deactivateAdmin, app.requireAuthentication(app.activationHandler))
	mux.HandleFunc(adminHome, app.homeHandler)
	mux.HandleFunc(adminLogin, app.loginHandler)
	mux.HandleFunc(adminLogout, app.logoutHandler)
	mux.HandleFunc(adminChgPwd, app.requireAuthentication(app.changePasswordHandler))
	mux.HandleFunc(addAgent, app.requireAuthentication(app.addHandler))
	mux.HandleFunc(activateAgent, app.requireAuthentication(app.activationHandler))
	mux.HandleFunc(deactivateAgent, app.requireAuthentication(app.activationHandler))
	mux.HandleFunc(agentHome, app.homeHandler)
	mux.HandleFunc(agentLogin, app.loginHandler)
	mux.HandleFunc(agentChgPwd, app.requireAuthentication(app.changePasswordHandler))
	mux.HandleFunc(agentLogout, app.logoutHandler)
	mux.HandleFunc(agentOnline, app.requireAuthentication(app.agentOnlineHandler))
	mux.HandleFunc(agentOffline, app.requireAuthentication(app.agentOfflineHandler))
	mux.HandleFunc("/agent/chat", app.requireAuthentication(app.agentChatHandler))
	return mux
}

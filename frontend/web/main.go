package main

import (
	"crypto/tls"
	"database/sql"
	"flag"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/v2"
	_ "github.com/go-sql-driver/mysql"
)

type contextKey string

const contextKeyIsAuthenticated = contextKey("isAuthenticated")

type sT struct {
	errorLog       *log.Logger
	infoLog        *log.Logger
	cache          map[string]*template.Template
	sessionManager *scs.SessionManager
	users          *userModel
	td             *templateData
}

type templateData struct {
	Form      *formData
	UserName  string
	LoggedIn  bool
	Flash     string
	CSRFToken string
}

func main() {
	var err error
	// var st sT
	ipAddress := flag.String("ipa", ":4000", "server ip address")
	dsn := flag.String("dsn", "toy:password@/toychat?parseTime=true",
		"MySQL data source name")
	pw := flag.String("pw", "password", "database password is always required")
	flag.Parse()
	dbAddress := strings.Replace(*dsn, "password", *pw, 1)

	infoLog := getInfoLogger(os.Stdout)
	errorLog := getErrorLogger(os.Stdout)

	db, err := openDB(dbAddress)
	if err != nil {
		errorLog().Fatal(err)
	}
	defer db.Close()

	st := &sT{
		infoLog:        infoLog(),
		errorLog:       errorLog(),
		sessionManager: scs.New(),
		users:          &userModel{dB: db},
		td: &templateData{
			Form: &formData{
				Fields: url.Values{},
				Errors: errOrs{},
			},
		},
		cache: newTemplateCache(allTmplFiles),
	}

	st.sessionManager.Store = mysqlstore.New(db)

	tlsConfig := &tls.Config{
		PreferServerCipherSuites: true,
		CurvePreferences:         []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	mux := st.routes()
	srv := &http.Server{
		Addr:         *ipAddress,
		ErrorLog:     st.errorLog,
		Handler:      st.dynamicRoutes(mux),
		TLSConfig:    tlsConfig,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	st.infoLog.Printf("Starting server on %s", *ipAddress)
	err = srv.ListenAndServeTLS(serverCrt, serverKey)
	st.errorLog.Fatal(err)
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

func (st *sT) routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/home", st.homeHandler)
	mux.Handle("/chat", st.requireAuthentication(http.HandlerFunc(st.chatHandler)))
	mux.Handle("/play", st.requireAuthentication(http.HandlerFunc(st.playHandler)))
	mux.HandleFunc("/playmat", st.playMatHandler)
	mux.HandleFunc("/mat", st.matHandler)
	mux.HandleFunc("/login", st.loginHandler)
	mux.HandleFunc("/logout", st.logoutHandler)
	mux.HandleFunc("/signup", st.signupHandler)
	return mux
}
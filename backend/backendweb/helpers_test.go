package main

import (
	"bytes"
	"context"
	"net/http/httptest"
	"strings"
	"testing"
)

var testSupertd = templateData{
	Scope:     "Super User",
	Home:      "/super/home",
	Login:     "/super/login",
	Logout:    "/super/logout",
	ChgPwd:    "",
	SideLink1: "/super/addAdmin",
	SideLink2: "/super/activateAdmin",
	SideLink3: "/super/deactivateAdmin",
	Super:     true,
	Admin:     false,
	Agent:     false,
	Msg:       "Please log in",
}
var testSuperapp = App{
	table:    "admins",
	role:     "superadmin",
	nextRole: "admin",
	redirect: "/super/home",
	td:       &testSupertd,
}

var testAdmintd = templateData{
	Scope:     "Admin User",
	Home:      "/admin/home",
	Login:     "/admin/login",
	Logout:    "/admin/logout",
	ChgPwd:    "/admin/changePassword",
	SideLink1: "/admin/addAgent",
	SideLink2: "/admin/activateAgent",
	SideLink3: "/admin/deactivateAgent",
	Super:     false,
	Admin:     true,
	Agent:     false,
	Msg:       "Please log in",
}
var testAdminapp = App{
	table:    "admins",
	role:     "admin",
	nextRole: "agent",
	redirect: "/admin/home",
	td:       &testAdmintd,
}

var testAgenttd = templateData{
	Scope:     "Agent",
	Home:      "/agent/home",
	Login:     "/agent/login",
	Logout:    "/agent/logout",
	ChgPwd:    "/agent/changePassword",
	SideLink1: "/agent/online",
	SideLink2: "/agent/offline",
	SideLink3: "",
	Super:     false,
	Admin:     false,
	Agent:     true,
	Msg:       "Please log in",
}
var testAgentapp = App{
	table:    "admins",
	role:     "agent",
	nextRole: "",
	redirect: "/agent/home",
	td:       &testAgenttd,
}

func (app *App) appCompare(testApp *App) bool {
	if app.table != testApp.table {
		return false
	}
	if app.role != testApp.role {
		return false
	}
	if app.nextRole != testApp.nextRole {
		return false
	}
	if app.redirect != testApp.redirect {
		return false
	}
	if app.td.Scope != testApp.td.Scope {
		return false
	}
	if app.td.Home != testApp.td.Home {
		return false
	}
	if app.td.Login != testApp.td.Login {
		return false
	}
	if app.td.Logout != testApp.td.Logout {
		return false
	}
	if app.td.ChgPwd != testApp.td.ChgPwd {
		return false
	}
	if app.td.SideLink1 != testApp.td.SideLink1 {
		return false
	}
	if app.td.SideLink2 != testApp.td.SideLink2 {
		return false
	}
	if app.td.SideLink3 != testApp.td.SideLink3 {
		return false
	}
	if app.td.Super != testApp.td.Super {
		return false
	}
	if app.td.Admin != testApp.td.Admin {
		return false
	}
	if app.td.Agent != testApp.td.Agent {
		return false
	}
	// if app.td.Msg != testApp.td.Msg {
	// 	return false
	// }
	return true
}

func TestBuildSuper(t *testing.T) {
	var app = App{
		td: &templateData{},
	}
	app.buildSuper()
	if !app.appCompare(&testSuperapp) {
		t.Errorf("expected: %v\ngot: %v\n", testSuperapp, app)
	}
	if app.appCompare(&testAdminapp) {
		t.Errorf("\nexp: %v\ngot: %v\n", testAdminapp, app)
	}
}

func TestBuildAdmin(t *testing.T) {
	var app = App{
		td: &templateData{},
	}
	app.buildAdmin()
	if !app.appCompare(&testAdminapp) {
		t.Errorf("\nexp app: %v\ngot app: %v\nexp td: %v\ngot td: %v\n",
			testAdminapp, app, testAdmintd, *app.td)
	}
	if app.appCompare(&testAgentapp) {
		t.Errorf("\nexp: %v\ngot: %v\n", testAgentapp, app)
	}
}

func TestBuildAgent(t *testing.T) {
	var app = App{
		td: &templateData{},
	}
	app.buildAgent()
	if !app.appCompare(&testAgentapp) {
		t.Errorf("\nexp app: %v\ngot app: %v\nexp td: %v\ngot td: %v\n",
			testAgentapp, app, testAgenttd, *app.td)
	}
	if app.appCompare(&testSuperapp) {
		t.Errorf("\nexp: %v\ngot: %v\n", testSuperapp, app)
	}
}

func TestPickPath(t *testing.T) {
	var app = App{
		td: &templateData{},
	}
	var err error
	var urlList = []string{"/super/home", "/super/login", "/super/logout",
		"/super/addAdmin", "/super/activateAdmin", "/super/deactivateAdmin",
		"/admin/home", "/admin/login", "/admin/logout", "/admin/changePassword",
		"/admin/addAgent", "/admin/activateAgent", "/admin/deactivateAgent",
		"/agent/home", "/agent/login", "/agent/logout", "/agent/changePassword",
		"/agent/online", "/agent/offline"}

	w := httptest.NewRecorder()

	for _, urlItem := range urlList {
		item := strings.Split(urlItem, "/")
		switch item[1] {
		case "super":
			r := httptest.NewRequest("GET", urlItem, nil)
			err = app.pickPath(w, r)
			if err != nil {
				t.Errorf("Error %v processing %s,", err, urlItem)
			}
			if !app.appCompare(&testSuperapp) {
				t.Errorf("\nexp: %v\ngot: %v\nexp: %v\ngot: %v\n",
					testSuperapp, app, testSuperapp.td, app.td)
			}
		case "admin":
			r := httptest.NewRequest("GET", urlItem, nil)
			err = app.pickPath(w, r)
			if err != nil {
				t.Errorf("Error %v processing %s,", err, urlItem)
			}
			if !app.appCompare(&testAdminapp) {
				t.Errorf("\nexp: %v\ngot: %v\nexp: %v\ngot: %v\n",
					testAdminapp, app, testAdminapp.td, app.td)
			}
		case "agent":
			r := httptest.NewRequest("GET", urlItem, nil)
			err = app.pickPath(w, r)
			if err != nil {
				t.Errorf("Error %v processing %s,", err, urlItem)
			}
			if !app.appCompare(&testAgentapp) {
				t.Errorf("\nexp: %v\ngot: %v\nexp: %v\ngot: %v\n",
					testAgentapp, app, testAgentapp.td, app.td)
			}
		}
	}
}

type tmTestType map[string][]string

var tm = tmTestType{
	"home": []string{`{{.Foo}}{{block "Moo" .}}{{end}}`,
		`{{define "Moo"}}{{.Bar}}{{end}}`},
	"away": []string{`{{.Bar}}{{block "Roo" .}}{{end}}`,
		`{{define "Roo"}}{{.Foo}}{{end}}`},
}

func (in tmTestType) tmpData() (*tmData, error) {
	out := tmData(in)
	return &out, nil
}

func TestNewTemplateCashe(t *testing.T) {
	type td struct {
		Foo string
		Bar string
	}
	exp := map[string]string{
		"home": "foobar",
		"away": "barfoo",
	}
	tmpl := newTemplateCache(tm)
	for key, value := range tmpl {
		buf := new(bytes.Buffer)
		value.Execute(buf, td{"foo", "bar"})
		var p = make([]byte, 6) //this is tricky, it has to match the length of the test
		n, err := buf.Read(p)
		if err != nil {
			t.Errorf("error in newTemplateCashe read: %d bytes got err %v", n, err)
		}
		if string(p) != exp[key] {
			t.Errorf("at %s expected %s, %d got %s, %d",
				key, exp[key], len(exp[key]), string(p), len(string(p)))
		}
	}
}

func TestIsAuthenticated(t *testing.T) {
	var isA bool
	app := App{}
	r := httptest.NewRequest("GET", "/super/home", nil)
	isA = app.isAuthenticated(r)
	if isA {
		t.Errorf("no context, expecting false, got %v", isA)
	}
	ctx := r.Context()
	ctx = context.WithValue(ctx, contextKeyIsAuthenticated, true)
	r = r.WithContext(ctx)
	isA = app.isAuthenticated(r)
	if !isA {
		t.Errorf("true context, expecting true, got %v", isA)
	}
	ctx = context.WithValue(ctx, contextKeyIsAuthenticated, false)
	r = r.WithContext(ctx)
	isA = app.isAuthenticated(r)
	if isA {
		t.Errorf("false context, expecting false, got %v", isA)
	}

}

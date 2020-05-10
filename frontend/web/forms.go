package main

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"
)

//methods on this page validate the html form fields and produce html error
//messages to be displayed as feedback to to he user.

//I have changed the errors from map[string][]string, the way that Alex Edwards
//has it to just map[string]string with semicolmn seperators in place of seperate
//slice elements.  It makes the html rendering much simpleer.
type errOrs map[string]string

//formData is for rendering the longin screen
type formData struct {
	Fields url.Values //this is a map[string][]string type, see net/url package
	Errors errOrs
}

//self explantary, it checks for email validity though Bootstrap css does some
//of that too.
var emailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

func (e errOrs) addError(field, message string) {
	message += ";"
	e[field] += message
}

func (st *sT) newForm(data url.Values, r *http.Request) *formData {
	return &formData{
		data,
		errOrs(map[string]string{}),
	}
}

func (s *formData) getField(field string) string {
	f := s.Fields[field]
	if len(f) == 0 {
		return ""
	}
	return f[0]
}

func (s *formData) fieldRequired(fields ...string) {
	for _, field := range fields {
		value := s.getField(field)
		if strings.TrimSpace(value) == "" {
			s.Errors.addError(field, "this field cannot be blank")
		}
	}
}

func (s *formData) maxLength(field string, d int) {
	value := s.getField(field)
	if value == "" {
		return
	}
	if utf8.RuneCountInString(value) > d {
		s.Errors.addError(field, fmt.Sprintf(`this field is too long
      max lengh is %d charachters`, d))
	}
}

func (s *formData) minLength(field string, d int) {
	value := s.getField(field)
	if value == "" {
		return
	}
	if utf8.RuneCountInString(value) < d {
		s.Errors.addError(field, fmt.Sprintf(`this field is too short
      min lengh is %d charachters`, d))
	}
}

func (s *formData) matchPattern(field string, pattern *regexp.Regexp) {
	value := s.getField(field)
	if value == "" {
		return
	}
	if !pattern.MatchString(value) {
		s.Errors.addError(field, "this field is invalid")
	}
}

func (s *formData) valid() bool {
	return len(s.Errors) == 0
}

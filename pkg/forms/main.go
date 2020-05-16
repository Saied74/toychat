package forms

import (
	"fmt"
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
type ErrOrs map[string]string

//FormData is for rendering the longin screen
type FormData struct {
	Fields url.Values //this is a map[string][]string type, see net/url package
	Errors ErrOrs
}

//self explantary, it checks for email validity though Bootstrap css does some
//of that too.
var EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

func (e ErrOrs) AddError(field, message string) {
	message += ";"
	e[field] += message
}

//NewForm returns an initalized form
func NewForm(data url.Values) *FormData {
	return &FormData{
		data,
		ErrOrs(map[string]string{}),
	}
}

func (s *FormData) GetField(field string) string {
	f := s.Fields[field]
	if len(f) == 0 {
		return ""
	}
	return f[0]
}

func (s *FormData) FieldRequired(fields ...string) {
	for _, field := range fields {
		value := s.GetField(field)
		if strings.TrimSpace(value) == "" {
			s.Errors.AddError(field, "this field cannot be blank")
		}
	}
}

func (s *FormData) MaxLength(field string, d int) {
	value := s.GetField(field)
	if value == "" {
		return
	}
	if utf8.RuneCountInString(value) > d {
		s.Errors.AddError(field, fmt.Sprintf(`this field is too long
      max lengh is %d charachters`, d))
	}
}

func (s *FormData) MinLength(field string, d int) {
	value := s.GetField(field)
	if value == "" {
		return
	}
	if utf8.RuneCountInString(value) < d {
		s.Errors.AddError(field, fmt.Sprintf(`this field is too short
      min lengh is %d charachters`, d))
	}
}

func (s *FormData) MatchPattern(field string, pattern *regexp.Regexp) {
	value := s.GetField(field)
	if value == "" {
		return
	}
	if !pattern.MatchString(value) {
		s.Errors.AddError(field, "this field is invalid")
	}
}

func (s *FormData) Valid() bool {
	return len(s.Errors) == 0
}

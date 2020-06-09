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

//ErrOrs - I have changed the errors from map[string][]string, the way that Alex Edwards
//has it to just map[string]string with semicolmn seperators in place of seperate
//slice elements.  It makes the html rendering much simpleer.
type ErrOrs map[string]string

//FormData is for rendering the longin screen
type FormData struct {
	Fields url.Values //this is a map[string][]string type, see net/url package
	Errors ErrOrs
}

//EmailRX is self explantary, it checks for email validity though Bootstrap
//css does some of that too.
var EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

//AddError adds error to the form to be displayed
func (e ErrOrs) AddError(field, message string) {
	message += ";"
	e[field] += message
}

//NewForm returns an initalized form
func NewForm(data url.Values) *FormData {
	return &FormData{
		Fields: data,
		Errors: ErrOrs{},
	}
}

//GetField gets the sapecified field by html name
func (f *FormData) GetField(field string) string {
	fl := f.Fields[field]
	if len(fl) == 0 {
		return ""
	}
	return fl[0]
}

//FieldRequired runs through the list of names provided and makes sure they exist
func (f *FormData) FieldRequired(fields ...string) {
	for _, field := range fields {
		value := f.GetField(field)
		if strings.TrimSpace(value) == "" {
			f.Errors.AddError(field, "this field cannot be blank")
		}
	}
}

//MaxLength tests the maximum length of the field provided by the form.
func (f *FormData) MaxLength(field string, d int) {
	value := f.GetField(field)
	if value == "" {
		return
	}
	if utf8.RuneCountInString(value) > d {
		f.Errors.AddError(field, fmt.Sprintf(`this field is too long
      max lengh is %d charachters`, d))
	}
}

//MinLength tests the minimum length of the field provided by the form
func (f *FormData) MinLength(field string, d int) {
	value := f.GetField(field)
	if value == "" {
		return
	}
	if utf8.RuneCountInString(value) < d {
		f.Errors.AddError(field, fmt.Sprintf(`this field is too short
      min lengh is %d charachters`, d))
	}
}

//MatchPattern tests the match between the stored password and the provided one.
func (f *FormData) MatchPattern(field string, pattern *regexp.Regexp) {
	value := f.GetField(field)
	if value == "" {
		return
	}
	if !pattern.MatchString(value) {
		f.Errors.AddError(field, "this field is invalid")
	}
}

//Valid tests that all the required fields are provided and there are no errors.
func (f *FormData) Valid() bool {
	return len(f.Errors) == 0
}

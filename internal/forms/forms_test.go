package forms

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestForm_Valid(t *testing.T) {
	r := httptest.NewRequest("POST", "/whatever", nil)
	form := New(r.PostForm)

	isValid := form.Valid()
	if !isValid {
		t.Error("got invalid when should have been valid")
	}
}

func TestForm_Required(t *testing.T) {
	r := httptest.NewRequest("POST", "/whatever", nil)
	form := New(r.PostForm)
	form.Required("a", "b", "c")
	if form.Valid() {
		t.Error("form shows valid when required fields are missing")
	}

	postedData := url.Values{}
	postedData.Add("a", "a")
	postedData.Add("b", "a")
	postedData.Add("c", "a")

	r, _ = http.NewRequest("POST", "/whatever", nil)

	r.PostForm = postedData
	form = New(r.PostForm)
	form.Required("a", "b", "c")
	if !form.Valid() {
		t.Error("shows does not have required fields when it does")
	}

}

func TestForm_Has(t *testing.T) {

	postedData := url.Values{}
	postedData.Add("a", "a")
	postedData.Add("b", "a")
	postedData.Add("c", "a")

	form := New(postedData)
	if !form.Has("c") {
		t.Error("Shows missing field when it exists")
	}

	if form.Has("d") {
		t.Error("Shows field when it is missing")
	}

}

func TestForm_IsEmail(t *testing.T) {

	postedData := url.Values{}
	postedData.Add("eMail1", "a")
	postedData.Add("eMail2", "a@g.com")

	form := New(postedData)

	if form.IsEmail("eMail1") {
		t.Error("Single char recognized as an email")
	}

	if !form.IsEmail("eMail2") {
		t.Error("Correct e-mail not recognized")
	}

}

func TestForm_HasMinLength(t *testing.T) {

	postedData := url.Values{}
	postedData.Add("eMail1", "a")
	postedData.Add("eMail2", "abc")

	form := New(postedData)

	if !form.HasMinLength("eMail1", 1) {
		t.Error("Field has exactly the required length but it shows that it does not")
	}

	if !form.HasMinLength("eMail2", 2) {
		t.Error("Field exceeds the required length but it shows that it does not")
	}

	isError := form.Errors.Get("eMail2")
	if isError != "" {
		t.Error("MinLength throws an error where no error should be")
	}

	if form.HasMinLength("eMail3", 2) {
		t.Error("Postively validated field that does not exist")
	}

	if form.HasMinLength("eMail1", 5) {
		t.Error("Field has not a length of 5 but 1")
	}

	isError = form.Errors.Get("eMail3")
	if isError == "" {
		t.Error("Should have an error but it does not")
	}
}

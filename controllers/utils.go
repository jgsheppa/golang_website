package controllers

import (
	"net/http"
	"net/url"

	"github.com/gorilla/schema"
)

// We want to use this for non-GET requests
func parseForm(r *http.Request, dst interface{}) error {
	if err := r.ParseForm(); err != nil {
		return err
	}
	return parseValues(r.PostForm, dst)
}

// We want to use this function for GET requests
func parseURLParams(r *http.Request, dst interface{}) error {
	if err := r.ParseForm(); err != nil {
		return err
	}
	return parseValues(r.Form, dst)
}

func parseValues(values url.Values, dst interface{}) error {
	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)

	if err := decoder.Decode(dst, values); err != nil {
		return err
	}
	return nil
}
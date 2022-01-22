package models

import "strings"

const (
	ErrNotFound modelError = "models: resource not found"
	ErrIDInvalid modelError = "models: ID provided was invalid"
	ErrPasswordIncorrect modelError = "models: incorrect password"
	ErrRememberTokenTooShort modelError = "models: remember token must be at least 32 bytes"
	ErrRememberRequired modelError = "models: remember token required"
	ErrEmailRequired modelError = "Email address is required"
	ErrEmailInvalid modelError = "Email address is not valid"
	ErrEmailTaken modelError = "models: email address is already taken"
	ErrPasswordMinLength modelError = "Password must be 8 characters long"
	ErrPasswordRequired modelError = "Password is required"

)

type modelError string 

func (e modelError) Error() string {
	return string(e)
}

func (e modelError) Public() string {
	s := strings.Replace(string(e), "models: ", "", 1)
	split := strings.Split(s, " ")
	split[0] = strings.Title(split[0])
	return strings.Join(split, " ")
}


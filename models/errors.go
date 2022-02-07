package models

import "strings"

const (
	ErrNotFound modelError = "models: resource not found"
	ErrIDInvalid privateError = "models: ID provided was invalid"
	ErrPasswordIncorrect modelError = "models: incorrect password"
	ErrRememberTokenTooShort privateError = "models: remember token must be at least 32 bytes"
	ErrRememberRequired privateError = "models: remember token required"
	ErrEmailRequired modelError = "Email address is required"
	ErrEmailInvalid modelError = "Email address is not valid"
	ErrEmailTaken modelError = "models: email address is already taken"
	ErrPasswordMinLength modelError = "Password must be 8 characters long"
	ErrPasswordRequired modelError = "Password is required"
	ErrUserIDRequired privateError = "models: user does not have id"
	ErrTitleRequired modelError = "A title is required for your gallery"

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

type privateError string

func (e privateError) Error() string {
	return string(e)
}

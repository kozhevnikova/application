package main

import (
	"html/template"
	"net/http"
	"regexp"
	"time"

	"github.com/kovetskiy/lorg"
)

func (server *Server) RegisterNewUser(
	writer http.ResponseWriter,
	request *http.Request,

) {
	login := template.HTMLEscapeString(request.FormValue("login"))
	password := template.HTMLEscapeString(request.FormValue("password"))
	email := template.HTMLEscapeString(request.FormValue("email"))

	if login != "" && password != "" && email != "" {
		if ok := MatchLoginContent(login); !ok {
			writer.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		if ok := MatchPasswordContent(password); !ok {
			writer.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		if ok := MatchPasswordLength(password); !ok {
			writer.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		if ok := MatchLoginLength(login); !ok {
			writer.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		if ok := MatchSpaces(password); !ok {
			writer.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		count, err := server.CountExistingUsersByLoginEmail(login, email)
		if err != nil {
			lorg.Error(err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		if count != 0 {
			http.Error(writer, "User already exist", http.StatusConflict)
			return

		} else {
			hashedPassword, err := HashPassword(password)
			if err != nil {
				lorg.Error(err)
				writer.WriteHeader(http.StatusInternalServerError)
				return
			}

			_, err = server.Database.Exec(`
				INSERT INTO users(
					login,
					password, 
					email,
					dateregistration,
					lastvisit) 
						VALUES($1, $2, $3, $4, $5)`,
				login, hashedPassword, email, time.Now(), time.Now())
			if err != nil {
				lorg.Error(err)
				writer.WriteHeader(http.StatusInternalServerError)
				return
			}

			http.Redirect(writer, request, "/", http.StatusFound)
		}

	} else {
		writer.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
}

func MatchPasswordContent(password string) bool {
	match, err := regexp.MatchString("([a-z]{3,}[A-Z]{3,}[0-9]{3,})", password)
	if err != nil {
		lorg.Error(err)
		return false
	}

	return match
}

func MatchLoginContent(login string) bool {
	match, err := regexp.MatchString(`[a-z]{1,}|[A-Z]{1,}|[0-9]{1,}`, login)
	if err != nil {
		lorg.Error(err)
		return false
	}

	return match
}

func MatchPasswordLength(password string) bool {
	return len(password) > 8 && len(password) < 255
}

func MatchLoginLength(login string) bool {
	return len(login) > 8 && len(login) < 255
}

func MatchSpaces(str string) bool {
	match, err := regexp.MatchString(`\s`, str)
	if err != nil {
		lorg.Error(err)
		return false
	}

	return match
}

package main

import (
	"html/template"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/kovetskiy/lorg"
	"github.com/kozhevnikova/channellogger"
)

func (server *Server) Login(
	writer http.ResponseWriter,
	request *http.Request,

) {
	login := template.HTMLEscapeString(request.FormValue("login"))
	password := template.HTMLEscapeString(request.FormValue("password"))

	if login != "" && password != "" {
		users, err := server.CountExistingUsersByLogin(login)
		if err != nil {
			lorg.Error(err)
			return
		}

		if users == 1 {
			var userid int
			var hashedPassword string

			err := server.Database.QueryRow(`
				SELECT password FROM users WHERE login = $1`,
				login,
			).Scan(&hashedPassword)
			if err != nil {
				lorg.Error(err)
				http.Error(writer, "No such user", http.StatusBadRequest)
				return
			}

			if ok := ValidatePassword(password, hashedPassword); !ok {
				http.Error(writer, "Login or password incorrect",
					http.StatusBadRequest)
				return
			}

			err = server.Database.QueryRow(`
				SELECT id FROM users WHERE login = $1 AND password = $2`,
				login,
				hashedPassword,
			).Scan(&userid)
			if err != nil {
				lorg.Error(err)
				channellogger.SendLogInfoToChannel(
					channelLogger,
					"APP::"+err.Error()+
						"::status::"+strconv.Itoa(http.StatusInternalServerError),
				)
				writer.WriteHeader(http.StatusInternalServerError)
				return
			}

			_, err = server.Database.Exec(`
				UPDATE users 
					SET lastvisit = $1 
					WHERE login = $2 AND password = $3 AND id = $4`,
				time.Now(),
				login,
				hashedPassword,
				userid,
			)
			if err != nil {
				lorg.Error(err)
				channellogger.SendLogInfoToChannel(
					channelLogger,
					"APP::"+err.Error()+
						"::status::"+strconv.Itoa(http.StatusInternalServerError),
				)
				writer.WriteHeader(http.StatusInternalServerError)
				return
			}

			user := &SaveUser{
				Userid: userid,
				Login:  login,
			}

			err = user.SetCookie(writer)
			if err != nil {
				lorg.Error(err)
				return
			}

			err = server.SaveSession(request, userid, login)
			if err != nil {
				lorg.Error(err)
				writer.WriteHeader(http.StatusBadRequest)
				return
			}

			http.Redirect(writer, request, "/active", http.StatusFound)

		} else {
			http.Error(writer, "Login or password wrong", http.StatusNotFound)
		}

	} else {
		http.Error(writer, "Login and password can't be empty",
			http.StatusUnprocessableEntity)
	}
}

func Logout(writer http.ResponseWriter, request *http.Request) {
	cookie := &http.Cookie{
		Name:   cookieName,
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}

	http.SetCookie(writer, cookie)
	http.Redirect(writer, request, "/", http.StatusFound)
}

func (server *Server) CountExistingUsersByLoginEmail(
	login string,
	email string,

) (int, error) {

	result, err := server.Database.Exec(`
			SELECT login, email FROM users WHERE login = $1 AND email = $2`,
		login,
		email,
	)
	if err != nil {
		return 0, err
	}

	count, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return int(count), nil
}

func (server *Server) CountExistingUsersByLogin(login string) (int, error) {
	result, err := server.Database.Exec(`
		SELECT login, password FROM users WHERE login = $1`,
		login,
	)
	if err != nil {
		return 0, err
	}

	count, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return int(count), nil
}

func CheckAuth(userid int, username string) bool {
	if userid == 0 || username == "" {
		return false
	}

	return true
}

func (server *Server) SaveSession(
	request *http.Request,
	userid int,
	login string,

) error {

	ip, _, err := net.SplitHostPort(request.RemoteAddr)
	if err != nil {
		lorg.Error(err)
		return err
	}

	_, err = server.Database.Exec(`
		INSERT INTO sessions(
			userid, login, ipaddress, visitdate) 
				VALUES($1, $2, $3, $4)`,
		userid,
		login,
		ip,
		time.Now(),
	)
	if err != nil {
		lorg.Error(err)
		return err
	}

	return nil
}

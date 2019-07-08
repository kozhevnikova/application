package main

import (
	"html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/csrf"
	"github.com/kovetskiy/lorg"
	"github.com/kozhevnikova/channellogger"
)

type CabinetData struct {
	Login              string
	Lastvisit          time.Time
	Dateofregistration time.Time
	Email              string
	CSRFField          template.HTML
}

type Session struct {
	IP        string
	Date      time.Time
	CSRFField template.HTML
}

type SessionsData struct {
	Sessions  []Session
	CSRFField template.HTML
}

func (server *Server) GetPersonalInformation(
	writer http.ResponseWriter,
	request *http.Request,

) (*CabinetData, int, error) {

	var cabinetData CabinetData

	userid, login, err := ReadCookies(writer, request)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	err = server.Database.QueryRow(`
		SELECT lastvisit, dateregistration, email, login 
			FROM users 
			WHERE id = $1 AND login = $2`,
		userid,
		login,
	).Scan(
		&cabinetData.Lastvisit,
		&cabinetData.Dateofregistration,
		&cabinetData.Email,
		&cabinetData.Login,
	)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	cabinetData = CabinetData{
		Login:              cabinetData.Login,
		Lastvisit:          cabinetData.Lastvisit,
		Dateofregistration: cabinetData.Dateofregistration,
		Email:              cabinetData.Email,
		CSRFField:          csrf.TemplateField(request),
	}

	return &cabinetData, http.StatusOK, nil
}

func (server *Server) ChangeEmail(
	writer http.ResponseWriter,
	request *http.Request,

) {
	email := template.HTMLEscapeString(request.FormValue("new_email"))

	userid, login, err := ReadCookies(writer, request)
	if err != nil {
		lorg.Error(err)
		channellogger.SendLogInfoToChannel(
			channelLogger,
			"APP::"+err.Error()+
				"::status::"+strconv.Itoa(http.StatusBadRequest),
		)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	if email != "" {
		_, err := server.Database.Exec(
			`UPDATE users SET email = $1 WHERE id = $2 AND login = $3`,
			email,
			userid,
			login,
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

		http.Redirect(writer, request, "/cabinet", http.StatusSeeOther)

	} else {
		writer.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
}

func (server *Server) ChangePassword(
	writer http.ResponseWriter,
	request *http.Request,

) {
	repeatedpassword := template.HTMLEscapeString(
		request.FormValue("repeated_password"))

	currentpassword := template.HTMLEscapeString(
		request.FormValue("current_password"))

	var dbpassword string

	userid, login, err := ReadCookies(writer, request)
	if err != nil {
		lorg.Error(err)
		channellogger.SendLogInfoToChannel(
			channelLogger,
			"APP::"+err.Error()+
				"::status::"+strconv.Itoa(http.StatusBadRequest),
		)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	err = server.Database.QueryRow(`
		SELECT password FROM users WHERE id = $1 AND login = $2`,
		userid,
		login,
	).Scan(&dbpassword)
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

	if ok := ValidatePassword(currentpassword, dbpassword); !ok {
		http.Error(writer, "Current password incorrect", http.StatusConflict)
		return

	} else {
		if repeatedpassword != "" {
			newRepeatedHashedPassword, err := HashPassword(repeatedpassword)
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
			_, err = server.Database.Exec(
				`UPDATE users SET password = $1 WHERE id = $2 AND login = $3`,
				newRepeatedHashedPassword,
				userid,
				login,
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

			http.Redirect(writer, request, "/cabinet", http.StatusSeeOther)

		} else {
			writer.WriteHeader(http.StatusUnprocessableEntity)
			return
		}
	}
}

func (server *Server) ShowSessions(
	writer http.ResponseWriter,
	request *http.Request,

) (*SessionsData, int, error) {

	var sessionsData SessionsData
	var sessions []Session

	userid, login, err := ReadCookies(writer, request)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	rows, err := server.Database.Query(
		`SELECT ipaddress, visitdate 
			FROM sessions 
			WHERE userid = $1 AND login = $2 
			ORDER BY id desc`,
		userid,
		login,
	)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	for rows.Next() {

		var session Session

		err := rows.Scan(&session.IP, &session.Date)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}

		sessions = append(sessions, session)
	}

	err = rows.Err()
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	err = rows.Close()
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	sessionsData = SessionsData{
		Sessions:  sessions,
		CSRFField: csrf.TemplateField(request),
	}

	return &sessionsData, http.StatusOK, nil
}

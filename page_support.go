package main

import (
	"html/template"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/kovetskiy/lorg"
	"github.com/kozhevnikova/channellogger"
)

func (server *Server) SaveFeedbackInDB(
	userid int,
	login string,
	theme string,
	message string,
	path string,

) (int, error) {

	_, err := server.Database.Exec(`
		INSERT INTO 
			feedbacks(userid, login, date, theme, message, files) 
			VALUES($1, $2, $3, $4, $5, $6)`,
		userid,
		login,
		time.Now(),
		theme,
		message,
		path,
	)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
}

func (server *Server) GetFeedback(
	writer http.ResponseWriter,
	request *http.Request,

) {
	userid, login, err := ReadCookies(writer, request)
	if err != nil {
		lorg.Error(err)
		writer.WriteHeader(http.StatusBadRequest)
		channellogger.SendLogInfoToChannel(
			channelLogger,
			"APP::"+err.Error()+
				"::status::"+strconv.Itoa(http.StatusBadRequest),
		)
		return
	}

	theme := template.HTMLEscapeString(request.FormValue("theme"))
	message := template.HTMLEscapeString(request.FormValue("message"))
	file, handle, err := request.FormFile("file")
	if err != nil {
		if strings.HasSuffix(err.Error(), "no such file") {
			return
		}
		lorg.Error(err)
		writer.WriteHeader(http.StatusBadRequest)
		channellogger.SendLogInfoToChannel(
			channelLogger,
			"APP::"+err.Error()+
				"::status::"+strconv.Itoa(http.StatusBadRequest),
		)

		return
	}

	defer file.Close()

	if len(theme) == 0 || len(message) == 0 {
		http.Error(writer,
			"Forms must be filled",
			http.StatusUnprocessableEntity,
		)
		return

	} else if len(theme) >= 255 || len(message) >= 500 {
		http.Error(writer,
			"Theme must be less than 255 symbols "+
				"and message less than 500 symbols",
			http.StatusUnprocessableEntity,
		)
		return

	} else {
		if ok := ValidateMimeType(handle); ok {
			path, status, err := SaveUploadedFile(file, handle, userid, login)
			if err != nil {
				lorg.Error(err)
				writer.WriteHeader(status)
				channellogger.SendLogInfoToChannel(
					channelLogger, "APP::"+
						err.Error()+
						"::status::"+strconv.Itoa(status),
				)
				return

			} else {
				http.Redirect(writer, request, "/support", http.StatusSeeOther)
			}

			status, err = server.SaveFeedbackInDB(
				userid,
				login,
				theme,
				message,
				path,
			)
			if err != nil {
				lorg.Error(err)
				writer.WriteHeader(status)
				channellogger.SendLogInfoToChannel(
					channelLogger, "APP::"+
						err.Error()+
						"::status::"+strconv.Itoa(status),
				)
				return
			}

		} else {
			http.Error(writer,
				"File type is not supported",
				http.StatusBadRequest,
			)
		}
	}
}

func CreateDirForUploadedFile(userid string, login string) error {
	err := os.MkdirAll("./"+userid+"/"+login+"/", os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

func SaveUploadedFile(
	file multipart.File,
	handle *multipart.FileHeader,
	userid int,
	login string,

) (string, int, error) {

	var path string

	if handle.Filename == "" {
		return "no file there", http.StatusUnsupportedMediaType, nil

	} else {

		id := strconv.Itoa(userid)

		path = "./" + id + "/" + login + "/" + handle.Filename

		err := CreateDirForUploadedFile(id, login)
		if err != nil {
			return "", http.StatusInternalServerError, err
		}

		data, err := ioutil.ReadAll(file)
		if err != nil {
			return "", http.StatusInternalServerError, err
		}

		err = ioutil.WriteFile(path, data, 0666)
		if err != nil {
			return "", http.StatusInternalServerError, err
		}
	}

	return path, http.StatusOK, nil
}

func ValidateMimeType(handler *multipart.FileHeader) bool {

	var whitelist = []string{
		"image/png",
		"image/jpeg",
	}

	for _, value := range whitelist {
		if handler.Header.Get("Content-Type") == value {
			return true
		}
	}

	return false
}

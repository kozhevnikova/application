package main

import (
	"net/http"
	"strconv"

	"github.com/gorilla/securecookie"
)

const cookieName = "_data"

var hashKey = []byte("")
var sc = securecookie.New(hashKey, nil)

type SaveUser struct {
	Userid int
	Login  string
}

func (user *SaveUser) GetCookie() (*http.Cookie, error) {
	value := map[string]string{
		"userid": strconv.Itoa(user.Userid),
		"login":  user.Login,
	}

	encoded, err := sc.Encode(cookieName, value)
	if err != nil {
		return nil, err
	}

	cookie := &http.Cookie{
		Name:     cookieName,
		Value:    encoded,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   1800,
	}

	return cookie, nil
}

func (user *SaveUser) SetCookie(writer http.ResponseWriter) error {
	cookie, err := user.GetCookie()
	if err != nil {
		return err
	}

	http.SetCookie(writer, cookie)

	return nil
}

func ReadCookies(
	writer http.ResponseWriter,
	request *http.Request,

) (int, string, error) {

	var userid int
	var login string

	if cookie, err := request.Cookie(cookieName); err == nil {
		value := make(map[string]string)
		if err = sc.Decode(cookieName, cookie.Value, &value); err == nil {
			userid, err = strconv.Atoi(value["userid"])
			if err != nil {
				return 0, "", err
			}
			login = value["login"]
		}
	} else {
		return 0, "", err
	}

	return userid, login, nil
}

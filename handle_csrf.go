package main

import (
	"html/template"
	"net/http"

	"github.com/gorilla/csrf"
)

type CSRF struct {
	CSRFField template.HTML
}

func ReturnCSRFField(request *http.Request) CSRF {
	csrfObject := CSRF{
		CSRFField: csrf.TemplateField(request),
	}

	return csrfObject
}

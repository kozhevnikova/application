package main

import (
	"html/template"
)

var LoginT *template.Template
var ActiveT *template.Template
var StatisticsT *template.Template
var StatisticsLastWeekT *template.Template
var StatisticsLastMonthT *template.Template
var CabinetT *template.Template
var HistoryT *template.Template
var SignupT *template.Template
var ScheduledT *template.Template
var SupportT *template.Template
var NotfoundT *template.Template
var SessionsT *template.Template

func init() {
	LoginT = template.Must(template.ParseFiles(
		"templates/pages/login.html"))

	SignupT = template.Must(template.ParseFiles(
		"templates/pages/signup.html"))

	ActiveT = template.Must(template.ParseFiles(
		"templates/pages/active.html",
		"templates/pages/nav.html",
		"templates/pages/responsive-nav.html",
		"templates/pages/show-nav-bar.html",
		"templates/pages/footer.html"))

	StatisticsT = template.Must(template.ParseFiles(
		"templates/pages/statistics.html",
		"templates/pages/nav.html",
		"templates/pages/responsive-nav.html",
		"templates/pages/show-nav-bar.html",
		"templates/pages/footer.html"))

	StatisticsLastWeekT = template.Must(template.ParseFiles(
		"templates/pages/statistics-last-week.html",
		"templates/pages/nav.html",
		"templates/pages/responsive-nav.html",
		"templates/pages/show-nav-bar.html",
		"templates/pages/footer.html"))

	StatisticsLastMonthT = template.Must(template.ParseFiles(
		"templates/pages/statistics-last-month.html",
		"templates/pages/nav.html",
		"templates/pages/responsive-nav.html",
		"templates/pages/show-nav-bar.html",
		"templates/pages/footer.html"))

	HistoryT = template.Must(template.ParseFiles(
		"templates/pages/history.html",
		"templates/pages/nav.html",
		"templates/pages/responsive-nav.html",
		"templates/pages/show-nav-bar.html",
		"templates/pages/footer.html"))

	ScheduledT = template.Must(template.ParseFiles(
		"templates/pages/scheduled.html",
		"templates/pages/nav.html",
		"templates/pages/responsive-nav.html",
		"templates/pages/show-nav-bar.html",
		"templates/pages/footer.html"))

	SupportT = template.Must(template.ParseFiles(
		"templates/pages/support.html",
		"templates/pages/nav.html",
		"templates/pages/responsive-nav.html",
		"templates/pages/show-nav-bar.html",
		"templates/pages/footer.html"))

	CabinetT = template.Must(template.ParseFiles(
		"templates/pages/cabinet.html",
		"templates/pages/nav.html",
		"templates/pages/responsive-nav.html",
		"templates/pages/show-nav-bar.html",
		"templates/pages/footer.html"))

	SessionsT = template.Must(template.ParseFiles(
		"templates/pages/sessions.html",
		"templates/pages/nav.html",
		"templates/pages/responsive-nav.html",
		"templates/pages/show-nav-bar.html",
		"templates/pages/footer.html"))

	NotfoundT = template.Must(template.ParseFiles(
		"templates/pages/notfound.html"))
}

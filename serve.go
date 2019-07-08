package main

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi"
)

func NewRouter(server *Server) chi.Router {
	router := chi.NewRouter()

	router.Group(func(group chi.Router) {
		group.Use(MiddlewareCheckCredentials)

		group.Get("/active", server.HandleActivePage)
		group.Get("/history", server.HandleHistoryPage)
		group.Get("/scheduled", server.HandleScheduledPage)
		group.Get("/cabinet", server.HandleCabinetPage)
		group.Get("/cabinet/sessions", server.HandleSessionsPage)
		group.Get("/support", server.HandleSupportPage)
		group.Get("/statistics", server.HandleStatisticsPage)
		group.Get("/statistics/lastweek", server.HandleStatisticsLastWeekPage)
		group.Get("/statistics/lastmonth", server.HandleStatisticsLastMonthPage)
		group.Get("/data/get/categories", server.SendCategorySubcategoryList)

		group.Route("/api/v1", func(api chi.Router) {
			api.Get(
				"/statistics/charts/currency/get/week/transactions",
				server.SendLastWeekToCurrencyCharts,
			)
			api.Get(
				"/statistics/charts/currency/get/month/transactions",
				server.SendLastMonthToCurrencyChart,
			)
			api.Get(
				"/statistics/charts/category/get/week/transactions",
				server.SendLastWeekToCategoryCharts,
			)
			api.Get(
				"/statistics/charts/category/get/month/transactions",
				server.SendLastMonthToCategoryChart,
			)
			api.Get(
				"/statistics/charts/currency/get/list",
				server.SendCurrencyList,
			)
			api.Get(
				"/statistics/charts/category/get/week/list",
				server.SendCategoryListForWeek,
			)
			api.Get(
				"/statistics/charts/category/get/month/list",
				server.SendCategoryListForMonth,
			)
			api.Get(
				"/statistics/charts/category/get/week/days",
				SendWeekSliceToCategoryChart,
			)
			api.Get(
				"/statistics/charts/category/get/month/days",
				SendMonthListToCategoryChart,
			)
		})

		group.Post("/active/record/add", server.AddTransaction)
		group.Post("/active/record/delete", server.DeleteRecordFromActivePage)
		group.Post("/history/record/delete", server.DeleteRecordFromHistoryPage)
		group.Post("/scheduled/payment/create", server.CreateScheduledPayment)
		group.Post("/scheduled/payment/delete", server.DeleteScheduledPayment)
		group.Post("/cabinet/password/change", server.ChangePassword)
		group.Post("/cabinet/email/change", server.ChangeEmail)
		group.Post("/support/feedback/send", server.GetFeedback)
	})

	router.Get("/", server.HandleLoginPage)
	router.Get("/signup", server.HandleSignupPage)
	router.Post("/login", server.Login)
	router.Post("/logout", Logout)
	router.Post("/register", server.RegisterNewUser)

	router.Get("/static/*", server.StaticFiles)
	router.Get("/dynamic/*", server.DynamicFiles)
	router.NotFound(server.PageNotFound)

	return router
}

func (server *Server) StaticFiles(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/static/") {
		r.URL.Path = strings.TrimPrefix(r.URL.Path, "/static/")
		server.Static.ServeHTTP(w, r)
	}
}

func (server *Server) DynamicFiles(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/dynamic/") {
		r.URL.Path = strings.TrimPrefix(r.URL.Path, "/dynamic/")
		server.Dynamic.ServeHTTP(w, r)
	}
}

func (server *Server) PageNotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	server.HandleNotFoundPage(w, r)
}

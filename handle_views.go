package main

import (
	"net/http"
	"strconv"

	"github.com/gorilla/csrf"
	"github.com/kovetskiy/lorg"
	"github.com/kozhevnikova/channellogger"
)

func (server *Server) HandleLoginPage(
	writer http.ResponseWriter,
	request *http.Request,

) {
	writer.Header().Set("Cache-Control", "no-store")
	writer.Header().Set("Cache-Control", "must-revalidate")

	err := LoginT.Execute(writer, map[string]interface{}{
		csrf.TemplateTag: csrf.TemplateField(request)})
	if err != nil {
		lorg.Error("INDEX PAGE >", err)
		return
	}
}

func (server *Server) HandleSignupPage(
	writer http.ResponseWriter,
	request *http.Request,

) {
	writer.Header().Set("Cache-Control", "no-store")
	writer.Header().Set("Cache-Control", "must-revalidate")

	err := SignupT.Execute(writer, map[string]interface{}{
		csrf.TemplateTag: csrf.TemplateField(request)})
	if err != nil {
		lorg.Error("SIGNUP PAGE >", err)
		return
	}
}

func (server *Server) HandleActivePage(
	writer http.ResponseWriter,
	request *http.Request,

) {
	writer.Header().Set("Cache-Control", "no-store")
	writer.Header().Set("Cache-Control", "must-revalidate")

	week, status, err := server.showLastWeek(writer, request)
	if err != nil {
		lorg.Error(err)
		writer.WriteHeader(status)
		channellogger.SendLogInfoToChannel(channelLogger, "APP::"+err.Error()+
			"::status::"+strconv.Itoa(status))
		return
	}

	err = ActiveT.Execute(writer, week)
	if err != nil {
		lorg.Error("ACTIVE PAGE >", err)
		return
	}
}

func (server *Server) HandleScheduledPage(
	writer http.ResponseWriter,
	request *http.Request,

) {
	writer.Header().Set("Cache-Control", "no-store")
	writer.Header().Set("Cache-Control", "must-revalidate")

	s, status, err := server.ShowScheduledList(writer, request)
	if err != nil {
		lorg.Error(err)
		writer.WriteHeader(status)
		channellogger.SendLogInfoToChannel(channelLogger, "APP::"+err.Error()+
			"::status::"+strconv.Itoa(status))
		return
	}

	err = ScheduledT.Execute(writer, s)
	if err != nil {
		lorg.Error("ROUTINE PAGE >", err)
		return
	}
}

func (server *Server) HandleStatisticsPage(
	writer http.ResponseWriter,
	request *http.Request,

) {
	writer.Header().Set("Cache-Control", "no-store")
	writer.Header().Set("Cache-Control", "must-revalidate")

	csrf := ReturnCSRFField(request)

	err := StatisticsT.Execute(writer, csrf)
	if err != nil {
		lorg.Error("STATISTICS PAGE >", err)
		channellogger.SendLogInfoToChannel(channelLogger, "APP::"+err.Error())
		return
	}
}

func (server *Server) HandleStatisticsLastWeekPage(
	writer http.ResponseWriter,
	request *http.Request,

) {
	writer.Header().Set("Cache-Control", "no-store")
	writer.Header().Set("Cache-Control", "must-revalidate")

	csrf := ReturnCSRFField(request)

	err := StatisticsLastWeekT.Execute(writer, csrf)
	if err != nil {
		lorg.Error("STATISTICS LAST WEEK PAGE >", err)
		channellogger.SendLogInfoToChannel(channelLogger, "APP::"+err.Error())
		return
	}
}

func (server *Server) HandleStatisticsLastMonthPage(
	writer http.ResponseWriter,
	request *http.Request,

) {
	writer.Header().Set("Cache-Control", "no-store")
	writer.Header().Set("Cache-Control", "must-revalidate")

	csrf := ReturnCSRFField(request)

	err := StatisticsLastMonthT.Execute(writer, csrf)
	if err != nil {
		lorg.Error("STATISTICS LAST MONTH PAGE >", err)
		channellogger.SendLogInfoToChannel(channelLogger, "APP::"+err.Error())
		return
	}
}

func (server *Server) HandleCabinetPage(
	writer http.ResponseWriter,
	request *http.Request,

) {
	writer.Header().Set("Cache-Control", "no-store")
	writer.Header().Set("Cache-Control", "must-revalidate")

	data, status, err := server.GetPersonalInformation(writer, request)
	if err != nil {
		lorg.Error(err)
		writer.WriteHeader(status)
		channellogger.SendLogInfoToChannel(channelLogger, "APP::"+err.Error()+
			"::status::"+strconv.Itoa(status))
		return
	}

	err = CabinetT.Execute(writer, map[string]interface{}{
		"data": data,
	})
	if err != nil {
		lorg.Error("CABINET PAGE >", err)
		return
	}
}

func (server *Server) HandleSessionsPage(
	writer http.ResponseWriter,
	request *http.Request,

) {
	writer.Header().Set("Cache-Control", "no-store")
	writer.Header().Set("Cache-Control", "must-revalidate")

	sessions, status, err := server.ShowSessions(writer, request)
	if err != nil {
		lorg.Error(err)
		writer.WriteHeader(status)
		channellogger.SendLogInfoToChannel(channelLogger, "APP::"+err.Error()+
			"::status::"+strconv.Itoa(status))
		return
	}

	err = SessionsT.Execute(writer, sessions)
	if err != nil {
		lorg.Error("SESSIONS PAGE >", err)
		return
	}
}

func (server *Server) HandleHistoryPage(
	writer http.ResponseWriter,
	request *http.Request,

) {
	writer.Header().Set("Cache-Control", "no-store")
	writer.Header().Set("Cache-Control", "must-revalidate")

	month, err := server.ShowLastMonth(writer, request)
	if err != nil {
		lorg.Error(err)
		return
	}

	err = HistoryT.Execute(writer, month)
	if err != nil {
		lorg.Error("HISTORY PAGE >", err)
		return
	}
}

func (server *Server) HandleSupportPage(
	writer http.ResponseWriter,
	request *http.Request,

) {
	writer.Header().Set("Cache-Control", "no-store")
	writer.Header().Set("Cache-Control", "must-revalidate")

	csrf := ReturnCSRFField(request)
	err := SupportT.Execute(writer, csrf)
	if err != nil {
		lorg.Error("SUPPORT PAGE >", err)
		return
	}
}

func (server *Server) HandleNotFoundPage(
	writer http.ResponseWriter,
	request *http.Request,

) {
	writer.Header().Set("Cache-Control", "no-store")
	writer.Header().Set("Cache-Control", "must-revalidate")

	csrf := ReturnCSRFField(request)
	err := NotfoundT.Execute(writer, csrf)
	if err != nil {
		lorg.Error("PAGE NOT FOUND >", err)
		return
	}
}

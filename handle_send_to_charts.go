package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/csrf"
	"github.com/kovetskiy/lorg"
	"github.com/kozhevnikova/channellogger"
)

func (server *Server) SendCurrencyList(
	w http.ResponseWriter,
	r *http.Request,

) {
	w.Header().Set("X-CSRF-Token", csrf.Token(r))

	buffer := new(bytes.Buffer)

	currencyType, status, err := server.GetCurrencyType()
	if err != nil {
		lorg.Error(err)
		w.WriteHeader(status)
		channellogger.SendLogInfoToChannel(channelLogger, "APP::"+err.Error()+
			"::status::"+strconv.Itoa(status))
		return
	}

	err = json.NewEncoder(buffer).Encode(currencyType)
	if err != nil {
		lorg.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		channellogger.SendLogInfoToChannel(channelLogger, "APP::"+err.Error()+
			"::status::"+strconv.Itoa(http.StatusInternalServerError))
		return
	}

	_, err = w.Write(buffer.Bytes())
	if err != nil {
		lorg.Error(err)
		return
	}
}

func (server *Server) SendLastWeekToCategoryCharts(
	w http.ResponseWriter,
	r *http.Request,

) {
	w.Header().Set("X-CSRF-Token", csrf.Token(r))

	buffer, status, err := server.GetDataForCategoryChartLastWeek(w, r)
	if err != nil {
		lorg.Error(err)
		w.WriteHeader(status)
		channellogger.SendLogInfoToChannel(channelLogger, "APP::"+err.Error()+
			"::status::"+strconv.Itoa(status))
		return
	}

	_, err = w.Write(buffer)
	if err != nil {
		lorg.Error(err)
		return
	}
}

func (server *Server) SendLastWeekToCurrencyCharts(
	w http.ResponseWriter,
	r *http.Request,

) {
	w.Header().Set("X-CSRF-Token", csrf.Token(r))

	buffer, status, err := server.GetDataForCurrencyChartLastWeek(w, r)
	if err != nil {
		lorg.Error(err)
		w.WriteHeader(status)
		channellogger.SendLogInfoToChannel(channelLogger, "APP::"+err.Error()+
			"::status::"+strconv.Itoa(status))
		return
	}

	_, err = w.Write(buffer)
	if err != nil {
		lorg.Error(err)
		return
	}
}

func SendWeekSliceToCategoryChart(
	w http.ResponseWriter,
	r *http.Request,

) {
	w.Header().Set("X-CSRF-Token", csrf.Token(r))

	buffer := new(bytes.Buffer)

	type Week struct {
		Day int `json:"day"`
	}

	var full []Week
	var ww Week
	week := MakeWeekSlice()
	for _, i := range week {
		ww = Week{
			Day: i,
		}
		full = append(full, ww)
	}

	err := json.NewEncoder(buffer).Encode(full)
	if err != nil {
		lorg.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = w.Write(buffer.Bytes())
	if err != nil {
		lorg.Error(err)
		return
	}
}

func (server *Server) SendLastMonthToCurrencyChart(
	w http.ResponseWriter,
	r *http.Request,

) {
	w.Header().Set("X-CSRF-Token", csrf.Token(r))
	buffer, status, err := server.GetDataForCurrencyChartLastMonth(w, r)
	if err != nil {
		lorg.Error(err)
		w.WriteHeader(status)
		channellogger.SendLogInfoToChannel(channelLogger, "APP::"+err.Error()+
			"::status::"+strconv.Itoa(status))
		return
	}

	_, err = w.Write(buffer)
	if err != nil {
		lorg.Error(err)
		return
	}
}

func (server *Server) SendCategoryListForWeek(
	w http.ResponseWriter,
	r *http.Request,

) {
	w.Header().Set("X-CSRF-Token", csrf.Token(r))

	buffer := new(bytes.Buffer)

	categories, status, err := server.GetCategoriesByPeriod(w, r, "week")
	if err != nil {
		lorg.Error(err)
		w.WriteHeader(status)
		channellogger.SendLogInfoToChannel(channelLogger, "APP::"+err.Error()+
			"::status::"+strconv.Itoa(status))
		return
	}

	err = json.NewEncoder(buffer).Encode(categories)
	if err != nil {
		lorg.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = w.Write(buffer.Bytes())
	if err != nil {
		lorg.Error(err)
		return
	}
}
func (server *Server) SendCategoryListForMonth(
	w http.ResponseWriter,
	r *http.Request,

) {
	w.Header().Set("X-CSRF-Token", csrf.Token(r))

	buffer := new(bytes.Buffer)

	categories, status, err := server.GetCategoriesByPeriod(w, r, "month")
	if err != nil {
		channellogger.SendLogInfoToChannel(channelLogger, "APP::"+err.Error()+
			"::status::"+strconv.Itoa(status))
		return
	}

	err = json.NewEncoder(buffer).Encode(categories)
	if err != nil {
		lorg.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = w.Write(buffer.Bytes())
	if err != nil {
		lorg.Error(err)
		return
	}
}

func SendMonthListToCategoryChart(
	w http.ResponseWriter,
	r *http.Request,

) {
	w.Header().Set("X-CSRF-Token", csrf.Token(r))
	month, _ := MakeMonthSlice()

	buffer := new(bytes.Buffer)

	err := json.NewEncoder(buffer).Encode(month)
	if err != nil {
		lorg.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = w.Write(buffer.Bytes())
	if err != nil {
		lorg.Error(err)
		return
	}
}

func (server *Server) SendLastMonthToCategoryChart(
	w http.ResponseWriter,
	r *http.Request,

) {
	w.Header().Set("X-CSRF-Token", csrf.Token(r))

	categories, status, err := server.GetDataForCategoryChartLastMonth(w, r)
	if err != nil {
		lorg.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		channellogger.SendLogInfoToChannel(channelLogger, "APP::"+err.Error()+
			"::status::"+strconv.Itoa(status))
		return
	}

	buffer := new(bytes.Buffer)
	err = json.NewEncoder(buffer).Encode(categories)
	if err != nil {
		lorg.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = w.Write(buffer.Bytes())
	if err != nil {
		lorg.Error(err)
		return
	}
}
func (server *Server) SendCategorySubcategoryList(
	w http.ResponseWriter,
	r *http.Request,

) {
	w.Header().Set("X-CSRF-Token", csrf.Token(r))

	buffer := new(bytes.Buffer)

	data, err := server.GetCategoriesList()
	if err != nil {
		lorg.Error(err)
		channellogger.SendLogInfoToChannel(channelLogger, "APP::"+err.Error()+
			"::status::"+strconv.Itoa(http.StatusInternalServerError))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(buffer).Encode(data)
	if err != nil {
		lorg.Error(err)
		channellogger.SendLogInfoToChannel(channelLogger, "APP::"+err.Error()+
			"::status::"+strconv.Itoa(http.StatusInternalServerError))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = w.Write(buffer.Bytes())
	if err != nil {
		lorg.Error(err)
		return
	}
}

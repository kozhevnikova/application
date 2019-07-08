package main

import (
	"html/template"
	"net/http"
	"strconv"

	"github.com/gorilla/csrf"
	"github.com/kovetskiy/lorg"
	"github.com/kozhevnikova/channellogger"
)

//transaction struct redeclared in active page
type HistoryData struct {
	Transactions []Transaction
	CSRFField    template.HTML
}

func (server *Server) DeleteRecordFromHistoryPage(
	writer http.ResponseWriter,
	request *http.Request,

) {
	transactionid := template.HTMLEscapeString(
		request.FormValue("transactionid"))

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

	_, err = server.Database.Exec(`
			DELETE FROM transactions
				WHERE id=$1 AND userid=$2 AND login=$3`,
		transactionid,
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

	http.Redirect(writer, request, "/history", http.StatusSeeOther)
}

func (server *Server) ShowLastMonth(
	writer http.ResponseWriter,
	request *http.Request,

) (*HistoryData, error) {

	var historyData HistoryData
	var transactions []Transaction

	userid, login, err := ReadCookies(writer, request)
	if err != nil {
		lorg.Error(err)
		channellogger.SendLogInfoToChannel(
			channelLogger,
			"APP::"+err.Error()+
				"::status::"+strconv.Itoa(http.StatusBadRequest),
		)
		writer.WriteHeader(http.StatusBadRequest)
		return nil, err
	}

	rows, err := server.Database.Query(
		`SELECT 
			t.id,
			t.transactiondate,
			c.type,
			s.type,
			t.place,
			t.price::money::numeric::float8,
			cu.type,
			m.type 
			FROM transactions t
				JOIN categories c ON t.categoryid = c.id
				JOIN subcategories s ON t.subcategoryid = s.id
				JOIN currencies cu ON t.currencyid = cu.id
				JOIN methods m ON t.methodid = m.id
				WHERE userid = $1 AND login = $2 
					ORDER BY transactiondate DESC`,
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
		return nil, err
	}

	for rows.Next() {
		var transaction Transaction
		err := rows.Scan(&transaction.Transactionid,
			&transaction.Transactiondate,
			&transaction.Category,
			&transaction.Subcategory,
			&transaction.Place,
			&transaction.Price,
			&transaction.Currency,
			&transaction.Method,
		)
		if err != nil {
			lorg.Error(err)
			channellogger.SendLogInfoToChannel(
				channelLogger,
				"APP::"+err.Error()+
					"::status::"+strconv.Itoa(http.StatusInternalServerError),
			)
			writer.WriteHeader(http.StatusInternalServerError)
			return nil, err
		}

		transactions = append(transactions, transaction)
	}

	err = rows.Err()
	if err != nil {
		lorg.Error(err)
		channellogger.SendLogInfoToChannel(
			channelLogger,
			"APP::"+err.Error()+
				"::status::"+strconv.Itoa(http.StatusInternalServerError),
		)
		writer.WriteHeader(http.StatusInternalServerError)
		return nil, err
	}

	err = rows.Close()
	if err != nil {
		lorg.Error(err)
		channellogger.SendLogInfoToChannel(
			channelLogger,
			"APP::"+err.Error()+
				"::status::"+strconv.Itoa(http.StatusInternalServerError),
		)
		writer.WriteHeader(http.StatusInternalServerError)
		return nil, err
	}

	historyData = HistoryData{
		Transactions: transactions,
		CSRFField:    csrf.TemplateField(request),
	}

	return &historyData, nil
}

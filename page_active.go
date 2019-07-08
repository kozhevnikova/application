package main

import (
	"html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/gorilla/csrf"
	"github.com/kovetskiy/lorg"
	"github.com/kozhevnikova/channellogger"
)

type Transaction struct {
	Transactionid   int
	Transactiondate time.Time
	Category        string
	Subcategory     string
	Place           string
	Price           float64
	Currency        string
	Method          string
}

type LastsTransactions struct {
	Transactions []Transaction
	CSRFField    template.HTML
}

func (server *Server) AddTransaction(
	writer http.ResponseWriter,
	request *http.Request,

) {
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

	category := template.HTMLEscapeString(request.FormValue("category"))
	subcategory := template.HTMLEscapeString(request.FormValue("subcategory"))
	date := template.HTMLEscapeString(request.FormValue("date"))
	place := template.HTMLEscapeString(request.FormValue("place"))
	price := template.HTMLEscapeString(request.FormValue("price"))
	currency := template.HTMLEscapeString(request.FormValue("currency"))
	method := template.HTMLEscapeString(request.FormValue("method"))

	if ok := ValidatePrice(price); !ok {
		writer.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	if ok := ValidatePlaceLen(place); ok {
		writer.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	if category != "0" ||
		subcategory != "0" ||
		date != "" ||
		place != "" ||
		price != "" ||
		currency != "" ||
		method != "" {

		_, err = server.Database.Exec(
			`INSERT INTO 
				transactions 
					(userid,login,insertdate,transactiondate,price, categoryid,
					subcategoryid,currencyid,place,methodid,scheduled)
				VALUES($1,$2,$3,$4,$5,
					(SELECT c.id FROM categories c WHERE c.type = $6),
					(SELECT s.id FROM subcategories s WHERE s.type = $7),
					(SELECT cu.id FROM currencies cu WHERE cu.type = $8),
					$9,
					(SELECT m.id FROM methods m WHERE m.type = $10),
					$11)`,
			userid,
			login,
			time.Now(),
			date,
			price,
			category,
			subcategory,
			currency,
			place,
			method,
			"false",
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

		} else {
			http.Redirect(writer, request, "/active", http.StatusSeeOther)
		}

	} else {
		writer.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
}

func (server *Server) showLastWeek(
	writer http.ResponseWriter,
	request *http.Request,

) (LastsTransactions, int, error) {

	var lastTransactions LastsTransactions
	var transactions []Transaction

	userid, login, err := ReadCookies(writer, request)
	if err != nil {
		return lastTransactions, http.StatusBadRequest, err
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
				JOIN categories c on t.categoryid = c.id
				JOIN subcategories s on t.subcategoryid = s.id
				JOIN currencies cu on t.currencyid = cu.id
				JOIN methods m on t.methodid = m.id
				WHERE userid = $1 
					AND login = $2 
					AND insertdate::date 
					BETWEEN current_date-integer'7' 
					AND current_date 
					ORDER BY id DESC`,
		userid,
		login,
	)
	if err != nil {
		return lastTransactions, http.StatusInternalServerError, err
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
			&transaction.Method)

		if err != nil {
			return lastTransactions, http.StatusInternalServerError, err
		}

		transactions = append(transactions, transaction)
	}

	lastTransactions = LastsTransactions{
		Transactions: transactions,
		CSRFField:    csrf.TemplateField(request),
	}

	err = rows.Err()
	if err != nil {
		return lastTransactions, http.StatusInternalServerError, err
	}

	err = rows.Close()
	if err != nil {
		return lastTransactions, http.StatusInternalServerError, err
	}

	return lastTransactions, http.StatusOK, nil
}

func (server *Server) DeleteRecordFromActivePage(
	writer http.ResponseWriter,
	request *http.Request,

) {
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

	transactionid := template.HTMLEscapeString(
		request.FormValue("transactionid"))

	_, err = server.Database.Exec(`
			DELETE FROM transactions 
			WHERE 
				id = $1 AND userid = $2 AND login = $3`,
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

	http.Redirect(writer, request, "/active", http.StatusSeeOther)
}

func ValidatePrice(price string) bool {
	return govalidator.IsInt(price)
}

func ValidatePlaceLen(place string) bool {
	return len(place) > 30
}

package main

import (
	"database/sql"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/csrf"
	"github.com/kovetskiy/lorg"
	"github.com/kozhevnikova/channellogger"
)

type Title struct {
	Count       int64
	Paymentid   int
	Category    string
	Subcategory string
	Createdate  time.Time
	Weekname    *string
	Monthday    *int
	Place       string
	Price       float64
	Currency    string
	Method      string
	Passed      []Passed
}

type Passed struct {
	Category    string
	Subcategory string
	Insertdate  time.Time
	Price       float64
	Currency    string
	Method      string
}

type ScheduledPaymentsAndTransactions struct {
	Titles    []Title
	CSRFField template.HTML
}

func (server *Server) ShowScheduledList(
	writer http.ResponseWriter,
	request *http.Request,

) (*ScheduledPaymentsAndTransactions, int, error) {

	var scheduledAndPassed ScheduledPaymentsAndTransactions
	var titles []Title
	var allTransactions []Title

	userid, login, err := ReadCookies(writer, request)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	titlesQuery, err := server.Database.Query(`
		SELECT 
			sc.id,
			c.type,
			s.type,
			sc.createdate,
			sc.weekname,
			sc.monthday,
			sc.place,
			sc.price::money::numeric::float8,
			cu.type,m.type 
			FROM scheduledpayments  sc
				JOIN categories c ON sc.categoryid = c.id
				JOIN subcategories s ON sc.subcategoryid = s.id
				JOIN currencies cu ON sc.currencyid = cu.id
				JOIN methods m ON sc.methodid = m.id
				WHERE userid = $1 AND login = $2 ORDER BY id DESC`,
		userid,
		login,
	)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	for titlesQuery.Next() {

		var mainTitle Title

		err := titlesQuery.Scan(
			&mainTitle.Paymentid,
			&mainTitle.Category,
			&mainTitle.Subcategory,
			&mainTitle.Createdate,
			&mainTitle.Weekname,
			&mainTitle.Monthday,
			&mainTitle.Place,
			&mainTitle.Price,
			&mainTitle.Currency,
			&mainTitle.Method,
		)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}

		err = server.Database.QueryRow(`
			SELECT COUNT(*) 
				FROM transactions 
				WHERE userid = $1 AND login = $2 AND scheduledid = $3`,
			userid,
			login,
			mainTitle.Paymentid,
		).Scan(&mainTitle.Count)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}

		titles = append(titles, mainTitle)
	}

	err = titlesQuery.Err()
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	err = titlesQuery.Close()
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	for _, oneTitle := range titles {
		passedTransactionsQuery, err := server.Database.Query(`
			SELECT 
				c.type, s.type,
				t.insertdate,
				t.price::money::numeric::float8,
				cu.type, m.type
			FROM transactions t
			JOIN categories c ON t.categoryid = c.id
			JOIN subcategories s ON t.subcategoryid = s.id
			JOIN currencies cu ON t.currencyid = cu.id
			JOIN methods m ON t.methodid = m.id
			WHERE userid = $1 AND login = $2 and t.scheduledid = $3`,
			userid,
			login,
			oneTitle.Paymentid,
		)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return nil, http.StatusInternalServerError, err
		}

		var passedTransactions []Passed

		for passedTransactionsQuery.Next() {
			var passedTransaction Passed

			err := passedTransactionsQuery.Scan(
				&passedTransaction.Category,
				&passedTransaction.Subcategory,
				&passedTransaction.Insertdate,
				&passedTransaction.Price,
				&passedTransaction.Currency,
				&passedTransaction.Method)
			if err != nil {
				return nil, http.StatusInternalServerError, err
			}

			passedTransactions = append(
				passedTransactions, passedTransaction)

			oneTitle = Title{
				Count:       oneTitle.Count,
				Paymentid:   oneTitle.Paymentid,
				Category:    oneTitle.Category,
				Subcategory: oneTitle.Subcategory,
				Createdate:  oneTitle.Createdate,
				Weekname:    oneTitle.Weekname,
				Monthday:    oneTitle.Monthday,
				Place:       oneTitle.Place,
				Price:       oneTitle.Price,
				Currency:    oneTitle.Currency,
				Method:      oneTitle.Method,
				Passed:      passedTransactions,
			}
		}

		allTransactions = append(
			allTransactions,
			oneTitle,
		)

		err = passedTransactionsQuery.Close()
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return nil, http.StatusInternalServerError, err
		}
	}

	scheduledAndPassed = ScheduledPaymentsAndTransactions{
		Titles:    allTransactions,
		CSRFField: csrf.TemplateField(request),
	}

	return &scheduledAndPassed, http.StatusOK, nil
}

func NullStringValue(str string) sql.NullString {
	if len(str) == 0 {
		return sql.NullString{}
	}

	return sql.NullString{
		String: str,
		Valid:  true,
	}
}

func (server *Server) CreateScheduledPayment(
	writer http.ResponseWriter,
	request *http.Request,

) {
	category := template.HTMLEscapeString(request.FormValue("category"))
	subcategory := template.HTMLEscapeString(request.FormValue("subcategory"))
	monthday := template.HTMLEscapeString(request.FormValue("month"))
	week := template.HTMLEscapeString(request.FormValue("week"))
	method := template.HTMLEscapeString(request.FormValue("method"))
	price, err := strconv.Atoi(
		template.HTMLEscapeString(request.FormValue("price")))
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
	currency := template.HTMLEscapeString(request.FormValue("currency"))
	place := template.HTMLEscapeString(request.FormValue("place"))

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

	if category != "" ||
		subcategory != "" ||
		method != "" ||
		currency != "" ||
		price != 0 {
		_, err := server.Database.Exec(`
			INSERT INTO 
				scheduledpayments(userid,
				login,
				createdate,
				weekname,
				monthday,
				categoryid,
				subcategoryid,
				price,
				place,
				methodid,
				currencyid)
					VALUES($1, $2, $3, $4, $5,
					(SELECT id FROM categories WHERE type = $6),
					(SELECT id FROM subcategories WHERE type = $7),
					$8, $9,
					(SELECT id FROM methods WHERE type = $10),
					(SELECT id FROM currencies WHERE type = $11))`,
			userid,
			login,
			time.Now(),
			NullStringValue(week),
			NullStringValue(monthday),
			category,
			subcategory,
			price,
			place,
			method,
			currency,
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

		http.Redirect(writer, request, "/scheduled", http.StatusSeeOther)

	} else {
		writer.WriteHeader(http.StatusUnprocessableEntity)
	}
}

func (server *Server) DeleteScheduledPayment(
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

	paymentid := template.HTMLEscapeString(request.FormValue("paymentid"))
	_, err = server.Database.Exec(`
		DELETE FROM scheduledpayments 
			WHERE id = $1 AND userid = $2 AND login = $3`,
		paymentid,
		userid,
		login,
	)
	if err != nil {
		lorg.Error(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.Redirect(writer, request, "/scheduled", http.StatusSeeOther)
}

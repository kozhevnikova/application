package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"time"
)

type CurrencyChart struct {
	CurrencyType
	CurrencyData []CurrencyData
}

type CurrencyData struct {
	Category string `json:"Category"`
	Sum      int    `json:"Sum"`
}

type CurrencyType struct {
	Name string `json:"Currency"`
}

type CategoryChart struct {
	Category
	DailyTransaction []DailyTransaction
}

type Category struct {
	Name string `json:"Category"`
}

type DayExpense struct {
	Day int `json:"Day"`
	Sum int `json:"Sum"`
}

type DailyTransaction struct {
	CurrencyType
	DayExpense []DayExpense
}

type CategoryList struct {
	Category        string
	SubcategoryList []SubcategoryList
}

type SubcategoryList struct {
	Subcategory string
}

func (server *Server) GetCategoriesByPeriod(
	writer http.ResponseWriter,
	request *http.Request,
	period string,

) ([]Category, int, error) {

	var categories []Category

	userid, login, err := ReadCookies(writer, request)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	if period == "week" {
		rows, err := server.Database.Query(
			`SELECT c.type 
				FROM transactions t 
				JOIN categories c ON t.categoryid=c.id
					WHERE userid=$1 and login=$2 AND transactiondate 
					BETWEEN current_date-integer'7' AND current_date 
					GROUP BY c.type`,
			userid,
			login,
		)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}

		for rows.Next() {
			var category Category
			err := rows.Scan(&category.Name)
			if err != nil {
				return nil, http.StatusInternalServerError, err
			}
			categories = append(categories, category)
		}

		err = rows.Err()
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}

		err = rows.Close()
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}
	}

	if period == "month" {
		rows, err := server.Database.Query(
			`SELECT c.type 
				FROM transactions t 
				JOIN categories c ON t.categoryid = c.id
					WHERE userid = $1 AND login = $2 AND transactiondate::date 
					BETWEEN date_trunc('month',NOW()) - '1 month'::interval AND 
					(date_trunc('month', 
					date_trunc('month',
					NOW()) - '1 month'::interval) + 
					interval '1 month - 1 day')::date 
					GROUP BY c.type`,
			userid,
			login,
		)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}

		for rows.Next() {
			var category Category

			err := rows.Scan(&category.Name)
			if err != nil {
				return nil, http.StatusInternalServerError, err
			}

			categories = append(categories, category)
		}

		err = rows.Err()
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}
		err = rows.Close()
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}
	}

	return categories, http.StatusOK, nil
}

func (server *Server) GetDataForCategoryChartLastWeek(
	writer http.ResponseWriter,
	request *http.Request,

) ([]byte, int, error) {

	var categoryChart []CategoryChart

	userid, login, err := ReadCookies(writer, request)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	categories, status, err := server.GetCategoriesByPeriod(
		writer,
		request,
		"week",
	)
	if err != nil {
		return nil, status, err
	}

	currencies, status, err := server.GetCurrencyType()
	if err != nil {
		return nil, status, err
	}

	for _, category := range categories {

		var dailyTransactions []DailyTransaction

		for _, currency := range currencies {

			rows, err := server.Database.Query(`
				SELECT EXTRACT(day from transactiondate) AS day,
					SUM(price::money::numeric::float8) FROM transactions t
					JOIN categories c ON t.categoryid=c.id
					JOIN currencies cu ON t.currencyid=cu.id
						WHERE userid=$1 
							AND login=$2 
							AND c.type=$3 
							AND cu.type=$4 
							AND transactiondate::date 
						BETWEEN current_date-integer'6' 
							AND current_date group by day order by day`,
				userid,
				login,
				category.Name,
				currency.Name,
			)

			if err != nil {
				return nil, http.StatusInternalServerError, err
			}

			var dayExpenses []DayExpense

			for rows.Next() {
				var dayExpense DayExpense

				err := rows.Scan(&dayExpense.Day, &dayExpense.Sum)
				if err != nil {
					return nil, http.StatusInternalServerError, err
				}

				dayExpenses = append(dayExpenses, dayExpense)
			}

			err = rows.Err()
			if err != nil {
				return nil, http.StatusInternalServerError, err
			}

			err = rows.Close()
			if err != nil {
				return nil, http.StatusInternalServerError, err
			}

			withMissedDaysLastWeek := AddMissedDaysForCategoryChartLastWeek(
				dayExpenses)

			dailyTransaction := DailyTransaction{
				CurrencyType: currency,
				DayExpense:   withMissedDaysLastWeek,
			}

			dailyTransactions = append(
				dailyTransactions, dailyTransaction)
		}

		categoryChartData := CategoryChart{
			Category:         category,
			DailyTransaction: dailyTransactions,
		}

		categoryChart = append(categoryChart, categoryChartData)
	}

	buffer := new(bytes.Buffer)

	err = json.NewEncoder(buffer).Encode(categoryChart)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return buffer.Bytes(), http.StatusOK, nil
}

func (server *Server) GetDataForCurrencyChartLastWeek(
	writer http.ResponseWriter,
	request *http.Request,

) ([]byte, int, error) {

	var currenciesChart []CurrencyChart

	userid, login, err := ReadCookies(writer, request)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	currencies, status, err := server.GetCurrencyType()
	if err != nil {
		return nil, status, err
	}

	var currencyChart CurrencyChart

	for _, currency := range currencies {

		rows, err := server.Database.Query(
			`SELECT 
				c.type, SUM(price::money::numeric::float8) 
				FROM transactions t join categories c on t.categoryid=c.id
				JOIN currencies cu on t.currencyid=cu.id
				WHERE cu.type=$1 AND login=$2 AND userid=$3 AND transactiondate 
				BETWEEN current_date-integer'7' AND current_date
				GROUP BY c.type`,
			currency.Name,
			login,
			userid,
		)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}

		var currenciesData []CurrencyData

		for rows.Next() {
			var currencyData CurrencyData

			err := rows.Scan(&currencyData.Category, &currencyData.Sum)
			if err != nil {
				return nil, http.StatusInternalServerError, err
			}

			currenciesData = append(currenciesData, currencyData)
		}

		currencyChart = CurrencyChart{
			CurrencyType: currency,
			CurrencyData: currenciesData,
		}

		currenciesChart = append(
			currenciesChart, currencyChart)
	}

	buffer := new(bytes.Buffer)

	err = json.NewEncoder(buffer).Encode(currenciesChart)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return buffer.Bytes(), http.StatusOK, nil
}

func (server *Server) GetCurrencyType() ([]CurrencyType, int, error) {

	var currenciesType []CurrencyType

	rows, err := server.Database.Query(`SELECT type FROM currencies`)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	for rows.Next() {
		var currencyType CurrencyType

		err := rows.Scan(&currencyType.Name)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}

		currenciesType = append(currenciesType, currencyType)
	}

	return currenciesType, http.StatusOK, nil
}

func AddMissedDaysForCategoryChartLastMonth(
	dayExpenses []DayExpense,

) []DayExpense {

	month, countOfDays := MakeMonthSlice()

	for day := range month {
		for _, value := range dayExpenses {
			if month[day] == value.Day {
				month = append(month[:day], month[day+1:]...)
			}
		}

		dayExpense := DayExpense{
			Day: month[day],
			Sum: 0,
		}

		dayExpenses = append(dayExpenses, dayExpense)

		if len(dayExpenses) == countOfDays {
			break
		}
	}

	sorted := SortMonthForCategoryChart(dayExpenses)

	return sorted
}

func AddMissedDaysForCategoryChartLastWeek(
	dayExpenses []DayExpense,

) []DayExpense {

	week := MakeWeekSlice()

	for day := range week {
		for _, dayExpenseValue := range dayExpenses {
			if week[day] == dayExpenseValue.Day {
				week = append(week[:day], week[day+1:]...)
			}
		}

		dayExpense := DayExpense{
			Day: week[day],
			Sum: 0,
		}

		dayExpenses = append(dayExpenses, dayExpense)

		if len(dayExpenses) == 7 {
			break
		}
	}

	sorted := SortSliceForDayExpense(dayExpenses)

	return sorted
}

func MakeWeekSlice() []int {

	today := time.Now().Day()

	var week []int

	var month time.Month

	if today < 7 {

		if time.Now().Month().String() == "January" {
			month = time.Now().Month() + 11
		} else {
			month = time.Now().Month() - 1
		}

		dayCount := GetDaysNumberInMonth(month.String())
		week = GetWeekSliceForBeginningOfMonth(dayCount, today)

	} else if today >= 7 {
		for y := today - 6; y <= today; y++ {
			week = append(week, y)
		}
	}

	return week
}

func MakeMonthSlice() ([]int, int) {

	var todaymonth time.Month

	if time.Now().Month().String() == "January" {
		todaymonth = time.Now().Month() + 11
	} else {
		todaymonth = time.Now().Month() - 1
	}

	totalCountOfDays := GetDaysNumberInMonth(todaymonth.String())

	var month []int

	for day := 1; day <= totalCountOfDays; day++ {
		month = append(month, day)
	}

	return month, totalCountOfDays
}

func SortSliceForDayExpense(
	dayExpenses []DayExpense,

) []DayExpense {

	sort.SliceStable(dayExpenses, func(i, j int) bool {
		return dayExpenses[i].Day < dayExpenses[j].Day
	})

	return dayExpenses
}

func SortWeek(week []int) []int {
	sort.Ints(week)
	return week
}

func SortMonthForCategoryChart(
	dayExpenses []DayExpense,

) []DayExpense {

	sort.SliceStable(dayExpenses, func(i, j int) bool {
		return dayExpenses[i].Day < dayExpenses[j].Day
	})

	return dayExpenses
}

func GetDaysNumberInMonth(month string) int {

	monthes := []struct {
		Name  string
		Count int
	}{
		{"January", 31},
		{"February", 28},
		{"March", 31},
		{"April", 30},
		{"May", 31},
		{"June", 30},
		{"July", 31},
		{"August", 31},
		{"September", 30},
		{"October", 31},
		{"November", 30},
		{"December", 31},
	}

	if time.Now().Year()%4 == 0 && month == "February" {
		return 29
	}

	count := 0
	for _, index := range monthes {
		if month == index.Name {
			count = index.Count
		}
	}

	return count
}

func GetWeekSliceForBeginningOfMonth(
	count int,
	today int,

) []int {

	additionaldays := 7 - today
	var lastmonthpart []int

	for day := count; day > count-additionaldays; day-- {
		lastmonthpart = append(lastmonthpart, day)
	}

	var week []int

	week = append(week, lastmonthpart...)
	week = SortWeek(week)

	for day := 1; day <= today; day++ {
		week = append(week, day)
	}

	return week
}

func (server *Server) GetDataForCurrencyChartLastMonth(
	writer http.ResponseWriter,
	request *http.Request,

) ([]byte, int, error) {

	var currenciesChart []CurrencyChart

	userid, login, err := ReadCookies(writer, request)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	currencies, status, err := server.GetCurrencyType()
	if err != nil {
		return nil, status, err
	}

	var currencyChart CurrencyChart

	for _, currency := range currencies {

		rows, err := server.Database.Query(
			`SELECT c.type, SUM(price::money::numeric::float8) 
				FROM transactions t 
				JOIN categories c ON t.categoryid = c.id
				JOIN currencies cu ON t.currencyid = cu.id
					WHERE cu.type=$1 
						AND login=$2 
						AND userid=$3 
						AND transactiondate 
					BETWEEN date_trunc('month',NOW()) - '1 month'::interval 
					AND (date_trunc('month', 
						date_trunc('month',NOW()) - '1 month'::interval) 
						+ interval '1 month - 1 day')::date 
					GROUP BY c.type`,
			currency.Name,
			login,
			userid,
		)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}

		var currenciesData []CurrencyData

		for rows.Next() {
			var currencyData CurrencyData

			err := rows.Scan(&currencyData.Category, &currencyData.Sum)
			if err != nil {
				return nil, http.StatusInternalServerError, err
			}

			currenciesData = append(currenciesData, currencyData)
		}

		currencyChart = CurrencyChart{
			CurrencyType: currency,
			CurrencyData: currenciesData,
		}

		currenciesChart = append(currenciesChart, currencyChart)
	}

	buffer := new(bytes.Buffer)

	err = json.NewEncoder(buffer).Encode(currenciesChart)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return buffer.Bytes(), http.StatusOK, nil
}

func (server *Server) GetDataForCategoryChartLastMonth(
	writer http.ResponseWriter,
	request *http.Request,

) ([]CategoryChart, int, error) {

	var categoryCharts []CategoryChart

	userid, login, err := ReadCookies(writer, request)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	categories, status, err := server.GetCategoriesByPeriod(
		writer,
		request,
		"month",
	)
	if err != nil {
		return nil, status, err
	}

	for _, value := range categories {
		if value.Name == "" {
			return nil, http.StatusNoContent, nil
		}
	}

	currencies, status, err := server.GetCurrencyType()
	if err != nil {
		return nil, status, err
	}

	for _, category := range categories {

		var dailyTransactions []DailyTransaction

		for _, currency := range currencies {
			rows, err := server.Database.Query(`
			SELECT 
				EXTRACT(day from transactiondate) AS day,
				SUM(price::money::numeric::float8) FROM transactions t
				JOIN categories c ON t.categoryid = c.id
				JOIN currencies cu ON t.currencyid = cu.id
				WHERE userid=$1 
					AND login=$2 
					AND c.type=$3
					AND cu.type=$4 
					AND transactiondate::date 
					BETWEEN date_trunc('month',NOW()) - '1 month'::interval 
					AND (date_trunc('month', 
						date_trunc('month',NOW()) - '1 month'::interval) +
						interval '1 month - 1 day')::date 
					GROUP BY day 
					ORDER BY day`,
				userid,
				login,
				category.Name,
				currency.Name,
			)

			if err != nil {
				return nil, http.StatusInternalServerError, err
			}

			var dayExpenses []DayExpense

			for rows.Next() {
				var dayExpense DayExpense

				err := rows.Scan(&dayExpense.Day, &dayExpense.Sum)
				if err != nil {
					return nil, http.StatusInternalServerError, err
				}

				dayExpenses = append(dayExpenses, dayExpense)
			}

			withMissedDaysForLastMonth :=
				AddMissedDaysForCategoryChartLastMonth(dayExpenses)

			dailyTransaction := DailyTransaction{
				CurrencyType: currency,
				DayExpense:   withMissedDaysForLastMonth,
			}

			dailyTransactions = append(
				dailyTransactions, dailyTransaction)

			err = rows.Err()
			if err != nil {
				return nil, http.StatusInternalServerError, err
			}

			err = rows.Close()
			if err != nil {
				return nil, http.StatusInternalServerError, err
			}
		}

		categoryChart := CategoryChart{
			Category:         category,
			DailyTransaction: dailyTransactions,
		}

		categoryCharts = append(categoryCharts, categoryChart)
	}

	return categoryCharts, http.StatusOK, nil
}

func (server *Server) GetCategoriesList() ([]CategoryList, error) {

	var categoryLists []CategoryList

	rows, err := server.Database.Query(`SELECT id,type FROM categories`)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	for rows.Next() {
		var categoryList CategoryList
		var id int

		err := rows.Scan(&id, &categoryList.Category)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		subcategories, err := server.GetSubcategoryListForCategory(id)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		categoryList = CategoryList{
			Category:        categoryList.Category,
			SubcategoryList: subcategories,
		}
		categoryLists = append(categoryLists, categoryList)
	}

	err = rows.Err()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	err = rows.Close()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return categoryLists, nil
}

func (server *Server) GetSubcategoryListForCategory(
	categoryid int,

) ([]SubcategoryList, error) {

	var subcategoryLists []SubcategoryList
	rows, err := server.Database.Query(
		`SELECT type FROM subcategories WHERE categoryid=$1`,
		categoryid,
	)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	for rows.Next() {
		var subcategoryList SubcategoryList
		err := rows.Scan(&subcategoryList.Subcategory)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		subcategoryLists = append(subcategoryLists, subcategoryList)
	}

	err = rows.Err()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	err = rows.Close()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return subcategoryLists, nil
}

package main

import (
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"strconv"
)

type Err struct {
	Msg      string `json:"message"`
	StatCode int    `json:"code"`
}

type Resp struct {
	From   string  `json:"from"`
	To     string  `json:"to"`
	Ratio  float64 `json:"exchange-rate"`
	Qty    float64 `json:"quantity"`
	Result float64 `json:"result"`
}

func errorParser(w http.ResponseWriter, errMsg string, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	e := Err{errMsg, code}
	_ = json.NewEncoder(w).Encode(e)
}

func responseParser(w http.ResponseWriter, fromCurrency string, toCurrency string, ratio, quantity, conversion float64) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	resp := Resp{fromCurrency, toCurrency, ratio, quantity, conversion}
	_ = json.NewEncoder(w).Encode(resp)
}

func calcConversion(quantityDecimal, ratio float64) float64 {
	return ratio * quantityDecimal
}

func uriBuilder(fromCurrency, toCurrency string) string {
	// Building the uri with the received query parameters fromCurrency and toCurrency,
	// and quantity is auto-set to 1.0 to get the exchange ratio 1 on 1.
	return "https://currency-exchange.p.rapidapi.com/exchange?from=" + fromCurrency + "&to=" + toCurrency + "&q=1.0"
}

func exchangeRatioRequest(uri string) float64 {

	// Building the request that will be sent to the 3rd party service.
	req, _ := http.NewRequest("GET", uri, nil)

	key := os.Getenv("RAPID_CURRENCY_KEY")
	host := os.Getenv("RAPID_CURRENCY_HOST")

	req.Header.Add("X-RapidAPI-Key", key)
	req.Header.Add("X-RapidAPI-Host", host)

	// Perform the request to the 3rd party service.
	res, _ := http.DefaultClient.Do(req)

	// Decodes the response value representing the rate of exchange between the currencies
	// into a float variable and returns it.
	var ratio float64
	err := json.NewDecoder(res.Body).Decode(&ratio)
	if err != nil {
		fmt.Println("Third party service is not available.")
	}

	return ratio
}

func convert(w http.ResponseWriter, r *http.Request) {

	// Extracting the necessary data from the requests queryString.
	queryString := r.URL.Query()

	fromCurrency := queryString.Get("from")
	toCurrency := queryString.Get("to")

	quantity := queryString.Get("q")
	quantityFloat, err := strconv.ParseFloat(quantity, 64)

	if err != nil {
		errMsg := "Something is wrong with the quantity parameter provided in the queryString."
		errorParser(w, errMsg, http.StatusBadRequest)
	}

	// uri that will be used to perform a request.
	uri := uriBuilder(fromCurrency, toCurrency)

	// The request performed against a 3rd party service to get the exchange rate between 2 currencies.
	ratio := exchangeRatioRequest(uri)

	// Calculating the conversion.
	conversion := calcConversion(quantityFloat, ratio)

	// Parsing the response json that will be sent back to the client.
	responseParser(w, fromCurrency, toCurrency, ratio, quantityFloat, conversion)
}

func requestsHandler() {
	http.HandleFunc("/convert", convert)
}

func loadEnvironment() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalln("Could not load environment variables.")
	}
}

func main() {
	loadEnvironment()
	requestsHandler()
	log.Fatalln(http.ListenAndServe(":8080", nil))
}

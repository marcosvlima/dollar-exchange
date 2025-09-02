package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const ExchangeAPI = "https://economia.awesomeapi.com.br/json/last/USD-BRL"

type Exchange struct {
	USDBRL struct {
		Code       string `json:"code"`
		Codein     string `json:"codein"`
		Name       string `json:"name"`
		High       string `json:"high"`
		Low        string `json:"low"`
		VarBid     string `json:"varBid"`
		PctChange  string `json:"pctChange"`
		Bid        string `json:"bid"`
		Ask        string `json:"ask"`
		Timestamp  string `json:"timestamp"`
		CreateDate string `json:"create_date"`
	} `json:"USDBRL"`
}

func main() {
	db, err := sql.Open("sqlite3", "./data/exchange.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = createTableExchange(db)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/cotacao", func(w http.ResponseWriter, r *http.Request) {
		ExchangeHandler(w, r, db)
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func ExchangeHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	exchange := getExchange()
	if exchange == nil {
		http.Error(w, "Failed to get exchange rate", http.StatusInternalServerError)
		return
	}

	if err := saveExchangeToDB(db, exchange); err != nil {
		log.Println("Error saving exchange to database:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(exchange.USDBRL.Bid)

}

func getExchange() *Exchange {

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*200)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", ExchangeAPI, nil)
	if err != nil {
		log.Println("Error fetching exchange rate:", err)
		return nil
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("Error fetching exchange rate:", err)
		return nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}

	var exchange Exchange
	if err := json.Unmarshal(body, &exchange); err != nil {
		return nil
	}
	return &exchange
}

func saveExchangeToDB(db *sql.DB, exchange *Exchange) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	stmt, err := db.PrepareContext(ctx, `
		INSERT INTO exchange (code, codein, name, high, low, varBid, pctChange, bid, ask, timestamp, create_date)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, exchange.USDBRL.Code, exchange.USDBRL.Codein, exchange.USDBRL.Name, exchange.USDBRL.High, exchange.USDBRL.Low, exchange.USDBRL.VarBid, exchange.USDBRL.PctChange, exchange.USDBRL.Bid, exchange.USDBRL.Ask, exchange.USDBRL.Timestamp, exchange.USDBRL.CreateDate)
	return err
}

func createTableExchange(db *sql.DB) error {
	// Create the exchange table if it doesn't exist
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS exchange (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			code TEXT,
			codein TEXT,
			name TEXT,
			high TEXT,
			low TEXT,
			varBid TEXT,
			pctChange TEXT,
			bid TEXT,
			ask TEXT,
			timestamp TEXT,
			create_date TEXT
		)
	`)
	return err
}

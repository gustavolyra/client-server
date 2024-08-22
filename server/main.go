package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type CurrencyData struct {
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
}

func main() {
	db, err := sql.Open("sqlite3", "./currency_data.db")
	if err != nil {
		panic(err)
	}
	dataBase(db)
	defer db.Close()

	http.HandleFunc("/cotacao", func(w http.ResponseWriter, r *http.Request) {
		handleGet(w, db)
	})
	http.ListenAndServe(":8080", nil)
}

func dataBase(db *sql.DB) {
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS currency_data (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		code TEXT,
		codein TEXT,
		name TEXT,
		high TEXT,
		low TEXT,
		var_bid TEXT,
		pct_change TEXT,
		bid TEXT,
		ask TEXT,
		timestamp TEXT,
		create_date TEXT,
		stored_at DATETIME
	);`

	_, err := db.Exec(createTableSQL)
	if err != nil {
		panic(err)
	}
	fmt.Printf("--DATABASE CREATED--\n")
}

func storeCurrencyData(db *sql.DB, data CurrencyData) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	insertSQL := `
	INSERT INTO currency_data (code, codein, name, high, low, var_bid, pct_change, bid, ask, timestamp, create_date, stored_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := db.ExecContext(ctx, insertSQL, data.Code, data.Codein, data.Name, data.High, data.Low, data.VarBid, data.PctChange, data.Bid, data.Ask, data.Timestamp, data.CreateDate, time.Now())
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("Timeout occurred while storing data: %w", err)
		}
		return fmt.Errorf("Error occurred while storing data: %w", err)
	}

	fmt.Printf("--DATA STORED--\n")
	return nil
}

func handleGet(w http.ResponseWriter, db *sql.DB) {
	usdbrlData, err := getUSDBRL()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(usdbrlData); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := storeCurrencyData(db, usdbrlData); err != nil {
		log.Println("Failed to store data:", err)
	}
}

func getUSDBRL() (CurrencyData, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		return CurrencyData{}, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			fmt.Printf("Request timeout: the server took too long to respond!")
			return CurrencyData{}, ctx.Err()
		} else {
			return CurrencyData{}, err
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return CurrencyData{}, fmt.Errorf(resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return CurrencyData{}, err
	}

	var responseJSON struct {
		USDBRL CurrencyData `json:"USDBRL"`
	}
	err = json.Unmarshal([]byte(body), &responseJSON)
	if err != nil {
		return CurrencyData{}, err
	}
	currencyData := responseJSON.USDBRL

	fmt.Printf("--DATA SEND--\n")
	return currencyData, nil
}

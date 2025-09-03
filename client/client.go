package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type BID struct {
	Value string `json:"bid"`
}

const exchangeEndpoint = "http://localhost:8080/cotacao"

func main() {

	// Get the exchange rate
	res, err := getExchangeRate(exchangeEndpoint)
	if err != nil {
		log.Println("Error getting exchange rate:", err)
		return
	}

	// Write the exchange rate to a file
	file := createFile("cotacao.txt")
	defer file.Close()
	writeFileString("Dólar: "+res, file)

}

func getExchangeRate(url string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*300)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}

	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	res, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	bid := deserialize(res)

	return bid.Value, nil
}

func createFile(nomeArquivo string) *os.File {
	f, err := os.Create(nomeArquivo)
	if err != nil {
		panic(err)
	}
	return f
}

func writeFileString(conteudo string, f *os.File) {
	_, err := f.WriteString(conteudo)
	if err != nil {
		panic(err)
	}

}

func deserialize(jsonBID []byte) BID {
	var bid BID
	err := json.Unmarshal(jsonBID, &bid)
	if err != nil {
		log.Println("Error deserializing JSON:", err)
	}
	return bid
}

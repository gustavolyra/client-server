package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type CotacaoJson struct {
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

	//Defini contexto com timeout de 300ms
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	// Cria a requisição ao servidor
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		panic(err)
	}

	//Envia a requisição
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			fmt.Println("Request timeout: the server took too long to respond!")
		} else {
			panic(err)
		}
	}
	defer resp.Body.Close()

	// Verificar o status da resposta
	if resp.StatusCode != http.StatusOK {
		fmt.Println("Error: received status ", resp.Status)
		body, _ := io.ReadAll(resp.Body)
		fmt.Println("Error: response body ", string(body))
		return
	}

	// Ler o corpo da resposta
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	// Decodificar o JSON
	var cotacaoJson CotacaoJson
	err = json.Unmarshal(body, &cotacaoJson)
	if err != nil {
		panic(err)
	}
	fmt.Println(cotacaoJson)

	var input string
	for {
		fmt.Print("Save data in file y/n: ")
		_, err = fmt.Scanln(&input)
		if err != nil {
			fmt.Println("Error reading input: ", err)
		}

		if input == "y" {
			saveFile(cotacaoJson)
			break
		}
		if input == "n" {
			break
		}
	}
}

func saveFile(cotacaoJson CotacaoJson) {
	file, err := os.Create("cotacao.txt")
	if err != nil {
		fmt.Print("Error creating file: ", err)
		return
	}
	defer file.Close()

	_, err = file.WriteString("Dólar: " + cotacaoJson.Ask)
	if err != nil {
		fmt.Print("Error writing to file: ", err)
		return
	}

	fmt.Println("Data saved successfully!")
}

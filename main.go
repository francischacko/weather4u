package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/francischacko/weather4u/models"
	"github.com/joho/godotenv"
)

func main() {
	atime := time.Now()
	fmt.Println("hello")
	errorCh1 := make(chan error)
	bodyCh1 := make(chan []byte)
	errorCh2 := make(chan error)
	bodyCh2 := make(chan []byte)
	err := godotenv.Load()
	if err != nil {
		panic("error loading env files")
	}
	fmt.Println("enter 2 cities")
	var city1, city2 string
	fmt.Scan(&city1, &city2)
	key := os.Getenv("key")
	url := fmt.Sprintf("http://api.weatherapi.com/v1/forecast.json?key=%s&q=%s&aqi=yes", key, city1)
	url2 := fmt.Sprintf("http://api.weatherapi.com/v1/forecast.json?key=%s&q=%s&aqi=yes", key, city2)

	go func() {
		resp, err := http.Get(url)
		if err != nil {
			log.Fatal("api not responding")
		}
		a, e := io.ReadAll(resp.Body)
		if e != nil {
			errorCh1 <- e
			fmt.Println("error in parsing")
		}
		bodyCh1 <- a
	}()

	a := <-bodyCh1
	var res models.ResponseBody
	err = json.Unmarshal(a, &res)
	if err != nil {
		log.Println(err)
	}

	go func() {
		resp2, err := http.Get(url2)
		if err != nil {
			log.Fatal("api not responding")
		}

		b, e := io.ReadAll(resp2.Body)
		if e != nil {
			errorCh2 <- e
			fmt.Println("error in parsing")
		}
		bodyCh2 <- b
	}()
	b := <-bodyCh2
	var res2 models.ResponseBody
	err = json.Unmarshal(b, &res2)
	if err != nil {
		log.Println(err)
	}
	select {
	case err := <-errorCh1:
		fmt.Println("error from request 1:", err)
	case err := <-errorCh2:
		fmt.Println("erro from request 2:", err)
	default:
		fmt.Printf("Locations(:%v , :%v) \n", res.Location.Name, res2.Location.Name)
		fmt.Printf("Real Feel(%v , %v) \n", res.Current.Realfeel, res2.Current.Realfeel)
		fmt.Printf("PM2.5 (%v µg/m³, %v µg/m³) \n", res.Current.Air.Pm, res2.Current.Air.Pm)

	}

	fmt.Println("time took: ", time.Since(atime))
	reqtime := time.Now()

	prompt := fmt.Sprintf("please  write a summary on air quality between %v and %v, here i'll provide current pm2.5 data of cities %v and %v in µg/m³. make sure the summary take account of the other airquality datas and make a comparison of two.", res.Location.Name, res2.Location.Name, res.Current.Air.Pm, res2.Current.Air.Pm)
	body := models.ChatRequest{
		Messages: []models.Message{
			{Role: "system", Content: "You are a helpful coding assistant."},
			{Role: "user", Content: prompt}, // Role is essential for distinguishing between different types of inputs that the model should treat differently
		},
		Temperature: 0.7, //closer to 2 more random , closer to 0 more deterministic
		MaxTokens:   400, //1 token 4 chars in english, we can set it accordingly and limit the output
		Stream:      true,
	}
	byt, err := json.Marshal(body)
	if err != nil {
		log.Println("error in marshalling")
	}
	url = "http://localhost:1234/v1/chat/completions/"
	client := http.Client{}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(byt))
	if err != nil {
		log.Println("error in making request locally to llama model")
	}
	if req == nil {
		log.Println("this is causing nil pointer error")
	} else {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("error in recieving response from llama model")
	}
	defer resp.Body.Close()
	if resp.Body == nil {
		log.Println("response body is nil")
		return // exit the function or handle the error as appropriate
	}
	c, e := io.ReadAll(resp.Body)
	if e != nil {
		log.Println("error in reading response body: ", e)
	}
	var fullJSON string
	lines := strings.Split(string(c), "\n")
	jsonDataCount := 0 // to count valid JSON data lines

	for _, line := range lines {
		trimmedLine := strings.TrimPrefix(line, "data: ")
		if trimmedLine != line { // Means the prefix was found and trimmed
			trimmedLine = strings.TrimSpace(trimmedLine) // Extra safety, remove any leading/trailing whitespace
			if trimmedLine == "[DONE]" || trimmedLine == "" {
				continue // Skip this line as it's not valid JSON data
			}
			if jsonDataCount > 0 {
				fullJSON += ","
			}
			fullJSON += trimmedLine
			jsonDataCount++
		}
	}

	if jsonDataCount == 0 {
		log.Println("No valid JSON data found")
		return
	}
	fullJSON = "[" + fullJSON + "]"
	var responses []models.ChatCompletionResponse
	er := json.Unmarshal([]byte(fullJSON), &responses)
	if er != nil {
		log.Println("error:", er)
		log.Fatal("error while unmarshalling")
	}

	Docs := []string{}
	for _, response := range responses {
		for _, choice := range response.Choices {
			Docs = append(Docs, choice.Delta.Content)
		}
	}

	concatenated := strings.Join(Docs, " ")

	log.Println("SUMMARY: ", concatenated)
	fmt.Println("time took for prompting:", time.Since(reqtime))
}

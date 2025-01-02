package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

func main() {
	if len(args) < 1 || args[0] != "Weekly" && args[0] != "Daily" {
		fmt.Print("Weekly or Daily Requried\n")
		return
	}
	file, err := os.ReadFile(".env")
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal([]byte(file), &env)
	if err != nil {
		panic(err)
	}
	defer cancel()
	cur, err := RequestWrapper(client, ctx, &cm, "GET", env.BLUE_URL, nil, nil)
	if err != nil {
		panic(err)
	}
	loginBody := map[string]interface{}{
		"identity": env.PB_USER,
		"password": env.PB_PASS,
	}
	apiHeaders := http.Header{}
	apiHeaders.Add("Content-Type", "application/json")
	loginUrl := string(env.PB_URL + env.LOGIN_PATH)
	log, err := RequestWrapper(client, ctx, &lm, "POST", loginUrl, loginBody, apiHeaders)
	if err != nil {
		panic(err)
	}
	apiHeaders.Add("Auhtorization", "Bearer "+log.Token)
	pricingBody := map[string]interface{}{
		"category": "Dollar blue",
		"amount":   cur.Blue.ValueSell,
		"unit":     30,
		"period":   args[0],
	}
	pricingUrl := string(env.PB_URL + env.PRICING_PATH)
	prc, err := RequestWrapper(client, ctx, &pm, "POST", pricingUrl, pricingBody, apiHeaders)
	if err != nil {
		panic(err)
	}
	if prc.Amount > 0 {
		fmt.Printf("New price: %v x %v %s\n", prc.Amount, prc.Unit, prc.Period)
	}
}

const setTimeout = 10 * time.Second

var (
	env         Environment
	cm          Currencies
	lm          Login
	pm          Pricing
	args        = os.Args[1:]
	client      = &http.Client{}
	ctx, cancel = context.WithTimeout(context.Background(), setTimeout)
)

func RequestWrapper[RM any](client *http.Client, ctx context.Context, model *RM, method string, url string, body map[string]interface{}, headers http.Header) (*RM, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	if headers != nil {
		req.Header = headers
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		fmt.Printf("%v", &resp.Body)
		return nil, fmt.Errorf("%s %s %s", resp.Status, resp.Request.Method, resp.Request.URL)
	}
	fmt.Printf("%s %s %s\n", resp.Status, resp.Request.Method, resp.Request.URL)
	rb, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}
	err = json.Unmarshal(rb, model)
	if err != nil {
		fmt.Printf("Error unmarshalling JSON: %s", err)
		return nil, err
	}
	return model, nil
}

type Environment struct {
	BLUE_URL     string `json:"BLUE_URL"`
	PB_URL       string `json:"PB_URL"`
	LOGIN_PATH   string `json:"LOGIN_PATH"`
	PRICING_PATH string `json:"PRICING_PATH"`
	PB_USER      string `json:"PB_USER"`
	PB_PASS      string `json:"PB_PASS"`
}
type Currencies struct {
	Oficial struct {
		ValueAvg  float64 `json:"value_avg"`
		ValueSell float64 `json:"value_sell"`
		ValueBuy  float64 `json:"value_buy"`
	} `json:"oficial"`
	Blue struct {
		ValueAvg  float64 `json:"value_avg"`
		ValueSell float64 `json:"value_sell"`
		ValueBuy  float64 `json:"value_buy"`
	} `json:"blue"`
	OficialEuro struct {
		ValueAvg  float64 `json:"value_avg"`
		ValueSell float64 `json:"value_sell"`
		ValueBuy  float64 `json:"value_buy"`
	} `json:"oficial_euro"`
	BlueEuro struct {
		ValueAvg  float64 `json:"value_avg"`
		ValueSell float64 `json:"value_sell"`
		ValueBuy  float64 `json:"value_buy"`
	} `json:"blue_euro"`
	Time string `json:"last_update"`
}
type Login struct {
	Token  string `json:"token"`
	Record struct {
		CollectionID    string `json:"collectionId"`
		CollectionName  string `json:"collectionName"`
		ID              string `json:"id"`
		Email           string `json:"email"`
		EmailVisibility bool   `json:"emailVisibility"`
		Verified        bool   `json:"verified"`
		Created         string `json:"created"`
		Updated         string `json:"updated"`
	} `json:"record"`
}
type Pricing struct {
	CollectionId   string `json:"collectionId"`
	CollectionName string `json:"collectionName"`
	Id             string `json:"id"`
	Period         string `json:"period"`
	Category       string `json:"category"`
	Amount         int64  `json:"amount"`
	Unit           int64  `json:"unit"`
	Created        string `json:"created"`
	Updated        string `json:"updated"`
}

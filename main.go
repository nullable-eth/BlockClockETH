package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"BlockClockETH/blockClock"
	coingecko2 "BlockClockETH/coingecko"
	"BlockClockETH/coingecko/types"
)

type TokenPrice struct {
	Type             string `json:"type"`
	Currency         string `json:"currency"`
	Symbol           string `json:"symbol"`
	DisplayCurrency  string `json:"display_currency"`
	ContractAddress  string `json:"contract_address"`
	Price            float32
	MarketCap        float32
	Volume           float32
	Change           float32
	LastUpdate       int32
	LightsPriceAbove float32       `json:"light_price_above"`
	LightsPriceBelow float32       `json:"light_price_below"`
	LightsPercent    float32       `json:"light_percent"`
	ShowDuration     time.Duration `json:"show_duration_seconds"`
	Notify           bool          `json:"notify"`
}

type Config struct {
	Sort              bool         `json:"sort_symbols"`
	BlockClockAddress string       `json:"block_clock_address"`
	BlockClockPass    string       `json:"block_clock_password"`
	SMTPServer        string       `json:"smtp_server"`
	SMTPPort          string       `json:"smtp_port"`
	SMTPUser          string       `json:"smtp_user"`
	SMTPPass          string       `json:"smtp_pass"`
	NotifyAddress     string       `json:"notify_address"`
	Tokens            []TokenPrice `json:"tokens"`
}

var config Config
var BC *blockClock.Client

func main() {
	jsonFile, err := os.Open("config.json")
	if err != nil {
		log.Fatal(err)
	}
	jsonBytes, _ := ioutil.ReadAll(jsonFile)
	_ = jsonFile.Close()

	if err := json.Unmarshal(jsonBytes, &config); err != nil {
		log.Fatal(err)
	}

	if config.Sort {
		sort.SliceStable(config.Tokens, func(i, j int) bool {
			return strings.ToLower(config.Tokens[i].Symbol) < strings.ToLower(config.Tokens[j].Symbol)
		})
	}

	var emailClient *blockClock.EmailClient
	if config.SMTPServer != "" && config.SMTPPort != "" && config.SMTPUser != "" && config.SMTPPass != "" && config.NotifyAddress != "" {
		emailClient, _ = blockClock.NewEmailClient(config.SMTPServer, config.SMTPPort, config.SMTPUser, config.SMTPPass)
		emailClient.NotifyAddress = config.NotifyAddress
	}

	httpClient := &http.Client{
		Timeout: time.Second * 10,
	}
	CG := coingecko2.NewClient(httpClient)
	BC = blockClock.NewClient(httpClient, config.BlockClockAddress, config.BlockClockPass)
	_ = BC.PauseBuiltInFunctions()

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		_ = BC.ResumeBuiltInFunctions()
		os.Exit(1)
	}()

	for {
		for _, token := range config.Tokens {
			switch token.Type {
			case "coingecko":
				log.Println("Getting token price:", token.Symbol)
				err := getTokenPrice(CG, &token)
				if err != nil {
					log.Println(err)
				} else {
					go updateBlockClock(emailClient, token)
				}
				time.Sleep(token.ShowDuration * time.Second)
			default:
				log.Println("Error, not implemented:", token)
			}
		}
	}
}

func getTokenPrice(client *coingecko2.Client, token *TokenPrice) error {
	token.ContractAddress = strings.ToLower(token.ContractAddress)
	token.Currency = strings.ToLower(token.Currency)
	var curPrice map[string]map[string]float32
	if token.ContractAddress == types.Ethereum {
		retVal, err := client.SimplePrice([]string{token.ContractAddress}, []string{token.Currency})
		if err != nil {
			return err
		}
		curPrice = *retVal
	} else {
		retVal, err := client.SimpleTokenPrice(types.Ethereum, []string{token.ContractAddress}, []string{token.Currency})
		if err != nil {
			return err
		}
		curPrice = *retVal
	}

	token.Price = curPrice[token.ContractAddress][token.Currency]
	token.MarketCap = curPrice[token.ContractAddress][token.Currency+"_market_cap"]
	token.Volume = curPrice[token.ContractAddress][token.Currency+"_24h_vol"]
	token.Change = curPrice[token.ContractAddress][token.Currency+"_24h_change"]
	token.LastUpdate = int32(curPrice[token.ContractAddress]["last_updated_at"])

	return nil
}

func updateBlockClock(client *blockClock.EmailClient, token TokenPrice) {
	p := message.NewPrinter(language.English)
	format := "%.6f"

	log.Println("Showing token price:", token.Symbol, "-", token.Price)
	var err error
	if (token.LightsPercent > 0 && token.Change > token.LightsPercent) || (token.LightsPriceAbove > 0 && token.Price >= token.LightsPriceAbove) {
		if token.Notify && client != nil {
			client.SendEmail(token.Currency+" "+p.Sprintf(format, token.Price), "ALERT - "+token.Symbol, client.NotifyAddress)
		}
		err = BC.LightsOn("00ff0040") //green
	} else if (token.LightsPercent != 0 && token.Change < -token.LightsPercent) || (token.LightsPriceBelow > 0 && token.Price <= token.LightsPriceBelow) {
		if token.Notify && client != nil {
			client.SendEmail(token.Currency+" "+p.Sprintf(format, token.Price), "ALERT - "+token.Symbol, client.NotifyAddress)
		}
		err = BC.LightsOn("ff000040") //red
	} else {
		err = BC.LightsOff()
	}

	if err != nil {
		log.Println(err)
	}

	price := p.Sprintf(format, token.Price)
	price = price + "0"

	limit := 0
	for x, character := range price {
		if character == ',' || character == '.' {
			continue
		}
		limit++
		if limit == 7 {
			price = price[0:x]
			break
		}
	}

	err = BC.DisplayLargeText(price, false)
	if err != nil {
		log.Println(err)
		return
	}
	time.Sleep(1 * time.Second)
	err = BC.DisplayOUText(6, token.Symbol, token.DisplayCurrency)
	if err != nil {
		log.Println(err)
	}
}

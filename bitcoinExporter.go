package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

//Currency struct
type Currency struct {
	Code        string
	Symbol      string
	Rate        string
	Description string
	Ratefloat   float64 `json:"rate_float"`
}

//Bitcoin Price Index struct
type Bitcoin struct {
	Time struct {
		Update string
	}
	Bpi struct {
		USD Currency
		GBP Currency
		EUR Currency
	}
}

var (
	// Logger
	Info *log.Logger
)

// Logger
func initLogger(infoHandle io.Writer) {
	Info = log.New(infoHandle, "INFO:", log.Ldate|log.Ltime|log.Lshortfile)
}

// GET Bitcoin Price
func getBitcoinPrice(b *Bitcoin) {
	//Making API Call
	resp, err := http.Get("https://api.coindesk.com/v1/bpi/currentprice.json")
	if err != nil {
		fmt.Println("Error with the request")
	}

	//Converting from string to []byte
	apiCall, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("error:", err)
	}

	// Unmarshalling API Call
	err = json.Unmarshal(apiCall, &b)
	if err != nil {
		fmt.Println("error:", err)
	}

	defer resp.Body.Close()
}

func main() {
	//Initialization of the logger
	initLogger(os.Stdout)

	//Creating the GaugeVec with label currency
	bitcoinPrice := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "bitcoinIndex",
		Subsystem: "Coindesk",
		Name:      "bitcoinPrice",
		Help:      "Bitcoin Price Index based on Coindesk API",
	},
		[]string{"currency"})

	//Registering the GaugeVec to the registry
	prometheus.MustRegister(bitcoinPrice)
	//Setting the first price to 0
	bitcoinPrice.WithLabelValues("Euro").Set(0)
	bitcoinPrice.WithLabelValues("USD").Set(0)
	bitcoinPrice.WithLabelValues("GBP").Set(0)

	var b Bitcoin
	getBitcoinPrice(&b)

	//Implement gatherers
	bitcoinPrice.WithLabelValues("Euro").Set(b.Bpi.EUR.Ratefloat)
	bitcoinPrice.WithLabelValues("USD").Set(b.Bpi.USD.Ratefloat)
	bitcoinPrice.WithLabelValues("GBP").Set(b.Bpi.GBP.Ratefloat)
	Info.Println("Setting the Current Rate")

	//Seeting the Handler to gather metrics
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":8080", nil))
}

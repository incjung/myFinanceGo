package main

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"slices"
	"time"
)

type Response struct {
  Chart struct {
    Result []struct {
      Meta       any     `json:"meta"`
      Timestamp  []int64 `json:"timestamp"`
      Indicators struct {
	Quote []struct {
	  Low    []float64 `json:"low"`
	  Volume []float64 `json:"volume"`
	  Open   []float64 `json:"open"`
	  High   []float64 `json:"high"`
	  Close  []float64 `json:"close"`	  
	} `json:"quote"`
	AdjClose []struct {
	  AdjClose []float64 `json:"adjclose"`
	} `json:"adjclose"`
      } `json:"indicators"`
    } `json:"result"`
    Error any `json:"error"`
  } `json:"chart"`
}

type IMeasure interface {
  CalMyMeasures ()
}

type Ticker struct {
  Name string
  Dates      []time.Time
  Indicators map[string][]float64
  Measure IMeasure
}

func NewTicker(symbol string) *Ticker {
  return &Ticker{Name: symbol,    
    Indicators: make(map[string][]float64),
  }
}


func (t *Ticker) getChart () error {
  url := fmt.Sprintf("https://query1.finance.yahoo.com/v8/finance/chart/%s?range=%s&interval=%s", t.Name, srange, interval)
  req, e := http.NewRequest(http.MethodGet, url, nil)
  if e != nil {
    fmt.Println("Cannot do http reqeust!!")
    return e
  }
  req.Header.Set("User-Agent", "Mozilla/5.0")
  res, e := new(http.Client).Do(req)
  if e != nil {
    fmt.Println("Cannot do http reqeust!!!!")
    return e
  }
  defer res.Body.Close()

  myRes := &Response{}
  e = json.NewDecoder(res.Body).Decode(myRes)
  if e != nil {
    fmt.Println("Something wrong during JSON decoding")
    return e
  }

  if len(myRes.Chart.Result) == 0 {
    fmt.Println("Chart Return is 0. Something wrong")
    panic("Something wrong"+url)
  }

  data := myRes.Chart.Result[0]
  
  t.Dates = parseDates(data.Timestamp)

  t.Indicators["open"]= data.Indicators.Quote[0].Open
  t.Indicators["high"]= data.Indicators.Quote[0].High
  t.Indicators["low"]= data.Indicators.Quote[0].Low
  t.Indicators["close"]= data.Indicators.Quote[0].Close

  return nil
}

func (t *Ticker) Print() {
  fmt.Println(t.Name)
  fmt.Println("================================")
  for i, d := range t.Dates {
    fmt.Printf("Date: %v, Open: %v, hight: %v, low: %v, close: %v\n",
      d,
      roundToDecimalPlaces(t.Indicators["open"][i],2),
      roundToDecimalPlaces(t.Indicators["high"][i],2),
      roundToDecimalPlaces(t.Indicators["low"][i],2),
      roundToDecimalPlaces(t.Indicators["close"][i],2),
    )
  }
}

type MyInterestMeasure struct {
  MyTicker *Ticker
  CurrentPrice float64
  LastHight map[string]float64  //last 10d, 30d, 6mo
  DropRate map[string]float64  //last 10d, 30d, 6mo
  Warning bool 
}

func NewMyInterestMeasure(t *Ticker) *MyInterestMeasure {
  return &MyInterestMeasure{
    MyTicker: t,
    CurrentPrice: 0,
    LastHight: make(map[string]float64),
    DropRate: make(map[string]float64),
  }
}

func (mi *MyInterestMeasure) CalMyMeasures () {
  shigh := mi.MyTicker.Indicators["high"]
  mLen := len(shigh)
  mi.CurrentPrice = roundToDecimalPlaces(mi.MyTicker.Indicators["close"][mLen-1],2)
  mi.LastHight["last05d"] = roundToDecimalPlaces(slices.Max(shigh[mLen-5:]),2)
  mi.DropRate["last05d"] = roundToDecimalPlaces((mi.CurrentPrice -  mi.LastHight["last05d"]) / mi.LastHight["last05d"]*100,2)
  mi.LastHight["last10d"] = roundToDecimalPlaces(slices.Max(shigh[mLen-10:]),2)
  mi.DropRate["last10d"] = roundToDecimalPlaces((mi.CurrentPrice -  mi.LastHight["last10d"]) / mi.LastHight["last10d"]*100,2)
  mi.LastHight["last30d"] = roundToDecimalPlaces(slices.Max(shigh[mLen-30:]),2)
  mi.DropRate["last30d"] = roundToDecimalPlaces((mi.CurrentPrice -  mi.LastHight["last30d"]) / mi.LastHight["last30d"]*100,2)
  mi.LastHight["last6mo"] = roundToDecimalPlaces(slices.Max(shigh),2)
  mi.DropRate["last6mo"] = roundToDecimalPlaces((mi.CurrentPrice -  mi.LastHight["last6mo"]) / mi.LastHight["last6mo"]*100,2)

  if mi.LastHight["last6mo"]*0.9 >= mi.CurrentPrice {
    mi.Warning = true
  }   
}

func (mi *MyInterestMeasure) Print() {
  fmt.Printf("|%6v|%7v|%7v|%7v|%7v|%7v|%7v|%7v|%7v|%7v|%7v|\n","NAME","Price","last05d","drop05d","last10d","drop10d","last30d","drop30d","last6mo","drop6mo","sell?")
  fmt.Printf("|%6v|%7v|%7v|%7v|%7v|%7v|%7v|%7v|%7v|%7v|%7v|\n",
    mi.MyTicker.Name,
    mi.CurrentPrice,
    mi.LastHight["last05d"],
    mi.DropRate["last05d"],
    mi.LastHight["last10d"],
    mi.DropRate["last10d"],
    mi.LastHight["last30d"],
    mi.DropRate["last30d"],
    mi.LastHight["last6mo"],
    mi.DropRate["last6mo"],
    mi.Warning,
  )
}


func roundToDecimalPlaces(num float64, places int) float64 {
  shift := math.Pow(10, float64(places))
  return math.Round(num*shift) / shift
}

func parseDates(unixTimes []int64) []time.Time {
  res := make([]time.Time, len(unixTimes))
  for i, t := range unixTimes {
    res[i] = time.Unix(t, 0)
  }
  return res
}

const (
  srange = "6mo" // "1mo"
  interval = "1d"
)

func main() {

  myInterestSymbol := []string{"FSF.NZ","FNZ.NZ", "CEN.NZ"}

  for _, symbol := range myInterestSymbol {
    myTicker := NewTicker(symbol)
    myTicker.getChart()
    //myTicker.Print()

    myMesure := NewMyInterestMeasure(myTicker)
    myMesure.CalMyMeasures()
    myMesure.Print()
  }
  
}


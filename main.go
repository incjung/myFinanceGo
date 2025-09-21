package main

import (
  "encoding/json"
  "fmt"
  "math"
  "net/http"
  "slices"
  "sync"
  "time"
)

// Response struct matches the Yahoo Finance API JSON structure.
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

// Ticker holds the financial data for a stock symbol.
type Ticker struct {
  Name       string
  Dates      []time.Time
  Indicators map[string][]float64
}

// NewTicker creates a new Ticker instance.
func NewTicker(symbol string) *Ticker {
  return &Ticker{
    Name:       symbol,
    Indicators: make(map[string][]float64),
  }
}

// getChart fetches chart data from Yahoo Finance.
// It uses a shared http.Client for efficiency.
func (t *Ticker) getChart(client *http.Client, srange, interval string) error {
  url := fmt.Sprintf("https://query1.finance.yahoo.com/v8/finance/chart/%s?range=%s&interval=%s", t.Name, srange, interval)
  req, err := http.NewRequest(http.MethodGet, url, nil)
  if err != nil {
    return fmt.Errorf("failed to create HTTP request for %s: %w", t.Name, err)
  }
  req.Header.Set("User-Agent", "Mozilla/5.0")

  res, err := client.Do(req)
  if err != nil {
    return fmt.Errorf("failed to execute HTTP request for %s: %w", t.Name, err)
  }
  defer res.Body.Close()

  myRes := &Response{}
  if err := json.NewDecoder(res.Body).Decode(myRes); err != nil {
    return fmt.Errorf("failed to decode JSON for %s: %w", t.Name, err)
  }

  if len(myRes.Chart.Result) == 0 || len(myRes.Chart.Result[0].Timestamp) == 0 {
    return fmt.Errorf("no data returned for symbol %s from URL %s", t.Name, url)
  }

  data := myRes.Chart.Result[0]
  t.Dates = parseDates(data.Timestamp)
  t.Indicators["open"] = data.Indicators.Quote[0].Open
  t.Indicators["high"] = data.Indicators.Quote[0].High
  t.Indicators["low"] = data.Indicators.Quote[0].Low
  t.Indicators["close"] = data.Indicators.Quote[0].Close

  return nil
}

// MyInterestMeasure holds the calculated metrics for a Ticker.
type MyInterestMeasure struct {
  MyTicker         *Ticker
  MyInterestedHeader []string
  CurrentPrice     float64
  LastHigh         map[string]float64 // Renamed from LastHight
  DropRate         map[string]float64
  Warning          bool
}

// DoCalculate performs the analysis on the Ticker data.
func (t *Ticker) DoCalculate() *MyInterestMeasure {
  myInterest := &MyInterestMeasure{
    MyTicker:     t,
    CurrentPrice: 0,
    LastHigh:     make(map[string]float64),
    DropRate:     make(map[string]float64),
  }
  myInterest.CalMyMeasures()
  return myInterest
}

// CalMyMeasures calculates the custom metrics based on the ticker data.
func (mi *MyInterestMeasure) CalMyMeasures() {
  highs := mi.MyTicker.Indicators["high"]
  closes := mi.MyTicker.Indicators["close"]
  mLen := len(highs)
  if mLen == 0 {
    return // No data to process
  }

  mi.MyInterestedHeader = []string{"NAME", "Price", "last05d", "drop05d", "last10d", "drop10d", "last30d", "drop30d", "last6mo", "drop6mo", "sell?"}
  mi.CurrentPrice = roundToDecimalPlaces(closes[mLen-1], 2)

  // Calculate metrics for different periods
  mi.calculatePeriodMetric("last05d", 5)
  mi.calculatePeriodMetric("last10d", 10)
  mi.calculatePeriodMetric("last30d", 30)
  mi.calculatePeriodMetric("last6mo", 180) // Approximate 6 months

  // Warning logic
  if sixMonthHigh, ok := mi.LastHigh["last6mo"]; ok && sixMonthHigh*0.9 >= mi.CurrentPrice {
    mi.Warning = true
  }
}

// calculatePeriodMetric is a helper to compute high and drop rate for a given period.
func (mi *MyInterestMeasure) calculatePeriodMetric(key string, days int) {
  highs := mi.MyTicker.Indicators["high"]
  mLen := len(highs)

  // Safety check: ensure we have enough data for the period. Use max to avoid negative index.
  start := max(0, mLen-days)

  periodHighs := highs[start:]
  if len(periodHighs) == 0 {
    return // No data for this period
  }

  lastHigh := roundToDecimalPlaces(slices.Max(periodHighs), 2)
  mi.LastHigh[key] = lastHigh
  if lastHigh > 0 {
    dropRate := (mi.CurrentPrice - lastHigh) / lastHigh * 100
    mi.DropRate[key] = roundToDecimalPlaces(dropRate, 2)
  }
}


// PrintHeader prints the header for the results table.
func (mi *MyInterestMeasure) PrintHeader() {
  fmt.Printf("|%6v|", mi.MyInterestedHeader[0])
  for j := 1; j < len(mi.MyInterestedHeader); j++ {
    fmt.Printf("%7v|", mi.MyInterestedHeader[j])
  }
  fmt.Println()
}

// PrintData prints the data row for a single stock.
func (mi *MyInterestMeasure) PrintData() {
  fmt.Printf("|%6v|%7.2f|%7.2f|%7.2f|%7.2f|%7.2f|%7.2f|%7.2f|%7.2f|%7.2f|%7v|\n",
    mi.MyTicker.Name,
    mi.CurrentPrice,
    mi.LastHigh["last05d"],
    mi.DropRate["last05d"],
    mi.LastHigh["last10d"],
    mi.DropRate["last10d"],
    mi.LastHigh["last30d"],
    mi.DropRate["last30d"],
    mi.LastHigh["last6mo"],
    mi.DropRate["last6mo"],
    mi.Warning,
  )
}

// roundToDecimalPlaces utility function.
func roundToDecimalPlaces(num float64, places int) float64 {
  shift := math.Pow(10, float64(places))
  return math.Round(num*shift) / shift
}

// parseDates utility function to convert unix timestamps to time.Time.
func parseDates(unixTimes []int64) []time.Time {
  res := make([]time.Time, len(unixTimes))
  for i, t := range unixTimes {
    res[i] = time.Unix(t, 0)
  }
  return res
}

const (
  srange   = "6mo"
  interval = "1d"
)

func main() {
  myInterestSymbols := []string{"FNZ.NZ", "CEN.NZ","FSF.NZ"}

  // Create a single, reusable HTTP client.
  client := &http.Client{Timeout: 10 * time.Second}

  // Use channels to handle results and errors from concurrent goroutines.
  measuresChan := make(chan *MyInterestMeasure, len(myInterestSymbols))
  errChan := make(chan error, len(myInterestSymbols))
	
  var wg sync.WaitGroup

  

  // Fetch data for each symbol concurrently.
  for _, symbol := range myInterestSymbols {
    wg.Add(1)
    go func(sym string) {
      defer wg.Done()
			
      myTicker := NewTicker(sym)
      if err := myTicker.getChart(client, srange, interval); err != nil {
	errChan <- err
	return
      }
			
      measuresChan <- myTicker.DoCalculate()
    }(symbol)
  }

  // Wait for all goroutines to finish, then close the channels.
  go func() {
    wg.Wait()
    close(measuresChan)
    close(errChan)
  }()

  // Process results and errors as they come in.
  var results []*MyInterestMeasure
  for {
    select {
    case measure, ok := <-measuresChan:
      if !ok {
	measuresChan = nil // Channel is closed and drained
      } else {
	results = append(results, measure)
      }
    case err, ok := <-errChan:
      if !ok {
	errChan = nil // Channel is closed and drained
      } else {
	fmt.Printf("Error: %v\n", err)
      }
    }
    // Exit loop when both channels are closed.
    if measuresChan == nil && errChan == nil {
      break
    }
  }

  // Print the collected results.
  if len(results) > 0 {
    //Sort results to maintain a consistent order
    slices.SortFunc(results, func(a, b *MyInterestMeasure) int {
      if a.MyTicker.Name < b.MyTicker.Name {
	return -1
      }
      if a.MyTicker.Name > b.MyTicker.Name {
	return 1
      }
      return 0
    })

    results[0].PrintHeader()
    for _, t := range results {
      t.PrintData()
    }
  } else {
    fmt.Println("No data to display.")
  }
}

# Go Finance Tracker

A simple command-line tool written in Go to track stock prices and analyze performance based on historical data from Yahoo Finance.

## Description

This application fetches the last 6 months of daily stock data for a predefined list of symbols. It then calculates several metrics, including the highest price over various recent periods (5-day, 10-day, 30-day, 6-month) and the percentage drop from those highs to the current price. It displays the results in a clean, formatted table in the console and flags stocks that have dropped more than 10% from their 6-month high.

The application is built for efficiency, fetching data for all symbols concurrently.

## Features

- Fetches historical stock data from Yahoo Finance.
- Calculates price drop percentages from 5-day, 10-day, 30-day, and 6-month highs.
- Provides a simple warning flag for stocks that have dropped significantly.
- Concurrent data fetching to quickly analyze multiple stocks.
- Clear, tabular console output.

## Requirements

- Go version 1.21 or later.

## How to Run

1. Clone or download the project.
2. Navigate to the project directory.
3. Run the application from your terminal:

    ```bash
    go run main.go
    ```

### Example Output

```
Fetching data for symbols...
--- Analysis Complete ---
|  NAME|  Price|last05d|drop05d|last10d|drop10d|last30d|drop30d|last6mo|drop6mo|  sell?|
|CEN.NZ|   2.55|   2.59|  -1.54|   2.64|  -3.41|   2.88| -11.46|   3.06| -16.67|   true|
|FNZ.NZ|   1.90|   1.91|  -0.52|   1.91|  -0.52|   1.95|  -2.56|   2.04|  -6.86|  false|
|FSF.NZ|   0.68|   0.68|   0.00|   0.69|  -1.45|   0.70|  -2.86|   0.79| -13.92|   true|
|SPK.NZ|   4.44|   4.44|   0.00|   4.52|  -1.77|   4.52|  -1.77|   4.92|  -9.76|  false|
|VOD.NZ|   0.71|   0.72|  -1.39|   0.73|  -2.74|   0.76|  -6.58|   0.84| -15.48|   true|
```

## How to Customize

To change the list of stocks to be analyzed, simply edit the `myInterestSymbols` string slice in the `main()` function within `main.go`:

```go
// main.go

func main() {
    // Add or remove Yahoo Finance ticker symbols here
	myInterestSymbols := []string{"FSF.NZ", "FNZ.NZ", "CEN.NZ", "SPK.NZ", "VOD.NZ", "AAPL", "GOOG"}

    // ... rest of the code
}
```

## License

This project is open-source. Feel free to use and modify it. Consider adding a license like MIT if you plan to distribute it.

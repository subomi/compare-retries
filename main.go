package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/sethvargo/go-retry"
)

type strategy struct {
	name  string
	retry retry.Backoff
}

var (
	// gnu retries file for plotting
	filename = "retries.txt"

	// base file headings
	baseFileHeading = "# Format: Attempts"

	// flags
	attempts = 20

	// table sub headings
	subheadingColumns = []string{"Duration", "Cumulative Duration"}

	// algorithms
	strategies = []strategy{
		strategy{
			name:  "Linear Retry",
			retry: retry.NewConstant(1 * time.Hour),
		},
		strategy{
			name:  "Exponential Backoff",
			retry: retry.NewExponential(20 * time.Second),
		},
		strategy{
			name: "Capped Duration",
			retry: func() retry.Backoff {
				ex := retry.NewExponential(20 * time.Second)
				jitter := retry.WithJitterPercent(10, ex)
				return retry.WithCappedDuration(7200*time.Second, jitter)
			}(),
		},
	}
)

func main() {

	// Open file with write permissions, create if not exists, truncate if exists
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		panic(fmt.Errorf("error opening file: %v", err))
	}
	defer file.Close()

	// Create a buffered writer for better performance
	writer := bufio.NewWriter(file)

	// Create a new table writer using os.Stdout
	table := tablewriter.NewWriter(os.Stdout)

	// Set heading
	table.SetHeader(generateHeading([]string{""}, getNames(strategies)))

	// Write file heading
	writer.WriteString(generateFileHeading(baseFileHeading, getNames(strategies)))

	// Set subheader
	table.Append(generateSubHeading([]string{"Attempts"}, len(strategies)))

	// Set table settings
	table.SetRowLine(true)

	var rows [][]string
	var line string
	for attempt := 0; attempt < attempts; attempt++ {
		row := []string{strconv.Itoa(attempt)}

		cells := generateCells([]string{}, rows, attempt, 0)

		line = fmt.Sprintf("%d", attempt)
		writer.WriteString(generateLine(line, cells))
		rows = append(rows, append(row, cells...))
	}

	// Flush the buffer to ensure all data is written to file
	err = writer.Flush()
	if err != nil {
		panic(fmt.Errorf("error flushing buffer: %v", err))
	}

	table.AppendBulk(rows)
	table.Render()
}

func generateHeading(heading []string, strategies []string) []string {
	if len(strategies) == 0 {
		return heading
	}

	heading = append(heading, strategies[0], "")
	return generateHeading(heading, strategies[1:])
}

func generateSubHeading(subheading []string, count int) []string {
	if count == 0 {
		return subheading
	}

	subheading = append(subheading, subheadingColumns...)
	return generateSubHeading(subheading, count-1)
}

func generateCells(cells []string, rows [][]string, attempt int, idx int) []string {
	if len(strategies) == idx {
		return cells
	}

	var duration, cumulativeDuration time.Duration
	duration, _ = strategies[idx].retry.Next()
	if attempt == 0 {
		cumulativeDuration = duration
	} else {
		cell := rows[attempt-1][2*(1+idx)]
		previousCumulativeDuration, err := time.ParseDuration(cell)
		if err != nil {
			panic(err)
		}

		cumulativeDuration = duration + previousCumulativeDuration
	}

	cells = append(cells, duration.String(), cumulativeDuration.String())

	return generateCells(cells, rows, attempt, idx+1)
}

func generateLine(line string, cells []string) string {
	if len(cells) == 0 {
		line = fmt.Sprintf("%s \n", line)
		return line
	}

	val, err := time.ParseDuration(cells[1])
	if err != nil {
		panic(err)
	}

	line = fmt.Sprintf("%s %d", line, int(val.Seconds()))
	return generateLine(line, cells[2:])
}

func generateFileHeading(line string, names []string) string {
	if len(names) == 0 {
		line = fmt.Sprintf("%s \n", line)
		return line
	}

	line = fmt.Sprintf("%s %s", line, names[0])

	return generateFileHeading(line, names[1:])
}

func getNames(items []strategy) []string {
	names := make([]string, len(items))
	for i, item := range items {
		names[i] = item.name
	}
	return names
}

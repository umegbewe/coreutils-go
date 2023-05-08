package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

func main() {
	year := time.Now().Year()
	month := int(time.Now().Month())

	if len(os.Args) == 2 {
		m, err := strconv.Atoi(os.Args[1])
		if err != nil || m < 1 || m > 12 {
			fmt.Println("cal: invalid month argument")
			os.Exit(1)
		}
		month = m
	} else if len(os.Args) == 3 {
		m, err := strconv.Atoi(os.Args[1])
		if err != nil || m < 1 || m > 12 {
			fmt.Println("cal: invalid month argument")
			os.Exit(1)
		}
		month = m

		y, err := strconv.Atoi(os.Args[2])
		if err != nil || y < 1 {
			fmt.Println("cal: invalid year argument")
			os.Exit(1)
		}
		year = y
	} else if len(os.Args) > 3 {
		fmt.Println("cal: too many arguments")
		os.Exit(1)
	}

	printCalendar(month, year)
}

func printCalendar(month, year int) {
	startOfMonth := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, -1)
	daysInMonth := endOfMonth.Day()

	fmt.Printf("    %s %d\n", startOfMonth.Month(), startOfMonth.Year())
	fmt.Println("Su Mo Tu We Th Fr Sa")

	day := 1
	for i := 0; i < int(startOfMonth.Weekday()); i++ {
		fmt.Print("   ")
	}

	for day <= daysInMonth {
		remainingDays := 7 - int(startOfMonth.Weekday())
		for i := 0; i < remainingDays && day <= daysInMonth; i++ {
			fmt.Printf("%2d ", day)
			day++
		}

		fmt.Println()

		startOfMonth = startOfMonth.AddDate(0, 0, remainingDays)
	}
}

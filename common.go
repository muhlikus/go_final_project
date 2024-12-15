package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func parseDate(t string) (time.Time, error) {
	parsedTime, err := time.Parse(dateFormat, t)
	if err != nil {
		return parsedTime, fmt.Errorf("error parsing date: %w", err)
	}
	return parsedTime, nil
}

func NextDate(now time.Time, date string, repeat string) (string, error) {

	parsedDate, err := parseDate(date)
	if err != nil {
		return "", err
	}

	parsedRepeat, err := parseRepeat(repeat)
	//log.Println(repeat, parsedRepeat)
	if err != nil {
		return "", err
	}

	var newDate string

	switch parsedRepeat.key {
	case "y":
		newDate = addYear(now, parsedDate)
	case "d":
		newDate = addDays(now, parsedDate, parsedRepeat.interval)
	}

	return newDate, nil
}

func parseRepeat(repeat string) (repeatSettings, error) {

	r := repeatSettings{}

	repeat = strings.TrimSpace(repeat)

	if repeat == "" {
		return r, errors.New("is empty")
	}

	repeatParts := strings.Fields(repeat)

	switch repeatParts[0] {
	case "y":
		r.key = "y"
		return r, nil
	case "d":
		if len(repeatParts) == 1 {
			return r, fmt.Errorf("undefined interval: %s", repeat)
		}

		offset, err := strconv.ParseInt(repeatParts[1], 10, 16)
		if err != nil {
			return r, fmt.Errorf("unxpected symbol %s in repeat string: %s", repeatParts[1], repeat)
		}

		if !(1 <= offset && offset <= maxDays) /* offset < 0 || maxDays < offset */ {
			return r, fmt.Errorf("interval %d is our of range [%d:%d] in repeat string: %s", offset, 1, maxDays, repeat)
		}

		r.key = "d"
		r.interval = int(offset)
		return r, nil

	default:
		return r, fmt.Errorf("unexpected repeat string: %s", repeat)
	}
}

func addYear(now, date time.Time) string {
	for {
		date = date.AddDate(1, 0, 0)
		if date.After(now) {
			break
		}
	}
	return date.Format(dateFormat)
}

func addDays(now, date time.Time, days int) string {
	for {
		date = date.AddDate(0, 0, days)
		if date.After(now) {
			break
		}
	}
	return date.Format(dateFormat)
}

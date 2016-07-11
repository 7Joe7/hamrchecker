package main

import (
	"log"
	"strconv"
	"strings"
	"fmt"
	"time"
	"errors"
)

func (s *search) Description() string {
	return fmt.Sprintf("%d half hours on %s between %s and %s started at %s.", s.Length, s.From.Format("2.1.2006"), s.From.Format("15:04"), s.Till.Format("15:04"), s.Start.Format("15:04 2.1.2006"))
}

func (s *search) parse(email, dateS, fromS, tillS, lengthS string) error {
	var date *time.Time
	if email == "" {
		return errors.New("E-mail not provided.")
	}
	s.Emails = []string{email}

	if dateS == "" {
		return errors.New("Date not provided.")
	}
	d, err := time.Parse("2006-01-02", dateS)
	if err != nil {
		return fmt.Errorf("Unable to parse date. %v", err)
	}
	if d.Before(time.Now()) {
		return fmt.Errorf("Date is in the past.")
	}
	date = &d

	if fromS == "" {
		return errors.New("From time not provided.")
	}
	if !strings.Contains(fromS, ":00") && !strings.Contains(fromS, ":30") {
		return errors.New("From time should be divisible by half hour.")
	}
	from, err := time.Parse("January 2 2006 15:04", fmt.Sprintf("%s %d %d %s", date.Month().String(), date.Day(), date.Year(), fromS))
	if err != nil {
		return fmt.Errorf("Unable to parse from time. %v", err)
	}
	if from.Before(time.Now()) {
		return fmt.Errorf("From time is in the past.")
	}
	s.From = &from

	if tillS == "" {
		return errors.New("Till time not provided.")
	}
	if !strings.Contains(tillS, ":00") && !strings.Contains(tillS, ":30") {
		return errors.New("Till time should be divisible by half hour.")
	}
	till, err := time.Parse("January 2 2006 15:04", fmt.Sprintf("%s %d %d %s", date.Month().String(), date.Day(), date.Year(), tillS))
	if err != nil {
		return fmt.Errorf("Unable to parse till time. %v", err)
	}
	if till.Before(from) {
		return fmt.Errorf("Till time is before from time.")
	}
	s.Till = &till

	if lengthS == "" {
		return errors.New("Number of consecutive half hours not provided.")
	}
	length, err := strconv.Atoi(lengthS)
	if err != nil {
		return fmt.Errorf("Unable to parse length. %v", err)
	}
	if length < 1 || length > 8 {
		return fmt.Errorf("Number of half hours is out of bounds <1,8>.")
	}
	s.Length = length

	log.Printf("Parsed data: %v", s)
	return nil
}

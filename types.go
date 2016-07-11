package main

import "time"

type Resp struct {
	Div Div `xml:"div"`
}

type Div struct {
	Table Table `xml:"table"`
}

type Table struct {
	Rows []Row `xml:"tr"`
}

type Row struct {
	Data  []*Data `xml:"td"`
	Class string  `xml:"class,attr"`
}

type Data struct {
	Class    string    `xml:"class,attr"`
	Id       string    `xml:"id,attr"`
	InnerDiv *InnerDiv `xml:"div"`
}

type InnerDiv struct {
	Class string `xml:"class,attr"`
}

type search struct {
	Emails []string   `json:"emails,attr"`
	From   *time.Time `json:"from,attr"`
	Till   *time.Time `json:"till,attr"`
	Length int        `json:"length,attr"`
	Start  *time.Time `json:"start,attr"`
}

type templateInfo struct {
	Searches  []*search
	Flash     string
	FlashType string
}

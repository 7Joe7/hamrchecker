package resources

import (
	"time"
	"strings"
	"net/http"
	"log"
	"strconv"
	"fmt"
	"errors"
)

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

func (data *Data) IsFree() bool {
	return strings.Contains(data.InnerDiv.Class, "rgs-free")
}

type InnerDiv struct {
	Class string `xml:"class,attr"`
}

type Search struct {
	Emails []string      `json:"emails,attr"`
	Sport  string
	Place  string
	From   *time.Time    `json:"from,attr"`
	Till   *time.Time    `json:"till,attr"`
	Length int           `json:"length,attr"`
	Start  *time.Time    `json:"start,attr"`
	Id     int           `json:"id,attr"`
	Session *http.Cookie `json:"-"`
}

func (s *Search) Description() string {
	if s.Start != nil {
		return fmt.Sprintf("%s/%s %d half hours on %s between %s and %s started at %s.", PlaceIdToName(s.Place), SportIdToName(s.Sport), s.Length, s.From.Format("2.1.2006"), s.From.Format("15:04"), s.Till.Format("15:04"), s.Start.Format("15:04 2.1.2006"))
	}
	return fmt.Sprintf("%s/%s %d half hours on %s between %s and %s runs.", PlaceIdToName(s.Place), SportIdToName(s.Sport), s.Length, s.From.Format("2.1.2006"), s.From.Format("15:04"), s.Till.Format("15:04"))
}

func (s *Search) Parse(email, dateS, fromS, tillS, lengthS, place, sport string) error {
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

	if place == "" {
		return errors.New("Place can't be empty.")
	}
	if sport == "" {
		return errors.New("Sport can't be empty.")
	}
	switch place {
	case HAMR_PLACE_CODE_BRANIK:
		switch sport {
		case HAMR_SPORT_CODE_BADMINTON, HAMR_SPORT_CODE_BEACH_VOLLEY, HAMR_SPORT_CODE_TENIS,
			HAMR_SPORT_CODE_FLORBAL, HAMR_SPORT_CODE_FOOTBALL:
		default:
			return errors.New("Sport unknown for this place.")
		}
	case HAMR_PLACE_CODE_STERBOHOLY:
		switch sport {
		case HAMR_SPORT_CODE_BADMINTON, HAMR_SPORT_CODE_TENIS, HAMR_SPORT_CODE_FOOTBALL:
		default:
			return errors.New("Sport unknown for this place.")
		}
	case HAMR_PLACE_CODE_ZABEHLICE:
		switch sport {
		case HAMR_SPORT_CODE_BEACH_VOLLEY, HAMR_SPORT_CODE_TENIS, HAMR_SPORT_CODE_SQUASH,
			HAMR_SPORT_CODE_TABLE_TENNIS:
		default:
			return errors.New("Sport unknown for this place.")
		}
	default:
		return errors.New("Unknown place selected.")
	}
	s.Place = place
	s.Sport = sport

	log.Printf("Parsed data: %v", s)
	return nil
}

type TemplateInfo struct {
	Searches  []*Search
	Flash     string
	FlashType string
}

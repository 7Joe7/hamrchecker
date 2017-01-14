package checker

import (
	"strings"
	"strconv"
	"fmt"
	"time"
	"encoding/xml"
	"unicode"
	"net/smtp"
	"log"

	"github.com/7joe7/hamrchecker/resources"
	"github.com/7joe7/hamrchecker/db"
	"github.com/7joe7/hamrchecker/hamrui"
)

var (
	emailConf *resources.EmailConf
)

func setEmailConfiguration(ec *resources.EmailConf) {
	emailConf = ec
}

func checkFreePlace(s *resources.Search, table *resources.Table) bool {
	beginningIndex := convertTimeToIndex(s.From.Format("15:04"))
	log.Printf("Beginning index %d", beginningIndex)
	date := s.From.Format("2006-01-02")
	halfHoursToSearch := int(s.Till.Sub(*s.From).Minutes() / 30)
	dateIndex := CalculateDateIndex(date, time.Now().Add(time.Hour).Format("2006-01-02"))
	log.Printf("Date index %d", dateIndex)
	freeInARow := 0
	for k := beginningIndex; k < beginningIndex + halfHoursToSearch; k++ {
		if k > 33 {
			break
		}
		if table.Rows[dateIndex].Data[k].IsFree() {
			freeInARow += 1
		} else {
			freeInARow = 0
		}
		if freeInARow == s.Length {
			message := fmt.Sprintf(
				"To: You\nSubject: Court is free\nHello,\n\ncourt you have requested at Hamr Sport %s is avaiable.\n\nSearch:\nPlace %s\nSport %s\nDate %s\nFrom %s\n\nHave a nice day.\n\nJOT", resources.PlaceIdToName(s.Place), resources.PlaceIdToName(s.Place), resources.SportIdToName(s.Sport), date, convertIndexToTime(k - s.Length + 1))
			if err := smtp.SendMail(emailConf.SmtpServerWithPort, smtp.PlainAuth("", emailConf.Address, emailConf.Password, emailConf.SmtpServer), emailConf.Address, s.Emails, []byte(message)); err != nil {
				log.Printf("Unable to send e-mail. %v", err)
			}
			err := db.RemoveSearch(s.Id)
			if err != nil {
				log.Printf("Unable to remove search %v. %v", s, err)
			}
			return true
		}
	}
	return false
}

func processResponse(resp string) *resources.Table {
	resp = resp[strings.Index(resp, "<table class=\"rg-table\" id=\"rgTable\" border=\"0\">"):]
	resp = resp[:strings.Index(resp, "</table>") + 8]
	resp = strings.Replace(resp, "&nbsp;", "", -1)
	resp = strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			if r == ' ' {
				return r
			}
			return -1
		}
		return r
	}, resp)
	v := &resources.Table{}
	if err := xml.Unmarshal([]byte(resp), &v); err != nil {
		log.Printf("Unable to process response. %v", err)
		return nil
	}
	return v
}

func CalculateDateIndex(date string, nowS string) int {
	now, _ := time.Parse(resources.TIME_FORMAT, nowS)
	t, _ := time.Parse(resources.TIME_FORMAT, date)
	duration := t.Sub(now)
	return int(duration.Hours()/24) + 1
}

func convertTimeToIndex(time string) int {
	if len(time) == 4 {
		time = fmt.Sprintf("0%s", time)
	}
	i, err := strconv.Atoi(time[0:2])
	if err != nil {
		log.Printf("Unable to convert time to index. %v", err)
		return -1
	}
	index := 2*i - 14
	if strings.HasSuffix(time, ":30") {
		index += 1
	}
	return index + 1
}

func convertIndexToTime(index int) string {
	hour := index + 14 - 1
	if hour % 2 == 1 {
		return fmt.Sprintf("%d:30", hour / 2)
	}
	return fmt.Sprintf("%d:00", hour / 2)
}

func runSearch(s *resources.Search) {
	log.Printf("Starting search %v", s)
	s.Start = resources.GetTimePointer(time.Now())
	ensureAdminEmail(s)
	var err error
	var tryCount int
	if s.Session == nil {
		for {
			err = prepareSession(s)
			if err == nil {
				break
			}
			log.Printf("Unable to prepare a session. %v", err)
			tryCount++
			time.Sleep(time.Duration((2^tryCount)) * time.Second)
		}
	}
	first := true
	for {
		if first {
			first = false
		} else {
			time.Sleep(60 * time.Second)
		}
		data, err := hamrui.GetData(s.Session)
		if err != nil {
			log.Printf("Unable to retrieve data. %v", err)
			continue
		}
		processedResponse := processResponse(data)
		if checkFreePlace(s, processedResponse) {
			log.Printf("Search finished successfully %v.", s)
			break
		}
	}
}

func prepareSession(s *resources.Search) error {
	var err error
	s.Session, err = hamrui.GetSession()
	if err != nil {
		return err
	}
	err = hamrui.ChangeSessionFocus(s.Session, s.Place, s.Sport, resources.SESSION_FOCUS_CHANGE_TYPE_PLACE)
	if err != nil {
		return err
	}
	err = hamrui.ChangeSessionFocus(s.Session, s.Place, s.Sport, resources.SESSION_FOCUS_CHANGE_TYPE_SPORT)
	if err != nil {
		return err
	}
	return nil
}

func ensureAdminEmail(s *resources.Search) {
	found := false
	for i := 0; i < len(s.Emails); i++ {
		if s.Emails[i] == emailConf.Address {
			found = true
			break
		}
	}
	if !found {
		s.Emails = append(s.Emails, emailConf.Address)
	}
}
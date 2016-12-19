package main

import (
	"strings"
	"strconv"
	"fmt"
	"time"
	"encoding/xml"
	"unicode"
	"regexp"
	"os/exec"
	"net/smtp"
	"log"
)

func init() {
	requestParams = []string{"http://hodiny.hamrsport.cz/Login.aspx",
		"-H", "Cookie: _ga=GA1.2.1761385536.1450711716; HamrOnline$SessionId=snt0dk45m2pr0obidq13hpe2; __utmt=1; __utma=74282507.1761385536.1450711716.1460970159.1461009504.4; __utmb=74282507.8.10.1461009504; __utmc=74282507; __utmz=74282507.1461009504.4.4.utmcsr=google|utmccn=(organic)|utmcmd=organic|utmctr=(not%20provided)",
		"-H", "Origin: http://hodiny.hamrsport.cz",
		"-H", "Accept-Encoding: gzip, deflate",
		"-H", "Accept-Language: en-US,en;q=0.8,cs;q=0.6",
		"-H", "User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/49.0.2623.112 Safari/537.36",
		"-H", "Content-Type: application/x-www-form-urlencoded; charset=UTF-8",
		"-H", "Accept: */*",
		"-H", "Cache-Control: no-cache",
		"-H", "X-Requested-With: XMLHttpRequest",
		"-H", "Connection: keep-alive",
		"-H", "X-MicrosoftAjax: Delta=true",
		"-H", "Referer: http://hodiny.hamrsport.cz/Login.aspx",
		"--data", "ctl00%24ToolkitScriptManager=ctl00%24workspace%24upTools%7Cctl00%24workspace%24ddlSport&ctl00_ToolkitScriptManager_HiddenField=%3B%3BAjaxControlToolkit%2C%20Version%3D3.5.7.123%2C%20Culture%3Dneutral%2C%20PublicKeyToken%3D28f01b0e84b6d53e%3Aen-US%3A5214fb5a-fe22-4e6b-a36b-906c0237d796%3Ade1feab2%3Af9cec9bc%3Aa67c2700%3Af2c8e708%3A720a52bf%3A589eaa30%3A8613aea7%3A3202a5a2%3Aab09e3fe%3A87104b7c%3Abe6fb298&__LASTFOCUS=&__EVENTTARGET=ctl00%24workspace%24ddlSport&__EVENTARGUMENT=&__VIEWSTATE=&__VIEWSTATEENCRYPTED=&ctl00%24toolboxRight%24tbLoginUserName=erneker&ctl00%24toolboxRight%24tbLoginPassword=626549&ctl00%24workspace%24ddlLocality=171&ctl00%24workspace%24ddlSport=140&__ASYNCPOST=true&",
		"--compressed"}
}


func checkFreePlace(s *search, div *Div) {
	beginningIndex := convertTimeToIndex(s.From.Format("15:04"))
	log.Printf("Beginning index %d", beginningIndex)
	date := s.From.Format("2006-01-02")
	halfHoursToSearch := int(s.Till.Sub(*s.From).Minutes() / 30)
	for {
		dateIndex := calculateDateIndex(date)
		log.Printf("Date index %d", dateIndex)
		freeInARow := 0
		for k := beginningIndex; k < beginningIndex + halfHoursToSearch; k++ {
			if k > 33 {
				break
			}
			if div.Table.Rows[dateIndex].Data[k].isFree() {
				freeInARow += 1
			} else {
				freeInARow = 0
			}
			if freeInARow == s.Length {
				message := fmt.Sprintf("To: You\nSubject: Court is free\nHello,\n\ncourt you have requested at Hamr Sport is avaiable.\n\nSearch:\ndate = %s\nbeginningTime = %s\n\nHave a nice day.\n\nJOT", date, convertIndexToTime(k - s.Length + 1))
				if err := smtp.SendMail("smtp.gmail.com:587", smtp.PlainAuth("", "jot.company@gmail.com", "harrison7", "smtp.gmail.com"), "jot.company@gmail.com", s.Emails, []byte(message)); err != nil {
					log.Printf("Unable to send e-mail. %v", err)
				}
				removeSearch(s)
				break
			}
		}
		if freeInARow == s.Length {
			break
		}
		log.Printf("Checking Hamr Sport for court with parameters: beginningTime = %s, date = %s, to = %v, reservationLength = %d, halfHoursToSearch = %d", s.From, date, s.Emails, s.Length, halfHoursToSearch)
		time.Sleep(time.Second * 60)
	}
}

func getAndProcessResponse() *Div {
	out, err := exec.Command("curl", requestParams...).Output()
	if err != nil {
		log.Printf("Unable to get response. %v", err)
		return nil
	}
	resp := string(out)
	log.Printf(resp)
	resp = resp[strings.Index(resp, "<div id=\"ctl00_workspace_ReservationGrid_divResGrid\" class=\"resgrid\">"):]
	reg := regexp.MustCompile(`\|\d+\|updatePanel\|ctl00_workspace_upLegend\|`)
	resp = resp[:strings.Index(resp, reg.FindString(resp))]
	resp = strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			if r == ' ' {
				return r
			}
			return -1
		}
		return r
	}, resp)
	v := &Div{}
	if err := xml.Unmarshal([]byte(strings.Replace(resp, "&nbsp;", "", -1)), &v); err != nil {
		log.Printf("Unable to process response. %v", err)
		return nil
	}
	return v
}

func calculateDateIndex(date string) int {
	t, _ := time.Parse(TIME_FORMAT, date)
	duration := t.Sub(time.Now())
	return int(duration.Hours()/24) + 2
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
	return index
}

func convertIndexToTime(index int) string {
	hour := index + 14
	if hour % 2 == 1 {
		return fmt.Sprintf("%d:30", hour / 2)
	}
	return fmt.Sprintf("%d:00", hour / 2)
}

func (data *Data) isFree() bool {
	return strings.Contains(data.InnerDiv.Class, "rgs-free")
}

func runSearch(s *search) {
	log.Printf("Starting search %v", s)
	found := false
	for i := 0; i < len(s.Emails); i++ {
		if s.Emails[i] == "jot.company@gmail.com" {
			found = true
			break
		}
	}
	if !found {
		s.Emails = append(s.Emails, "jot.company@gmail.com")
	}
	checkFreePlace(s, getAndProcessResponse())
}
package main

import (
	"os/exec"
	"fmt"
	"strings"
	"unicode"
	"encoding/xml"
	"strconv"
	"time"
	"net/smtp"
	"flag"
	"os"
	"regexp"
)

const(
//	LOC_BRANIK = 171
//	SPORT_BADMINTON = 140
	TIME_FORMAT = "2006-01-02"
)

var(
	requestParams []string
	beginningTime, date, toAddr string
	reservationLength, halfHoursToSearch int
)

func init() {
	requestParams = []string{"http://hodiny.hamrsport.cz/Login.aspx",
		"-H", "Cookie: _ga=GA1.2.1761385536.1450711716; HamrOnline$SessionId=5zw1pv45b0mocffcv5e3lvq5; __utmt=1; __utma=74282507.1761385536.1450711716.1460970159.1461009504.4; __utmb=74282507.8.10.1461009504; __utmc=74282507; __utmz=74282507.1461009504.4.4.utmcsr=google|utmccn=(organic)|utmcmd=organic|utmctr=(not%20provided)",
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
	flag.StringVar(&beginningTime, "beginningTime", "", "")
	flag.StringVar(&date, "date", "", "")
	flag.StringVar(&toAddr, "to", "", "")
	flag.IntVar(&reservationLength, "reservationLength", 2, "")
	flag.IntVar(&halfHoursToSearch, "halfHoursToSearch", 4, "")
}

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
	Data []*Data `xml:"td"`
	Class string `xml:"class,attr"`
}

type Data struct {
	Class string `xml:"class,attr"`
	Id string    `xml:"id,attr"`
	InnerDiv *InnerDiv `xml:"div"`
}

type InnerDiv struct {
	Class string `xml:"class,attr"`
}

func main() {
	flag.Parse()
	if len(date) != 10 {
		fmt.Println("Date formatting is wrong. Input leading zeroes if needed.")
		os.Exit(1)
	}
	toAddresses := strings.Split(toAddr, ",")
	beginningIndex := convertTimeToindex(beginningTime)
	checkFreePlace(beginningIndex, toAddresses, getAndProcessResponse())
}

func checkFreePlace(beginningIndex int, toAddresses []string, div Div) {
	for {
		dateIndex := calculateDateIndex(date)
		freeInARow := 0
		for k := beginningIndex; k < beginningIndex + halfHoursToSearch; k++ {
			if k > 32 {
				break
			}
			if div.Table.Rows[dateIndex].Data[k].isFree() {
				freeInARow += 1
			} else {
				freeInARow = 0
			}
			if freeInARow == reservationLength {
				message := fmt.Sprintf("To: %s\nSubject: Court is free\nHello,\n\ncourt you have requested at Hamr Sport is avaiable.\n\nSearch:\ndate = %s\nbeginningTime = %s\n\nHave a nice day.\n\nJOT", toAddr, date, beginningTime)
				handleError(smtp.SendMail("smtp.gmail.com:587", smtp.PlainAuth("", "jot.company@gmail.com", "moderator7", "smtp.gmail.com"), "jot.company@gmail.com", toAddresses, []byte(message)))
				break
			}
		}
		if freeInARow == reservationLength {
			break
		}
		fmt.Println(fmt.Sprintf("Checking Hamr Sport for court with parameters: beginningTime = %s, date = %s, to = %v, reservationLength = %d, halfHoursToSearch = %d", beginningTime, date, toAddresses, reservationLength, halfHoursToSearch))
		time.Sleep(time.Second * 60)
	}
}

func getAndProcessResponse() Div {
	out, err := exec.Command("curl", requestParams...).Output()
	handleError(err)
	resp := string(out)
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
	v := Div{}
	handleError(xml.Unmarshal([]byte(strings.Replace(resp, "&nbsp;", "", -1)), &v))
	return v
}

func calculateDateIndex(date string) int {
	t, _ := time.Parse(TIME_FORMAT, date)
	duration := t.Sub(time.Now())
	return int(duration.Hours() / 24) + 1
}

func convertTimeToindex(time string) int {
	if len(time) == 4 {
		time = fmt.Sprintf("0%s", time)
	}
	i, err := strconv.Atoi(time[0:2])
	handleError(err)
	index := 2 * i - 13
	if strings.HasSuffix(time, ":30") {
		index += 1
	}
	return index
}

func (data *Data) isFree() bool {
	return strings.Contains(data.InnerDiv.Class, "rgs-free")
}

func handleError(err error) {
	if err != nil {
		panic(err)
	}
}
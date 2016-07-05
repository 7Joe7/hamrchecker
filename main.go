package main

import (
	"encoding/xml"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/smtp"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
	"encoding/json"
	"io/ioutil"
	"os"
	"net/url"
	"errors"
	"encoding/base64"
)

const (
	//	LOC_BRANIK = 171
	//	SPORT_BADMINTON = 140
	TIME_FORMAT    = "2006-01-02"
	SEARCHES_STORE = "searches.json"
)

var (
	requestParams                        []string
	beginningTime, date, toAddr          string
	reservationLength, halfHoursToSearch int
	searchesMutex                        sync.Mutex
	searches                             []*search
)

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
	Email  string     `json:"email,attr"`
	From   *time.Time `json:"from,attr"`
	Till   *time.Time `json:"till,attr"`
	Length int        `json:"length,attr"`
	Start  *time.Time `json:"start,attr"`
}

func (s *search) isEmpty() bool {
	if s.Email == "" && s.From == nil && s.Till == nil && s.Length == 0 {
		return true
	}
	return false
}

func (s *search) isValid() bool {
	if s.Email != "" && s.From != nil && s.Till != nil && s.Length > 0 {
		return true
	}
	return false
}

func (s *search) Description() string {
	return fmt.Sprintf("%d half hours on %s between %s and %s started at %s.", s.Length, s.From.Format("2.1.2006"), s.From.Format("15:04"), s.Till.Format("15:04"), s.Start.Format("15:04 2.1.2006"))
}

func (s *search) parse(form url.Values) error {
	var date *time.Time
	if len(form["email"]) > 0 {
		s.Email = form["email"][0]
	} else {
		return errors.New("E-mail not provided.")
	}
	if len(form["date"]) > 0 {
		d, err := time.Parse("2006-01-02", form["date"][0])
		if err != nil {
			return fmt.Errorf("Unable to parse date. %v", err)
		} else {
			date = &d
		}
	} else {
		return errors.New("Date not provided.")
	}
	if len(form["from"]) > 0 {
		if strings.Contains(form["from"][0], ":00") || strings.Contains(form["from"][0], ":30") {
			from, err := time.Parse("January 2 2006 15:04", fmt.Sprintf("%s %d %d %s", date.Month().String(), date.Day(), date.Year(), form["from"][0]))
			if err != nil {
				return fmt.Errorf("Unable to parse from time. %v", err)
			} else {
				s.From = &from
			}
		} else {
			return errors.New("From time should be divisible by half hour.")
		}
	} else {
		return errors.New("From time not provided.")
	}
	if len(form["till"]) > 0 {
		if strings.Contains(form["till"][0], ":00") || strings.Contains(form["till"][0], ":30") {
			till, err := time.Parse("January 2 2006 15:04", fmt.Sprintf("%s %d %d %s", date.Month().String(), date.Day(), date.Year(), form["till"][0]))
			if err != nil {
				return fmt.Errorf("Unable to parse till time. %v", err)
			} else {
				s.Till = &till
			}
		} else {
			return errors.New("Till time should be divisible by half hour.")
		}
	} else {
		return errors.New("Till time not provided.")
	}
	if len(form["length"]) > 0 {
		if length, err := strconv.Atoi(form["length"][0]); err != nil {
			log.Printf("Unable to parse length. %v", err)
		} else {
			s.Length = length
		}
	} else {
		return errors.New("Number of consecutive half hours not provided.")
	}
	log.Printf("Parsed data: %v", s)
	return nil
}


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
}

func saveSearchesToFile() error {
	searchesB, err := json.Marshal(&searches)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(SEARCHES_STORE, searchesB, 777)
}

func loadSearches() {
	searchesB, err := ioutil.ReadFile(SEARCHES_STORE)
	if err != nil {
		// consider searches file non-existing
		searches = []*search{}
	}
	if err := json.Unmarshal(searchesB, &searches); err != nil {
		// searches file is corrupted
		log.Printf("Searches store is corrupted. Backing it up to %s.old.\n", SEARCHES_STORE)
		backupFileName := fmt.Sprintf("%s.old", SEARCHES_STORE)
		os.Remove(backupFileName)
		if err := ioutil.WriteFile(backupFileName, searchesB, 777); err != nil {
			log.Printf("Unable to backup corrupted searches store.\n")
		}
		searches = []*search{}
	}
}

func addSearch(s *search) {
	searchesMutex.Lock()
	defer searchesMutex.Unlock()
	now := time.Now()
	s.Start = &now
	searches = append(searches, s)
	if err := saveSearchesToFile(); err != nil {
		log.Printf("Unable to save searches to store. %v", err)
	}
}

func removeSearch(s *search) {
	searchesMutex.Lock()
	defer searchesMutex.Unlock()
	for i := 0; i < len(searches); i++ {
		if searches[i] == s {
			searches = append(searches[:i], searches[i+1:]...)
			break
		}
	}
	if err := saveSearchesToFile(); err != nil {
		log.Printf("Unable to save searches to store. %v", err)
	}
}

func removeSearchByIndex(i int) {
	searchesMutex.Lock()
	defer searchesMutex.Unlock()
	searches = append(searches[:i], searches[i+1:]...)
	if err := saveSearchesToFile(); err != nil {
		log.Printf("Unable to save searches to store. %v", err)
	}
}

func runSearch(s *search) {

}

func index(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Fatal(err)
	}
	search := &search{}
	search.parse(r.Form)
	if !search.isEmpty() {
		if search.isValid() {
			addSearch(search)
			go runSearch(search)
		} else {
			// save the state
			SetFlash(w, "message", []byte("This is a flashed message!"))
			// TODO inform
		}
	}
	templates := template.Must(template.ParseFiles("resources/html/hamrchecker.html"))
	templates.ExecuteTemplate(w, "hamrchecker", searches)
}

func main() {
	mux := http.DefaultServeMux
	mux.HandleFunc("/", index)
	http.Handle("/resources/", http.StripPrefix("/resources/", http.FileServer(http.Dir("resources"))))
	server := &http.Server{
		Addr:    "0.0.0.0:8080",
		Handler: mux,
	}
	loadSearches()
	for i := 0; i < len(searches); i++ {
		if time.Now().Before(*searches[i].Till) {
			removeSearchByIndex(i)
		} else {
			go runSearch(searches[i])
		}
	}
	server.ListenAndServe()
}

func checkFreePlace(beginningIndex int, toAddresses []string, div Div) {
	for {
		dateIndex := calculateDateIndex(date)
		freeInARow := 0
		for k := beginningIndex; k < beginningIndex+halfHoursToSearch; k++ {
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
	return int(duration.Hours()/24) + 1
}

func convertTimeToindex(time string) int {
	if len(time) == 4 {
		time = fmt.Sprintf("0%s", time)
	}
	i, err := strconv.Atoi(time[0:2])
	handleError(err)
	index := 2*i - 13
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

func SetFlash(w http.ResponseWriter, name string, value []byte) {
	c := &http.Cookie{Name: name, Value: base64.URLEncoding.EncodeToString(value)}
	http.SetCookie(w, c)
}

func GetFlash(w http.ResponseWriter, r *http.Request, name string) ([]byte, error) {
	c, err := r.Cookie(name)
	if err != nil {
		switch err {
		case http.ErrNoCookie:
			return nil, nil
		default:
			return nil, err
		}
	}
	value, err := base64.URLEncoding.DecodeString(c.Value)
	if err != nil {
		return nil, err
	}
	dc := &http.Cookie{Name: name, MaxAge: -1, Expires: time.Unix(1, 0)}
	http.SetCookie(w, dc)
	return value, nil
}

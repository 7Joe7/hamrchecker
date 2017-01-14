package hamrui

import (
	"net/http"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"github.com/7joe7/hamrchecker/resources"
)

func getSession() (*http.Cookie, error) {
	log.Printf("Getting new session.")
	req, err := http.NewRequest("GET", "http://hodiny.hamrsport.cz/", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Connection", "keep-alive")

	c := getClient()

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	if len(resp.Cookies()) == 0 {
		return nil, fmt.Errorf("Unable to retrieve session id.")
	}
	log.Printf("Cookies: %v", resp.Cookies())
	for i := 0; i < len(resp.Cookies()); i++ {
		if resp.Cookies()[i].Name == "HamrOnline$SessionId" {
			log.Printf("Successful retrieved new session.")
			return resp.Cookies()[i], nil
		}
	}
	return nil, fmt.Errorf("Unable to retrieve session id.")
}

func changeSessionFocus(sessionId *http.Cookie, place, sport string, changeType int) error {
	log.Printf("Changing session focus to %s, %s, type %d.", place, sport, changeType)
	changeBody := "ctl00%24workspace%24ddlLocality=" + place + "&ctl00%24workspace%24ddlSport=" + sport
	switch changeType {
	case resources.SESSION_FOCUS_CHANGE_TYPE_PLACE:
		changeBody = "ctl00%24ToolkitScriptManager=ctl00%24workspace%24upTools%7Cctl00%24workspace%24ddlLocality&" + changeBody + "&__EVENTTARGET=ctl00%24workspace%24ddlLocality"
	case resources.SESSION_FOCUS_CHANGE_TYPE_SPORT:
		changeBody = "ctl00%24ToolkitScriptManager=ctl00%24workspace%24upTools%7Cctl00%24workspace%24ddlSport&" + changeBody + "&__EVENTTARGET=ctl00%24workspace%24ddlSport"
	}

	req, err := http.NewRequest("POST", "http://hodiny.hamrsport.cz/Login.aspx",
		strings.NewReader(changeBody))
	if err != nil {
		return err
	}

	req.AddCookie(sessionId)

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/55.0.2883.95 Safari/537.36")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("X-MicrosoftAjax", "Delta=true")

	c := getClient()
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Expected status code 200, got %d. %s", resp.StatusCode, resp.Status)
	}
	log.Printf("Successfully changed session focus.")
	return nil
}

func getData(sessionId *http.Cookie) (string, error) {
	req, err := http.NewRequest("GET", "http://hodiny.hamrsport.cz/Login.aspx", nil)
	if err != nil {
		return "", err
	}

	req.AddCookie(sessionId)

	req.Header.Set("Connection", "keep-alive")

	c := getClient()
	resp, err := c.Do(req)
	if err != nil {
		return "", err
	}
	var body []byte
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
	}
	if resp.StatusCode != 200 {
		return string(body), fmt.Errorf("Expected status code 200, got %d. %s", resp.StatusCode, resp.Status)
	}
	return string(body), err
}

func getClient() *http.Client {
	return &http.Client{}
}

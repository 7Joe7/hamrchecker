package hamrui

import "net/http"

func GetSession() (*http.Cookie, error) {
	return getSession()
}

func ChangeSessionFocus(sessionId *http.Cookie, place, sport string, changeType int) error {
	return changeSessionFocus(sessionId, place, sport, changeType)
}

func GetData(sessionId *http.Cookie) (string, error) {
	return getData(sessionId)
}

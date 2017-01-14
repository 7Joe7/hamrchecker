package checker

import (
	"testing"
	"time"
)

var (
	testCases = []calcDateIndexTestCase{
		{Result: 3, WantedDate: "2017-01-16", NowDate: getParsedTime("02 Jan 06 15:04 -0700", "14 Jan 17 00:25 +0100")},
		{Result: 1, WantedDate: "2017-01-14", NowDate: getParsedTime("02 Jan 06 15:04 -0700", "14 Jan 17 00:25 +0100")},
	}
)

type calcDateIndexTestCase struct {
	Result int
	WantedDate string
	NowDate time.Time
}

func TestCalculateDateIndex(t *testing.T) {
	for i := 0; i < len(testCases); i++ {
		actualResult := CalculateDateIndex(testCases[i].WantedDate, testCases[i].NowDate.Format("2006-01-02"))
		if testCases[i].Result != actualResult {
			t.Errorf("For data %s, %s. Expected result is %d, got %d.", testCases[i].WantedDate, testCases[i].NowDate, testCases[i].Result, actualResult)
		}
	}
}

func getParsedTime(layout, value string) time.Time {
	t, _ := time.Parse(layout, value)
	return t
}

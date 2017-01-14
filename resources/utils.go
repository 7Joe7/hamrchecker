package resources

import "time"

func PlaceIdToName(id string) string {
	switch id {
	case "171":
		return "Braník"
	case "172":
		return "Štěrboholy"
	case "169":
		return "Záběhlice"
	default:
		return "Unknown"
	}
}

func SportIdToName(id string) string {
	switch id {
	case "137":
		return "Tenis"
	case "138":
		return "Squash"
	case "140":
		return "Badminton"
	case "142":
		return "Fotbal"
	case "144":
		return "Florbal"
	case "149":
		return "Beach volejbal"
	case "150":
		return "Stolní tenis"
	default:
		return "Unknown"
	}
}

func GetTimePointer(t time.Time) *time.Time {
	return &t
}

func GetTimeFromTimePointer(t *time.Time) time.Time {
	return *t
}

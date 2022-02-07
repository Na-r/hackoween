package countdown

import (
	"math"
	"strconv"
	"time"
)

const START_MONTH = time.January //time.October
const START_DAY int = 1          //18
const ADVENT_DURATION int = 13

func Get_duration() string {

	t := time.Now()
	month := t.Month()
	day := t.Day()
	if month == START_MONTH && day >= START_DAY && day < START_DAY+ADVENT_DURATION {
		cd := time.Date(t.Year(), START_MONTH, START_DAY+((day-START_DAY)+1), 0, 0, 0, 0, time.Now().Location())
		duration_obj := time.Until(cd)
		hours := duration_obj.Hours()
		minutes := int(math.Round(duration_obj.Minutes())) % 60
		time := (strconv.Itoa(int(hours)) + "h " + strconv.Itoa(int(minutes)) + "m")
		return time
	} else {
		return "Thanks for playing!"
	}

}

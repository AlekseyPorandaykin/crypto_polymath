package scheduler

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func Test_durationToNextMinute(t *testing.T) {
	from := time.Date(2024, 05, 1, 1, 1, 20, 0, time.UTC)
	dur := durationToMinute(from, 1)
	require.EqualValues(t, 40, dur.Seconds())
}

func Test_durationToNextHour(t *testing.T) {
	now := time.Now()
	from := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 20, 0, 0, time.UTC)
	dur := durationToNextHour(from)
	require.EqualValues(t, 40, dur.Minutes())
}

func Test_durationToNextDay(t *testing.T) {
	now := time.Now()
	from := time.Date(now.Year(), now.Month(), now.Day(), 10, 00, 0, 0, time.UTC)
	dur := durationToNextDay(from)
	require.EqualValues(t, 14, dur.Hours())
}

func Test_durationToNextWeek(t *testing.T) {
	from := time.Date(2024, 05, 6, 00, 00, 0, 0, time.UTC)
	dur := durationToNextWeek(from)
	require.EqualValues(t, 7, dur.Hours()/24)
}

func Test_durationToNextMonth(t *testing.T) {
	from := time.Date(2024, 05, 1, 00, 00, 0, 0, time.UTC)
	dur := durationToNextMonth(from)
	require.EqualValues(t, 31, int(dur.Hours()/24))
}

func Test_durationToMinute(t *testing.T) {
	from := time.Now()
	dur := durationToMinute(from, 1)
	fmt.Println(dur)
	fmt.Println(from.Format(time.DateTime))
	fmt.Println(from.Add(dur).Format(time.DateTime))
}

func Test_durationToHour(t *testing.T) {
	from := time.Now()
	dur := durationToHour(from, 5)
	fmt.Println(dur)
	fmt.Println(from.Format(time.DateTime))
	fmt.Println(from.Add(dur).Format(time.DateTime))
}

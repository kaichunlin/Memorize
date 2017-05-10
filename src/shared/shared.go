package shared

import (
	"os"
	"time"
)

const (
	TimeFormatCreated = "2006-01-02 15:04:05-07:00"
	TimeFormatNotify  = "2006-01-02 15"
)

var offset = [...]int{1, 2, 3, 5, 8, 13, 21, 34, 55, 89, 144}

func NotifyTimeUnit() time.Duration {
	if os.Getenv("TEST_MODE") == "true" {
		return time.Hour
	} else {
		return 24 * time.Hour
	}
}

func NextNotifyTime(n int, creationTime string) string {
	t, _ := time.Parse(TimeFormatCreated, creationTime)
	t = StandardizeTime(t)
	t = t.Add(NotifyTimeUnit() * time.Duration(notifyOffset(n)))
	return t.Format(TimeFormatNotify)
}

func CalculateDay(n int) int {
	return notifyOffset(n)
}

func CurrentNotifyTime() string {
	return StandardizeTime(time.Now()).Format(TimeFormatNotify)
}

func CreationTime() string {
	return StandardizeTime(time.Now()).Format(TimeFormatCreated)
}

func StandardizeTime(t time.Time) time.Time {
	utc, err := time.LoadLocation("Etc/UTC")
	if err != nil {
		panic(err)
	}
	return t.In(utc)
}

func notifyOffset(n int) int {
	return offset[n]
}

func NotifyOffsetCount() int {
	return len(offset)
}

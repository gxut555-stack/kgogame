package common

import (
	"time"
)

func GetZero(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
}

func GetTimeZero(t int64) int64 {
	t2 := time.Unix(t, 0)
	return time.Date(t2.Year(), t2.Month(), t2.Day(), 0, 0, 0, 0, time.Local).Unix()

}

func GetMinKeyByTime(t time.Time) int64 {
	zeroTime := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	return (t.Unix() - zeroTime.Unix()) / 60
}

const (
	DATE_NORMAL          = "2006-01-02 15:04:05"
	DATE_Milli           = "2006-01-02 15:04:05.000"     //毫秒
	DATE_Milli_Zone      = "2006-01-02 MST 15:04:05.000" //毫秒带时区
	DATE_YYYYMMDD        = "20060102"
	DATE_YYYY_MM_DD      = "2006-01-02"
	DATE_YYYY_MM_DD_HHMM = "2006-01-02 15:04"
)

// unix时间戳转换成日期格式
func UnixToDate(unix int64, format string) string {
	if unix <= 0 {
		return ""
	}
	return time.Unix(unix, 0).Format(format)
}

// 当前毫秒对应的时间
func CurrentMilli() string {
	return time.Now().Format(DATE_Milli_Zone)
}

// 日期转换成时间戳
func YmdToUnix[T Number](year, month, day T) int64 {
	return DateToUnix(year, month, day, 0, 0, 0)
}

// 日期转换成时间戳
func DateToUnix[T Number](year, month, day, hour, minute, second T) int64 {
	return time.Date(int(year), time.Month(month), int(day), int(hour), int(minute), int(second), 0, time.Local).Unix()
}

// 日期转换成time结构体
func YmdToTime[T Number](year, month, day T) time.Time {
	return DateToTime(year, month, day, 0, 0, 0)
}

// 日期转换成时间戳
func DateToTime[T Number](year, month, day, hour, minute, second T) time.Time {
	return time.Date(int(year), time.Month(month), int(day), int(hour), int(minute), int(second), 0, time.Local)
}

// 获取今日0点unix时间戳
func GetTodayZeroUnixTime() int64 {
	ts := time.Now()
	return time.Date(ts.Year(), ts.Month(), ts.Day(), 0, 0, 0, 0, time.Local).Unix()

}

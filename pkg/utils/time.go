package utils

import "time"

func StartOfWeek(t time.Time) time.Time {
    year, month, day := t.Date()
    start := time.Date(year, month, day, 0, 0, 0, 0, t.Location())
    offset := int(start.Weekday())
    if offset == 0 {
        offset = 7
    }
    return start.AddDate(0, 0, -offset+1)
}

func EndOfWeek(t time.Time) time.Time {
    start := StartOfWeek(t)
    return start.AddDate(0, 0, 6)
}
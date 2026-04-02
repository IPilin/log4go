package utils

import (
	"bytes"
	"time"
)

func AppendFormatTime(buf *bytes.Buffer, t time.Time) {
	year, month, day := t.Date()
	hour, min, sec := t.Clock()

	b := buf.AvailableBuffer()

	b = append(b, byte(year/1000)+'0', byte(year/100%10)+'0', byte(year/10%10)+'0', byte(year%10)+'0')
	b = append(b, '-')
	b = append2Digits(b, int(month))
	b = append(b, '-')
	b = append2Digits(b, day)
	b = append(b, ' ')
	b = append2Digits(b, hour)
	b = append(b, ':')
	b = append2Digits(b, min)
	b = append(b, ':')
	b = append2Digits(b, sec)

	buf.Write(b)
}

func append2Digits(b []byte, d int) []byte {
	return append(b, byte(d/10)+'0', byte(d%10)+'0')
}

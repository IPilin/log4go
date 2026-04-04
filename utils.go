package log4go

import (
	"bytes"
	"time"
)

// appendFormatTime writes the formatted time "YYYY-MM-DD HH:MM:SS" to the buffer.
func appendFormatTime(buf *bytes.Buffer, t time.Time) {
	year, month, day := t.Date()
	hour, min, sec := t.Clock()

	var a [19]byte
	a[0] = byte(year/1000) + '0'
	a[1] = byte(year/100%10) + '0'
	a[2] = byte(year/10%10) + '0'
	a[3] = byte(year%10) + '0'
	a[4] = '-'
	a[5] = byte(month/10) + '0'
	a[6] = byte(month%10) + '0'
	a[7] = '-'
	a[8] = byte(day/10) + '0'
	a[9] = byte(day%10) + '0'
	a[10] = ' '
	a[11] = byte(hour/10) + '0'
	a[12] = byte(hour%10) + '0'
	a[13] = ':'
	a[14] = byte(min/10) + '0'
	a[15] = byte(min%10) + '0'
	a[16] = ':'
	a[17] = byte(sec/10) + '0'
	a[18] = byte(sec%10) + '0'

	buf.Write(a[:])
}

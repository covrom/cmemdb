package db

import (
	"errors"
	"strconv"
	"time"
)

type IDEntry uint32

func (id IDEntry) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatUint(uint64(id), 10)), nil
}

func (id *IDEntry) UnmarshalJSON(data []byte) error {
	i, err := strconv.ParseUint(string(data), 10, 32)
	if err != nil {
		return err
	}
	*id = IDEntry(i)
	return nil
}

type IDEntryBytes []byte

func (id IDEntryBytes) Int() IDEntry {
	var us uint64
	for _, digit := range id {
		d := digit - '0'
		if d > 9 {
			break
		}
		us = us*10 + uint64(d)
	}
	return IDEntry(us)
}

var NullTime int32

func init() {
	zt, _ := time.Parse(time.RFC3339, "1949-12-31T23:59:59Z")
	NullTime = int32(zt.Unix())
}

type TimeStamp []byte

func (ts TimeStamp) Time() time.Time {
	return time.Unix(int64(ts.Int()), 0).UTC()
}

func (ts TimeStamp) Int() int32 {
	if len(ts) == 0 {
		return NullTime
	}
	var us int64
	mul := int32(1)
	for i, digit := range ts {
		if i == 0 && digit == '-' {
			mul = -1
			continue
		}
		d := digit - '0'
		if d > 9 {
			break
		}
		us = us*10 + int64(d)
	}

	return mul * int32(us)
}

var ErrNotTimeStamp = errors.New("not a timestamp")

func (ts TimeStamp) Validate() error {
	for i, digit := range ts {
		if i == 0 && digit == '-' {
			continue
		}
		d := digit - '0'
		if d > 9 {
			return ErrNotTimeStamp
		}
	}
	return nil
}

func (ts TimeStamp) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatUint(uint64(ts.Int()), 10)), nil
}

func (ts *TimeStamp) UnmarshalJSON(data []byte) error {
	*ts = TimeStamp(data)
	return ts.Validate()
}

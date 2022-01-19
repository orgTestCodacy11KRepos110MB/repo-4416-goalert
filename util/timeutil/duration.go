package timeutil

import (
	"regexp"
	"strconv"
	"time"
	"unicode"

	"github.com/pkg/errors"
)

// ISODuration represents an ISO duration string.
// The time components are combined, and the weeks component
// is interpreted as a shorthand for 7 days.
type ISODuration struct {
	Years, Months, Days int
	TimePart            time.Duration
}

var re = regexp.MustCompile(`^P\B(\d+Y)?(\d+M)?(\d+W)?(\d+D)?(T\B(\d+H)?(\d+M)?(\d+S)?)?$`)

// ParseISODuration parses the components of an ISO Duration string.
// The time components are accurate and are aggregated into one TimePart.
// The nominal date components cannot be aggregated without accounting for daylight savings time.
// Supported formats are "PnYnMnDTnHnMnS" and "PnW".
// Negative and decimal units are not supported.
func ParseISODuration(s string) (d ISODuration, err error) {
	if !re.MatchString(s) {
		return d, errors.Errorf(`invalid format: %s must be an ISO Duration`, s)
	}

	left, right := 1, 1 // sliding window
	isTime := false

	for _, c := range s[1:] {
		if unicode.IsDigit(c) {
			right++
			continue
		}

		if string(c) == "T" {
			isTime = true
			right++
			left = right
			continue
		}

		digits, err := strconv.Atoi(s[left:right])
		if err != nil {
			return d, err
		}

		switch string(c) {
		case "Y":
			d.Years += digits
		case "M":
			if isTime {
				digits *= 60
			} else {
				d.Months += digits
			}
		case "D":
			d.Days += digits
		case "W":
			d.Days += (digits * 7)
		case "H":
			digits *= 3600
		case "S":
			// ok
		default:
			return d, errors.Errorf("invalid character encountered: %s", string(c))
		}

		if isTime {
			dur, err := time.ParseDuration(strconv.Itoa(digits) + "s")
			if err != nil {
				return d, err
			}

			d.TimePart += dur
		}

		right++
		left = right
	}

	return d, err
}
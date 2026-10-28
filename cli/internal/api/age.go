package api

import "time"

// MinimumAge is the minimum age, in years, to hold an account.
const MinimumAge = 18

// IsAdult reports whether someone born on birthdate (YYYY-MM-DD) is at least
// MinimumAge years old on the given day. Unparseable dates are not adults.
func IsAdult(birthdate string, now time.Time) bool {
	born, err := time.Parse("2006-01-02", birthdate)
	if err != nil {
		return false
	}
	cutoff := now.UTC().AddDate(-MinimumAge, 0, 0)
	return !born.After(cutoff)
}

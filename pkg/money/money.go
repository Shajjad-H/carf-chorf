// Package money provides display formatting for monetary amounts.
package money

import "fmt"

// FormatCents formats a USD cent amount, for example 1900 -> "$19.00".
func FormatCents(cents uint) string {
	return fmt.Sprintf("$%d.%02d", cents/100, cents%100)
}

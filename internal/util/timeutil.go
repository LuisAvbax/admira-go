package util

import "time"

func ParseYMD(s string) time.Time {
	t, _ := time.Parse("2006-01-02", s); return t
}
func ParseRFC3339(s string) time.Time {
	t, _ := time.Parse(time.RFC3339, s); return t
}
func NonNegInt(v int) int       { if v < 0 { return 0 }; return v }
func NonNegFloat(v float64) float64 { if v < 0 { return 0 }; return v }
func SafeDiv(a, b float64) float64  { if b == 0 { return 0 }; return a / b }

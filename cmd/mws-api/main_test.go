package main

import (
	"testing"
	"time"
)

func TestStartupTerms(t *testing.T) {
	tests := []struct {
		name string
		date time.Time
		want []string
	}{
		{
			name: "January",
			date: time.Date(2026, time.January, 15, 0, 0, 0, 0, time.Local),
			want: []string{"2263", "2265"},
		},
		{
			name: "March",
			date: time.Date(2026, time.March, 31, 0, 0, 0, 0, time.Local),
			want: []string{"2263", "2265"},
		},
		{
			name: "April",
			date: time.Date(2026, time.April, 1, 0, 0, 0, 0, time.Local),
			want: []string{"2265", "2267"},
		},
		{
			name: "September",
			date: time.Date(2026, time.September, 30, 0, 0, 0, 0, time.Local),
			want: []string{"2265", "2267"},
		},
		{
			name: "October",
			date: time.Date(2026, time.October, 1, 0, 0, 0, 0, time.Local),
			want: []string{"2273"},
		},
		{
			name: "December",
			date: time.Date(2026, time.December, 31, 0, 0, 0, 0, time.Local),
			want: []string{"2273"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := startupTerms(test.date)

			if len(got) != len(test.want) {
				t.Fatalf("startupTerms() = %v, expected %v", got, test.want)
			}

			for i := range got {
				if got[i] != test.want[i] {
					t.Fatalf("startupTerms() = %v, expected %v", got, test.want)
				}
			}
		})
	}
}

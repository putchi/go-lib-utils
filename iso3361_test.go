package utils

import (
	"testing"
)

func TestGetISO3361CodeCountry(t *testing.T) {
	type args struct {
		country string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"test Kenya",
			args{"Kenya"},
			"KE",
		},
		{
			"test AUSTRIA",
			args{"AUSTRIA"},
			"AT",
		},
		{
			"test empty",
			args{""},
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetISO3361CodeCountry(tt.args.country); got != tt.want {
				t.Errorf("GetISOCodeCountry() = %v, want %v", got, tt.want)
			}
		})
	}
}

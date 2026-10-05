package domain

import "testing"

func TestAgreementTypeValid(t *testing.T) {
	tests := []struct {
		in   AgreementType
		want bool
	}{
		{AgreementTypeICLA, true},
		{AgreementTypeCCLA, true},
		{"", false},
		{"icla", false},
		{"OTHER", false},
	}
	for _, tt := range tests {
		if got := tt.in.Valid(); got != tt.want {
			t.Errorf("AgreementType(%q).Valid() = %v, want %v", tt.in, got, tt.want)
		}
	}
}

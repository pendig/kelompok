package organizations

import "testing"

func TestNormalizeClaimStatus(t *testing.T) {
	cases := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{"", "", false},
		{"   ", "", false},
		{"all", "", false},
		{"Any", "", false},
		{"*", "", false},
		{"pending", "pending", false},
		{" Pending ", "pending", false},
		{"approved", "approved", false},
		{"REJECTED", "rejected", false},
		{"weird", "", true},
		{"published", "", true},
	}

	for _, tc := range cases {
		got, err := NormalizeClaimStatus(tc.in)
		if tc.wantErr {
			if err == nil {
				t.Errorf("NormalizeClaimStatus(%q): expected error, got %q", tc.in, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("NormalizeClaimStatus(%q): unexpected error %v", tc.in, err)
			continue
		}
		if got != tc.want {
			t.Errorf("NormalizeClaimStatus(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestNormalizeClaimDecision(t *testing.T) {
	cases := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{"approve", "approve", false},
		{"approved", "approve", false},
		{"  Approve ", "approve", false},
		{"reject", "reject", false},
		{"Rejected", "reject", false},
		{"REJECTED", "reject", false},
		{"", "", true},
		{"deny", "", true},
		{"approve-now", "", true},
	}

	for _, tc := range cases {
		got, err := NormalizeClaimDecision(tc.in)
		if tc.wantErr {
			if err == nil {
				t.Errorf("NormalizeClaimDecision(%q): expected error, got %q", tc.in, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("NormalizeClaimDecision(%q): unexpected error %v", tc.in, err)
			continue
		}
		if got != tc.want {
			t.Errorf("NormalizeClaimDecision(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

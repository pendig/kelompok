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

func TestNormalizeClaimListLimit(t *testing.T) {
	cases := []struct {
		name    string
		limit   int
		want    int
		wantErr bool
	}{
		{"minimum", 1, 1, false},
		{"default", 50, 50, false},
		{"maximum", MaxClaimListLimit, MaxClaimListLimit, false},
		{"zero", 0, 0, true},
		{"negative", -1, 0, true},
		{"too high", MaxClaimListLimit + 1, 0, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NormalizeClaimListLimit(tc.limit)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error for limit %d, got nil", tc.limit)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for limit %d: %v", tc.limit, err)
			}
			if got != tc.want {
				t.Fatalf("NormalizeClaimListLimit(%d) = %d, want %d", tc.limit, got, tc.want)
			}
		})
	}
}

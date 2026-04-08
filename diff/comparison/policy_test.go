package comparison

import "testing"

func TestComparisonPolicy_String(t *testing.T) {
	tests := []struct {
		p    ComparisonPolicy
		want string
	}{
		{PolicyDefault, "DEFAULT"},
		{PolicyTrimWhitespaces, "TRIM_WHITESPACES"},
		{PolicyIgnoreWhitespaces, "IGNORE_WHITESPACES"},
	}
	for _, tt := range tests {
		if got := tt.p.String(); got != tt.want {
			t.Errorf("ComparisonPolicy(%d).String() = %q, want %q", tt.p, got, tt.want)
		}
	}
}

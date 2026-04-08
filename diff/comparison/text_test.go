package comparison

import "testing"

func TestIsEqual_Default(t *testing.T) {
	if !IsEqual("abc", "abc", PolicyDefault) {
		t.Fatal("expected equal")
	}
	if IsEqual("abc", "abc ", PolicyDefault) {
		t.Fatal("expected not equal")
	}
}

func TestIsEqual_TrimWhitespaces(t *testing.T) {
	if !IsEqual("  abc  ", "abc", PolicyTrimWhitespaces) {
		t.Fatal("expected equal after trim")
	}
	if IsEqual("ab c", "abc", PolicyTrimWhitespaces) {
		t.Fatal("trim should not ignore inner whitespace")
	}
}

func TestIsEqual_IgnoreWhitespaces(t *testing.T) {
	if !IsEqual("a b c", "abc", PolicyIgnoreWhitespaces) {
		t.Fatal("expected equal ignoring spaces")
	}
	if !IsEqual("  a  b  ", "ab", PolicyIgnoreWhitespaces) {
		t.Fatal("expected equal ignoring all spaces")
	}
	if IsEqual("abc", "abd", PolicyIgnoreWhitespaces) {
		t.Fatal("expected not equal")
	}
}

func TestHashCode_ConsistentWithPolicy(t *testing.T) {
	// Same strings under same policy should produce same hash
	h1 := HashCode("hello", PolicyDefault)
	h2 := HashCode("hello", PolicyDefault)
	if h1 != h2 {
		t.Fatal("same string should have same hash")
	}

	// Trimmed versions should match under TrimWhitespaces
	h3 := HashCode("  hello  ", PolicyTrimWhitespaces)
	h4 := HashCode("hello", PolicyTrimWhitespaces)
	if h3 != h4 {
		t.Fatal("trimmed strings should have same hash")
	}

	// Ignore whitespace
	h5 := HashCode("h e l l o", PolicyIgnoreWhitespaces)
	h6 := HashCode("hello", PolicyIgnoreWhitespaces)
	if h5 != h6 {
		t.Fatal("strings ignoring whitespace should have same hash")
	}
}

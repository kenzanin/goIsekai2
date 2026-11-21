package plugin

import "testing"

func TestValidateMethod(t *testing.T) {
	for _, method := range AllowedMethods {
		if err := ValidateMethod(method); err != nil {
			t.Errorf("ValidateMethod(%q) = %v, want nil", method, err)
		}
	}
	const unknown = "pages2"
	if err := ValidateMethod(unknown); err == nil {
		t.Errorf("ValidateMethod(%q) = nil, want error", unknown)
	}
}

package domain

import "testing"

func TestSecretTypeValid(t *testing.T) {
	tests := []struct {
		name string
		typ  SecretType
		want bool
	}{
		{name: "login", typ: SecretTypeLogin, want: true},
		{name: "text", typ: SecretTypeText, want: true},
		{name: "binary", typ: SecretTypeBinary, want: true},
		{name: "card", typ: SecretTypeCard, want: true},
		{name: "unknown", typ: SecretType("unknown"), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.typ.Valid(); got != tt.want {
				t.Fatalf("Valid() = %v, want %v", got, tt.want)
			}
		})
	}
}

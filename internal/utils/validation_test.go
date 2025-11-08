package utils

import "testing"

func TestValidateSourceType(t *testing.T) {
	tests := []struct {
		name      string
		sourceType string
		wantErr   bool
	}{
		{"valid game", "game", false},
		{"valid server", "server", false},
		{"valid payment", "payment", false},
		{"invalid type", "invalid", true},
		{"empty string", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSourceType(tt.sourceType)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSourceType() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateState(t *testing.T) {
	tests := []struct {
		name    string
		state   string
		wantErr bool
	}{
		{"valid win", "win", false},
		{"valid lose", "lose", false},
		{"invalid state", "draw", true},
		{"empty string", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateState(tt.state)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateState() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateAmount(t *testing.T) {
	tests := []struct {
		name    string
		amount  string
		wantErr bool
	}{
		{"valid integer", "100", false},
		{"valid one decimal", "10.5", false},
		{"valid two decimals", "10.50", false},
		{"invalid three decimals", "10.500", true},
		{"invalid format", "10.5.5", true},
		{"invalid characters", "abc", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAmount(tt.amount)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAmount() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateUserID(t *testing.T) {
	tests := []struct {
		name    string
		userID  string
		wantErr bool
	}{
		{"valid ID", "1", false},
		{"valid large ID", "999999", false},
		{"invalid zero", "0", true},
		{"invalid negative", "-1", true},
		{"invalid format", "abc", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateUserID(tt.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUserID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFormatBalance(t *testing.T) {
	tests := []struct {
		name     string
		balance  float64
		expected string
	}{
		{"integer", 100.0, "100.00"},
		{"one decimal", 100.5, "100.50"},
		{"two decimals", 100.55, "100.55"},
		{"zero", 0.0, "0.00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatBalance(tt.balance)
			if result != tt.expected {
				t.Errorf("FormatBalance() = %v, want %v", result, tt.expected)
			}
		})
	}
}


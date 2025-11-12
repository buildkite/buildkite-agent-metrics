package cloudfunction

import (
	"os"
	"testing"
)

func TestInitTokenProvider(t *testing.T) {
	// Save original env vars
	origToken := os.Getenv(BKAgentTokensEnvVar)
	origSecret := os.Getenv(BKAgentTokenSecretNamesEnvVar)
	defer func() {
		os.Setenv(BKAgentTokensEnvVar, origToken)
		os.Setenv(BKAgentTokenSecretNamesEnvVar, origSecret)
	}()

	tests := []struct {
		name        string
		tokenEnv    string
		secretEnv   string
		wantErr     bool
		description string
	}{
		{
			name:        "both_env_vars_set",
			tokenEnv:    "token123",
			secretEnv:   "projects/test/secrets/token",
			wantErr:     true,
			description: "Should error when both env vars are set (mutually exclusive)",
		},
		{
			name:        "neither_env_var_set",
			tokenEnv:    "",
			secretEnv:   "",
			wantErr:     true,
			description: "Should error when neither env var is set",
		},
		{
			name:        "single_token_from_env",
			tokenEnv:    "token123",
			secretEnv:   "",
			wantErr:     false,
			description: "Should create one memory provider for single token",
		},
		{
			name:        "token_with_spaces",
			tokenEnv:    " token123 ",
			secretEnv:   "",
			wantErr:     false,
			description: "Should trim spaces from token",
		},
		{
			name:        "empty_token",
			tokenEnv:    "   ",
			secretEnv:   "",
			wantErr:     true,
			description: "Should error when token is only spaces",
		},
		{
			name:        "secret_manager_token",
			tokenEnv:    "",
			secretEnv:   "projects/test/secrets/token/versions/latest",
			wantErr:     false,
			description: "Should create secret manager provider",
		},
		{
			name:        "secret_manager_with_spaces",
			tokenEnv:    "",
			secretEnv:   "  projects/test/secrets/token/versions/latest  ",
			wantErr:     false,
			description: "Should trim spaces from secret name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			os.Setenv(BKAgentTokensEnvVar, tt.tokenEnv)
			os.Setenv(BKAgentTokenSecretNamesEnvVar, tt.secretEnv)

			// Call the function
			provider, err := initTokenProvider()

			// Check error expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("%s: initTokenProvider() error = %v, wantErr %v", tt.description, err, tt.wantErr)
				return
			}

			// If we expect an error, we're done
			if tt.wantErr {
				if provider != nil {
					t.Errorf("%s: Expected nil provider when error occurs, got %v", tt.description, provider)
				}
				return
			}

			// Verify we got a provider
			if provider == nil {
				t.Errorf("%s: Expected non-nil provider, got nil", tt.description)
				return
			}

			// Verify we got a multiTokenProvider (the interface)
			if _, ok := provider.(*multiTokenProvider); !ok {
				t.Errorf("%s: Expected multiTokenProvider, got %T", tt.description, provider)
			}
		})
	}
}

func TestMemoryTokenProvider(t *testing.T) {
	provider := &memoryTokenProvider{
		token: "test-token-123",
	}

	token, err := provider.Get()
	if err != nil {
		t.Errorf("memoryTokenProvider.Get() error = %v", err)
	}

	if token != "test-token-123" {
		t.Errorf("memoryTokenProvider.Get() = %v, want %v", token, "test-token-123")
	}
}

func TestToIntWithDefault(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		defaultValue int
		want         int
		wantErr      bool
	}{
		{
			name:         "empty_string_returns_default",
			input:        "",
			defaultValue: 15,
			want:         15,
			wantErr:      false,
		},
		{
			name:         "valid_integer",
			input:        "42",
			defaultValue: 15,
			want:         42,
			wantErr:      false,
		},
		{
			name:         "invalid_integer",
			input:        "not-a-number",
			defaultValue: 15,
			want:         0,
			wantErr:      true,
		},
		{
			name:         "negative_number",
			input:        "-10",
			defaultValue: 15,
			want:         -10,
			wantErr:      false,
		},
		{
			name:         "zero_value",
			input:        "0",
			defaultValue: 15,
			want:         0,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := toIntWithDefault(tt.input, tt.defaultValue)

			if (err != nil) != tt.wantErr {
				t.Errorf("toIntWithDefault() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("toIntWithDefault() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsQuietMode(t *testing.T) {
	// Save original env var
	orig := os.Getenv("BUILDKITE_QUIET")
	defer os.Setenv("BUILDKITE_QUIET", orig)

	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{"not_set", "", false},
		{"set_to_1", "1", true},
		{"set_to_true", "true", true},
		{"set_to_false", "false", false},
		{"set_to_0", "0", false},
		{"set_to_random", "random", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("BUILDKITE_QUIET", tt.value)
			if got := isQuietMode(); got != tt.want {
				t.Errorf("isQuietMode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsDebugMode(t *testing.T) {
	// Save original env var
	orig := os.Getenv("BUILDKITE_DEBUG")
	defer os.Setenv("BUILDKITE_DEBUG", orig)

	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{"not_set", "", false},
		{"set_to_1", "1", true},
		{"set_to_true", "true", true},
		{"set_to_false", "false", false},
		{"set_to_0", "0", false},
		{"set_to_random", "random", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("BUILDKITE_DEBUG", tt.value)
			if got := isDebugMode(); got != tt.want {
				t.Errorf("isDebugMode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsDebugHTTPMode(t *testing.T) {
	// Save original env var
	orig := os.Getenv("BUILDKITE_AGENT_METRICS_DEBUG_HTTP")
	defer os.Setenv("BUILDKITE_AGENT_METRICS_DEBUG_HTTP", orig)

	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{"not_set", "", false},
		{"set_to_1", "1", true},
		{"set_to_true", "true", true},
		{"set_to_false", "false", false},
		{"set_to_0", "0", false},
		{"set_to_random", "random", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("BUILDKITE_AGENT_METRICS_DEBUG_HTTP", tt.value)
			if got := isDebugHTTPMode(); got != tt.want {
				t.Errorf("isDebugHTTPMode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetEndpoint(t *testing.T) {
	// Save original env var
	orig := os.Getenv(BKAgentEndpointEnvVar)
	defer os.Setenv(BKAgentEndpointEnvVar, orig)

	tests := []struct {
		name  string
		value string
		want  string
	}{
		{
			name:  "default_endpoint",
			value: "",
			want:  "https://agent.buildkite.com/v3",
		},
		{
			name:  "custom_endpoint",
			value: "https://custom.buildkite.com/v3",
			want:  "https://custom.buildkite.com/v3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv(BKAgentEndpointEnvVar, tt.value)
			if got := getEndpoint(); got != tt.want {
				t.Errorf("getEndpoint() = %v, want %v", got, tt.want)
			}
		})
	}
}

package server

import "testing"

func TestAllowedCORSOriginsIncludesLocalDevelopmentClients(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "")

	origins := allowedCORSOrigins()

	for _, want := range []string{
		"http://localhost:5000",
		"http://127.0.0.1:5000",
		"http://localhost:5173",
		"http://127.0.0.1:5173",
		"http://localhost:8010",
		"http://127.0.0.1:8010",
	} {
		if !containsString(origins, want) {
			t.Fatalf("expected CORS origin %q in %v", want, origins)
		}
	}
}

func TestAllowedCORSOriginsIncludesConfiguredOrigins(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://preview.example.com, https://staging.example.com ")

	origins := allowedCORSOrigins()

	for _, want := range []string{
		"https://preview.example.com",
		"https://staging.example.com",
	} {
		if !containsString(origins, want) {
			t.Fatalf("expected configured CORS origin %q in %v", want, origins)
		}
	}
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

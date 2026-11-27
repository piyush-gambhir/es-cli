package config

import "testing"

func TestResolveOverrideCredentialsPickAuthMethod(t *testing.T) {
	for _, k := range []string{"ES_URL", "ES_USERNAME", "ES_PASSWORD", "ES_API_KEY_ID", "ES_API_KEY", "ES_TOKEN"} {
		t.Setenv(k, "")
	}
	profile := &Profile{URL: "http://es:9200", AuthMethod: "basic", Username: "admin", Password: "secret"}

	if rc := Resolve("", "", "", "", "", "", "", false, profile, Defaults{}); rc.AuthMethod != "basic" {
		t.Fatalf("profile only: auth method = %q, want basic", rc.AuthMethod)
	}
	if rc := Resolve("", "", "", "", "", "tok", "", false, profile, Defaults{}); rc.AuthMethod != "bearer" {
		t.Fatalf("--token: auth method = %q, want bearer", rc.AuthMethod)
	}
	t.Setenv("ES_API_KEY", "scoped")
	if rc := Resolve("", "", "", "", "", "", "", false, profile, Defaults{}); rc.AuthMethod != "api_key" {
		t.Fatalf("ES_API_KEY: auth method = %q, want api_key", rc.AuthMethod)
	}
}

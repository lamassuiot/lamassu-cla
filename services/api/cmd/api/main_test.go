package main

import (
	"strings"
	"testing"
	"time"
)

func env(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

var complete = map[string]string{
	"CLA_PORTAL_ORIGIN":           "https://cla.example",
	"CLA_GITHUB_CALLBACK_URL":     "https://cla.example/api/auth/github/callback",
	"CLA_TABLE_EPHEMERAL":         "ephemeral",
	"CLA_TABLE_CORE":              "core",
	"CLA_TABLE_AUDIT":             "audit",
	"CLA_GITHUB_CLIENT_ID":        "client-id",
	"CLA_GITHUB_CLIENT_SECRET_ID": "secret-id",
}

func TestLoadConfigDefaults(t *testing.T) {
	c, err := loadConfig(env(complete))
	if err != nil {
		t.Fatal(err)
	}
	if c.SignInEnabled {
		t.Fatal("sign-in must be disabled by default")
	}
	if c.SessionIdle != 30*time.Minute || c.SessionAbsolute != 8*time.Hour {
		t.Fatalf("session policy = %v/%v", c.SessionIdle, c.SessionAbsolute)
	}
	if len(c.ReturnToAllowlist) != 1 || c.ReturnToAllowlist[0] != "/" || c.SignInErrorRoute != "/sign-in/error" {
		t.Fatalf("defaults = %+v", c)
	}
	if c.GitHubWebBaseURL != "https://github.com" || c.GitHubAPIBaseURL != "https://api.github.com" {
		t.Fatalf("github urls = %q %q", c.GitHubWebBaseURL, c.GitHubAPIBaseURL)
	}
}

func TestLoadConfigOverrides(t *testing.T) {
	m := map[string]string{}
	for k, v := range complete {
		m[k] = v
	}
	m["CLA_SIGN_IN_ENABLED"] = "true"
	m["CLA_SESSION_IDLE"] = "600"
	m["CLA_RETURN_TO_ALLOWLIST"] = "/, /agreements/ ,/profile"
	c, err := loadConfig(env(m))
	if err != nil {
		t.Fatal(err)
	}
	if !c.SignInEnabled || c.SessionIdle != 10*time.Minute || len(c.ReturnToAllowlist) != 3 || c.ReturnToAllowlist[1] != "/agreements/" {
		t.Fatalf("config = %+v", c)
	}
}

func TestLoadConfigMissingAndInvalid(t *testing.T) {
	m := map[string]string{}
	for k, v := range complete {
		m[k] = v
	}
	delete(m, "CLA_TABLE_AUDIT")
	delete(m, "CLA_GITHUB_CLIENT_SECRET_ID")
	_, err := loadConfig(env(m))
	if err == nil || !strings.Contains(err.Error(), "CLA_TABLE_AUDIT") || !strings.Contains(err.Error(), "CLA_GITHUB_CLIENT_SECRET_ID") {
		t.Fatalf("missing variables must be named: %v", err)
	}
	m["CLA_TABLE_AUDIT"], m["CLA_GITHUB_CLIENT_SECRET_ID"] = "audit", "secret-id"
	m["CLA_SESSION_IDLE"] = "soon"
	if _, err := loadConfig(env(m)); err == nil {
		t.Fatal("a non-numeric duration must be rejected")
	}
}

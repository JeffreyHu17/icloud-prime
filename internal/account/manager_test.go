package account

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewManagerLoadsDocumentedAccountsArray(t *testing.T) {
	dir := t.TempDir()
	raw := `{
  "accounts": [
    {
      "id": "acc_1",
      "name": "main",
      "host": "icloud.com",
      "cookies": {
        "X-APPLE-WEBAUTH-TOKEN": "token-value"
      },
      "app_password": "app-pass"
    }
  ]
}`

	if err := os.WriteFile(filepath.Join(dir, "accounts.json"), []byte(raw), 0600); err != nil {
		t.Fatal(err)
	}

	mgr, err := NewManager(dir)
	if err != nil {
		t.Fatal(err)
	}

	acc, ok := mgr.GetAccount("acc_1")
	if !ok {
		t.Fatal("expected account acc_1 to load")
	}
	if got := acc.Cookies["X-APPLE-WEBAUTH-TOKEN"]; got != "token-value" {
		t.Fatalf("expected cookie to load, got %q", got)
	}
	if acc.AppPassword != "app-pass" {
		t.Fatalf("expected app password to load, got %q", acc.AppPassword)
	}
}

func TestNewManagerImportsEditedExampleWhenRuntimeFileMissing(t *testing.T) {
	dir := t.TempDir()
	raw := `{
  "accounts": {
    "acc_1": {
      "id": "acc_1",
      "name": "main",
      "host": "icloud.com",
      "cookies": {
        "X-APPLE-WEBAUTH-TOKEN": "real-token-value",
        "X-APPLE-WEBAUTH-USER": "real-user-value"
      },
      "icloud_email": "user@icloud.com",
      "app_password": "real-app-password"
    },
    "acc_placeholder": {
      "id": "acc_placeholder",
      "name": "Example account",
      "host": "icloud.com",
      "cookies": {
        "X-APPLE-WEBAUTH-TOKEN": "PASTE_YOUR_COOKIE_VALUE_HERE",
        "X-APPLE-WEBAUTH-USER": "PASTE_YOUR_COOKIE_VALUE_HERE",
        "X-APPLE-DS-WEB-SESSION-TOKEN": "PASTE_YOUR_COOKIE_VALUE_HERE"
      },
      "icloud_email": "your_email@icloud.com",
      "app_password": "xxxx-xxxx-xxxx-xxxx",
      "status": "pending"
    }
  }
}`

	if err := os.WriteFile(filepath.Join(dir, "accounts.example.json"), []byte(raw), 0600); err != nil {
		t.Fatal(err)
	}

	mgr, err := NewManager(dir)
	if err != nil {
		t.Fatal(err)
	}

	acc, ok := mgr.GetAccount("acc_1")
	if !ok {
		t.Fatal("expected edited accounts.example.json to be imported")
	}
	if got := acc.Cookies["X-APPLE-WEBAUTH-TOKEN"]; got != "real-token-value" {
		t.Fatalf("expected imported cookie, got %q", got)
	}
	if _, ok := mgr.GetAccount("acc_placeholder"); ok {
		t.Fatal("expected placeholder example account to be ignored")
	}
	if _, err := os.Stat(filepath.Join(dir, "accounts.json")); err != nil {
		t.Fatalf("expected imported accounts to be saved to accounts.json: %v", err)
	}
}

func TestNewManagerLoadsBrowserCookieExportAndLegacyAppPasswords(t *testing.T) {
	dir := t.TempDir()
	raw := `{
  "accounts": {
    "acc_1": {
      "id": "acc_1",
      "name": "main",
      "host": "icloud.com",
      "cookies": [
        {
          "name": "X-APPLE-WEBAUTH-TOKEN",
          "value": "token-value"
        }
      ],
      "app_passwords": [
        {
          "icloud_email": "user@icloud.com",
          "password": "app-pass"
        }
      ]
    }
  }
}`

	if err := os.WriteFile(filepath.Join(dir, "accounts.json"), []byte(raw), 0600); err != nil {
		t.Fatal(err)
	}

	mgr, err := NewManager(dir)
	if err != nil {
		t.Fatal(err)
	}

	acc, ok := mgr.GetAccount("acc_1")
	if !ok {
		t.Fatal("expected account acc_1 to load")
	}
	if got := acc.Cookies["X-APPLE-WEBAUTH-TOKEN"]; got != "token-value" {
		t.Fatalf("expected browser cookie export to load, got %q", got)
	}
	if acc.ICloudEmail != "user@icloud.com" {
		t.Fatalf("expected icloud email to load, got %q", acc.ICloudEmail)
	}
	if acc.AppPassword != "app-pass" {
		t.Fatalf("expected legacy app password to load, got %q", acc.AppPassword)
	}
}

func TestListAccountsRedactsSensitiveFields(t *testing.T) {
	dir := t.TempDir()
	raw := `{
  "accounts": {
    "acc_1": {
      "id": "acc_1",
      "name": "main",
      "cookies": {
        "X-APPLE-WEBAUTH-TOKEN": "token-value"
      },
      "app_password": "app-pass"
    }
  }
}`

	if err := os.WriteFile(filepath.Join(dir, "accounts.json"), []byte(raw), 0600); err != nil {
		t.Fatal(err)
	}

	mgr, err := NewManager(dir)
	if err != nil {
		t.Fatal(err)
	}

	accounts := mgr.ListAccounts()
	if len(accounts) != 1 {
		t.Fatalf("expected one account, got %d", len(accounts))
	}
	if accounts[0].Cookies != nil {
		t.Fatal("expected cookies to be redacted")
	}
	if accounts[0].AppPassword != "" {
		t.Fatal("expected app password to be redacted")
	}
}

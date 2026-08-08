package mail

import "testing"

func TestNewWebClientNormalizesAppleMailHost(t *testing.T) {
	client := NewWebClient(nil, "", "imap.mail.me.com")

	if client.host != "icloud.com" {
		t.Fatalf("expected host to normalize to icloud.com, got %q", client.host)
	}
}

func TestNewWebClientNormalizesChinaICloudHost(t *testing.T) {
	client := NewWebClient(nil, "", "https://www.icloud.com.cn")

	if client.host != "icloud.com.cn" {
		t.Fatalf("expected host to normalize to icloud.com.cn, got %q", client.host)
	}
}

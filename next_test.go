package aria2rpc

import (
	"testing"
)

func startAria2NextForTest(t *testing.T, secret, scheme string, opts ...Option) *Client {
	t.Helper()
	return startAria2ForTestShared(t, "aria2-next", secret, scheme, opts)
}

func TestNext_SetBtPeerBlocklist(t *testing.T) {
	ctx := t.Context()

	c := startAria2NextForTest(t, secret, "http")
	defer c.Close()

	rules := []string{
		"192.168.1.0/24",
		"10.0.0.1",
		"172.16.0.0/16",
		"8.8.8.8",
	}

	result, err := c.SetBtPeerBlocklist(ctx, rules)
	if err != nil {
		t.Fatalf("SetBtPeerBlocklist failed: %v", err)
	}
	if result == nil {
		t.Fatal("SetBtPeerBlocklist returned nil result")
	}
	if result.RuleCount == 0 {
		t.Fatal("SetBtPeerBlocklist returned ruleCount=0")
	}
	if result.Revision == 0 {
		t.Fatal("SetBtPeerBlocklist returned revision=0")
	}

	t.Logf("SetBtPeerBlocklist: ruleCount=%d, revision=%d", result.RuleCount, result.Revision)
}

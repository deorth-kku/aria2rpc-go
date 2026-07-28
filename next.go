package aria2rpc

import (
	"context"
)

// BtPeerBlocklistResult is the result of setBtPeerBlocklist.
type BtPeerBlocklistResult struct {
	RuleCount int `json:"ruleCount"`
	Revision  int `json:"revision"`
}

type nextClient struct {
	SetBtPeerBlocklist func(context.Context, string, []string) (*BtPeerBlocklistResult, error) `rpc_method:"aria2.setBtPeerBlocklist"`
}

// SetBtPeerBlocklist sets the BitTorrent peer blocklist rules.
func (c *Client) SetBtPeerBlocklist(ctx context.Context, rules []string) (*BtPeerBlocklistResult, error) {
	return c.next.SetBtPeerBlocklist(ctx, c.secret, rules)
}

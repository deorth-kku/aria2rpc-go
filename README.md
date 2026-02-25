# THIS IS A VIBE CODING ARTIFACT

`aria2rpc-go` is a Go JSON-RPC client for aria2, built on top of `github.com/filecoin-project/go-jsonrpc`.

## Status

This project is currently focused on practical compatibility with `github.com/siku2/arigo`-style usage.

## Key Design

- The low-level RPC function signatures use fixed parameter counts.
- Every low-level `aria2.*` call includes `secret` as the first parameter.
- The public API does not expose `secret` in method signatures.
- If no secret is configured, an empty value is passed.
- Optional aria2 parameters are handled by dispatching to multiple fixed-signature bindings.

## WebSocket Notifications

Supported callbacks include:

- `onDownloadStart`
- `onDownloadPause`
- `onDownloadStop`
- `onDownloadComplete`
- `onDownloadError`
- `onBtDownloadComplete`

To align with aria2 WebSocket behavior and reduce noisy `id:null` response logs, client ping is disabled by default.

## Testing

Tests use a real local `aria2c` process when available.

- Default integration tests: `GOEXPERIMENT=jsonv2 go test ./...`
- Slow BT completion test:
  - `I_HAVE_A_LOT_OF_MEMORY_AND_TIME=1 GOEXPERIMENT=jsonv2 go test ./... -run TestAria2WS_OnBtDownloadComplete -v`

If `aria2c` is not installed, integration tests are skipped.

## Quick Start

```go
package main

import (
	"context"
	"log"

	"github.com/deorth-kku/aria2rpc-go"
)

func main() {
	ctx := context.Background()
	c, err := aria2rpc.New(ctx, "http://127.0.0.1:6800/jsonrpc", aria2rpc.WithSecret("my-secret"))
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	gid, err := c.AddURI(ctx, []string{"https://example.com/file.iso"}, nil, nil)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("gid:", gid)
}
```

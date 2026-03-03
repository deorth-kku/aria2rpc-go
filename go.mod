module github.com/deorth-kku/aria2rpc-go

go 1.25.0

require github.com/filecoin-project/go-jsonrpc v0.10.1

require (
	github.com/deorth-kku/go-common v0.0.0-20260130130410-826603dc6e46 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
)

replace github.com/filecoin-project/go-jsonrpc v0.10.1 => ../go-jsonrpc

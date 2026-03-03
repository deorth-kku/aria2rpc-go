package aria2rpc

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/filecoin-project/go-jsonrpc"
)

// Option configures client construction.
type Option func(*config)

type config struct {
	secret       string
	headers      http.Header
	rpcOpts      []jsonrpc.Option
	callbacks    NotificationCallbacks
	hasCallbacks bool
}

// WithSecret sets the aria2 RPC secret token. Both "secret" and "token:secret" are accepted.
func WithSecret(secret string) Option {
	return func(c *config) {
		c.secret = normalizeSecret(secret)
	}
}

// WithHeader sets custom HTTP headers used by go-jsonrpc.
func WithHeader(headers http.Header) Option {
	return func(c *config) {
		c.headers = headers
	}
}

// WithJSONRPCOptions appends raw go-jsonrpc options.
func WithJSONRPCOptions(opts ...jsonrpc.Option) Option {
	return func(c *config) {
		c.rpcOpts = append(c.rpcOpts, opts...)
	}
}

// NotificationCallbacks are callbacks for aria2 websocket notifications.
// These callbacks are only effective when using ws:// or wss:// endpoint.
type NotificationCallbacks struct {
	OnDownloadStart      func(context.Context, DownloadEvent)
	OnDownloadPause      func(context.Context, DownloadEvent)
	OnDownloadStop       func(context.Context, DownloadEvent)
	OnDownloadComplete   func(context.Context, DownloadEvent)
	OnDownloadError      func(context.Context, DownloadEvent)
	OnBtDownloadComplete func(context.Context, DownloadEvent)
}

// WithNotificationCallbacks sets websocket notification callbacks for aria2.onDownload* events.
func WithNotificationCallbacks(callbacks NotificationCallbacks) Option {
	return func(c *config) {
		c.callbacks = callbacks
		c.hasCallbacks = true
	}
}

// Client is an aria2 JSON-RPC client.
type Client struct {
	secret string
	raw    rawClient
	close  jsonrpc.ClientCloser

	cbMu      sync.RWMutex
	callbacks NotificationCallbacks
}

// New creates an aria2 RPC client at addr, for example: http://127.0.0.1:6800/jsonrpc.
func New(ctx context.Context, addr string, opts ...Option) (*Client, error) {
	cfg := &config{}
	for _, opt := range opts {
		if opt != nil {
			opt(cfg)
		}
	}

	c := &Client{secret: cfg.secret}
	if cfg.hasCallbacks {
		c.SetNotificationCallbacks(cfg.callbacks)
		cfg.rpcOpts = append(cfg.rpcOpts, c.wsNotificationOptions()...)
	}
	// aria2 websocket endpoint may emit id:null error responses for control frames.
	// Disable client ping by default to avoid periodic noise from that behavior.
	allOpts := make([]jsonrpc.Option, 0, 1+len(cfg.rpcOpts))
	allOpts = append(allOpts, jsonrpc.WithPingInterval(0))
	allOpts = append(allOpts, cfg.rpcOpts...)

	closer, err := jsonrpc.NewClient(ctx, addr, "", &c.raw, cfg.headers, allOpts...)
	if err != nil {
		return nil, err
	}
	c.close = closer
	return c, nil
}

// Close closes the underlying go-jsonrpc client connection.
func (c *Client) Close() {
	if c == nil || c.close == nil {
		return
	}
	c.close()
}

// SetNotificationCallbacks updates websocket notification callbacks at runtime.
func (c *Client) SetNotificationCallbacks(callbacks NotificationCallbacks) {
	if c == nil {
		return
	}
	c.cbMu.Lock()
	c.callbacks = callbacks
	c.cbMu.Unlock()
}

func (c *Client) getNotificationCallbacks() NotificationCallbacks {
	c.cbMu.RLock()
	defer c.cbMu.RUnlock()
	return c.callbacks
}

func (c *Client) wsNotificationOptions() []jsonrpc.Option {
	handler := &wsNotificationHandler{client: c}
	return []jsonrpc.Option{
		jsonrpc.WithClientHandler("aria2", handler),
		jsonrpc.WithClientHandlerAlias("aria2.onDownloadStart", "aria2.OnDownloadStart"),
		jsonrpc.WithClientHandlerAlias("aria2.onDownloadPause", "aria2.OnDownloadPause"),
		jsonrpc.WithClientHandlerAlias("aria2.onDownloadStop", "aria2.OnDownloadStop"),
		jsonrpc.WithClientHandlerAlias("aria2.onDownloadComplete", "aria2.OnDownloadComplete"),
		jsonrpc.WithClientHandlerAlias("aria2.onDownloadError", "aria2.OnDownloadError"),
		jsonrpc.WithClientHandlerAlias("aria2.onBtDownloadComplete", "aria2.OnBtDownloadComplete"),
	}
}

type wsNotificationHandler struct {
	client *Client
}

func (h *wsNotificationHandler) OnDownloadStart(ctx context.Context, event DownloadEvent) {
	if fn := h.client.getNotificationCallbacks().OnDownloadStart; fn != nil {
		fn(ctx, event)
	}
}

func (h *wsNotificationHandler) OnDownloadPause(ctx context.Context, event DownloadEvent) {
	if fn := h.client.getNotificationCallbacks().OnDownloadPause; fn != nil {
		fn(ctx, event)
	}
}

func (h *wsNotificationHandler) OnDownloadStop(ctx context.Context, event DownloadEvent) {
	if fn := h.client.getNotificationCallbacks().OnDownloadStop; fn != nil {
		fn(ctx, event)
	}
}

func (h *wsNotificationHandler) OnDownloadComplete(ctx context.Context, event DownloadEvent) {
	if fn := h.client.getNotificationCallbacks().OnDownloadComplete; fn != nil {
		fn(ctx, event)
	}
}

func (h *wsNotificationHandler) OnDownloadError(ctx context.Context, event DownloadEvent) {
	if fn := h.client.getNotificationCallbacks().OnDownloadError; fn != nil {
		fn(ctx, event)
	}
}

func (h *wsNotificationHandler) OnBtDownloadComplete(ctx context.Context, event DownloadEvent) {
	if fn := h.client.getNotificationCallbacks().OnBtDownloadComplete; fn != nil {
		fn(ctx, event)
	}
}

func normalizeSecret(secret string) string {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return ""
	}
	if strings.HasPrefix(secret, "token:") {
		return secret
	}
	return "token:" + secret
}

func map2Optional[M ~map[K]V, K comparable, V any](m M) jsonrpc.Optional[M] {
	if m == nil {
		return nil
	}
	return &m
}

func slice2Optional[S ~[]E, E any](s S) jsonrpc.Optional[S] {
	if s == nil {
		return nil
	}
	return &s
}

func (c *Client) AddURI(ctx context.Context, uris []string, options map[string]string, position *int) (string, error) {
	if len(uris) == 0 {
		return "", errors.New("uris must not be empty")
	}

	if position != nil && options == nil {
		options = map[string]string{}
	}

	return c.raw.AddURI(ctx, c.secret, uris, map2Optional(options), position)
}

func (c *Client) AddTorrent(ctx context.Context, torrent string, uris []string, options map[string]string, position *int) (string, error) {
	if torrent == "" {
		return "", errors.New("torrent must not be empty")
	}

	if position != nil && options == nil {
		options = map[string]string{}
	}

	urisArg := uris
	switch {
	case len(urisArg) == 0 && options == nil && position == nil:
		urisArg = nil
	case len(urisArg) == 0 && (options != nil || position != nil):
		urisArg = []string{}
	}

	return c.raw.AddTorrent(ctx, c.secret, torrent, slice2Optional(urisArg), map2Optional(options), position)
}

func (c *Client) AddMetalink(ctx context.Context, metalink string, options map[string]string, position *int) (string, error) {
	if metalink == "" {
		return "", errors.New("metalink must not be empty")
	}

	if position != nil && options == nil {
		options = map[string]string{}
	}

	return c.raw.AddMetalink(ctx, c.secret, metalink, map2Optional(options), position)
}

func (c *Client) Remove(ctx context.Context, gid string) (string, error) {
	return c.raw.Remove(ctx, c.secret, gid)
}

func (c *Client) ForceRemove(ctx context.Context, gid string) (string, error) {
	return c.raw.ForceRemove(ctx, c.secret, gid)
}

func (c *Client) Pause(ctx context.Context, gid string) (string, error) {
	return c.raw.Pause(ctx, c.secret, gid)
}

func (c *Client) PauseAll(ctx context.Context) (string, error) {
	return c.raw.PauseAll(ctx, c.secret)
}

func (c *Client) ForcePause(ctx context.Context, gid string) (string, error) {
	return c.raw.ForcePause(ctx, c.secret, gid)
}

func (c *Client) ForcePauseAll(ctx context.Context) (string, error) {
	return c.raw.ForcePauseAll(ctx, c.secret)
}

func (c *Client) Unpause(ctx context.Context, gid string) (string, error) {
	return c.raw.Unpause(ctx, c.secret, gid)
}

func (c *Client) UnpauseAll(ctx context.Context) (string, error) {
	return c.raw.UnpauseAll(ctx, c.secret)
}

func (c *Client) TellStatus(ctx context.Context, gid string, keys ...string) (*Status, error) {
	if gid == "" {
		return nil, errors.New("gid must not be empty")
	}

	return c.raw.TellStatus(ctx, c.secret, gid, slice2Optional(keys))
}

func (c *Client) GetURIs(ctx context.Context, gid string) ([]URIStatus, error) {
	return c.raw.GetURIs(ctx, c.secret, gid)
}

func (c *Client) GetFiles(ctx context.Context, gid string) ([]FileInfo, error) {
	return c.raw.GetFiles(ctx, c.secret, gid)
}

func (c *Client) GetPeers(ctx context.Context, gid string) ([]PeerInfo, error) {
	return c.raw.GetPeers(ctx, c.secret, gid)
}

func (c *Client) GetServers(ctx context.Context, gid string) ([]ServerInfo, error) {
	return c.raw.GetServers(ctx, c.secret, gid)
}

func (c *Client) TellActive(ctx context.Context, keys ...string) ([]*Status, error) {
	return c.raw.TellActive(ctx, c.secret, slice2Optional(keys))
}

func (c *Client) TellWaiting(ctx context.Context, offset, num int, keys ...string) ([]*Status, error) {
	return c.raw.TellWaiting(ctx, c.secret, offset, num, slice2Optional(keys))
}

func (c *Client) TellStopped(ctx context.Context, offset, num int, keys ...string) ([]*Status, error) {
	return c.raw.TellStopped(ctx, c.secret, offset, num, slice2Optional(keys))
}

func (c *Client) ChangePosition(ctx context.Context, gid string, pos int, how string) (int, error) {
	return c.raw.ChangePosition(ctx, c.secret, gid, pos, how)
}

func (c *Client) ChangeURI(ctx context.Context, gid string, fileIndex int, delURIs, addURIs []string, position *int) ([]int, error) {
	var optPosition jsonrpc.Optional[int]
	if position != nil {
		optPosition = position
	}
	return c.raw.ChangeURI(ctx, c.secret, gid, fileIndex, delURIs, addURIs, optPosition)
}

func (c *Client) GetOption(ctx context.Context, gid string) (map[string]string, error) {
	return c.raw.GetOption(ctx, c.secret, gid)
}

func (c *Client) ChangeOption(ctx context.Context, gid string, options map[string]string) (string, error) {
	return c.raw.ChangeOption(ctx, c.secret, gid, options)
}

func (c *Client) GetGlobalOption(ctx context.Context) (map[string]string, error) {
	return c.raw.GetGlobalOption(ctx, c.secret)
}

func (c *Client) ChangeGlobalOption(ctx context.Context, options map[string]string) (string, error) {
	return c.raw.ChangeGlobalOption(ctx, c.secret, options)
}

func (c *Client) GetGlobalStat(ctx context.Context) (*GlobalStat, error) {
	return c.raw.GetGlobalStat(ctx, c.secret)
}

func (c *Client) PurgeDownloadResult(ctx context.Context) (string, error) {
	return c.raw.PurgeDownloadResult(ctx, c.secret)
}

func (c *Client) RemoveDownloadResult(ctx context.Context, gid string) (string, error) {
	return c.raw.RemoveDownloadResult(ctx, c.secret, gid)
}

func (c *Client) GetVersion(ctx context.Context) (*VersionInfo, error) {
	return c.raw.GetVersion(ctx, c.secret)
}

func (c *Client) GetSessionInfo(ctx context.Context) (*SessionInfo, error) {
	return c.raw.GetSessionInfo(ctx, c.secret)
}

func (c *Client) Shutdown(ctx context.Context) (string, error) {
	return c.raw.Shutdown(ctx, c.secret)
}

func (c *Client) ForceShutdown(ctx context.Context) (string, error) {
	return c.raw.ForceShutdown(ctx, c.secret)
}

func (c *Client) SaveSession(ctx context.Context) (string, error) {
	return c.raw.SaveSession(ctx, c.secret)
}

func (c *Client) ListMethods(ctx context.Context) ([]string, error) {
	return c.raw.ListMethods(ctx)
}

func (c *Client) ListNotifications(ctx context.Context) ([]string, error) {
	return c.raw.ListNotifications(ctx)
}

func (c *Client) Multicall(ctx context.Context, calls []Multicall) ([][]string, error) {
	if len(calls) == 0 {
		return nil, errors.New("calls must not be empty")
	}

	rawCalls := make([]map[string]any, 0, len(calls))
	for i, mc := range calls {
		if mc.MethodName == "" {
			return nil, fmt.Errorf("calls[%d].methodName is empty", i)
		}
		params := append([]any(nil), mc.Params...)
		if strings.HasPrefix(mc.MethodName, "aria2.") {
			params = append([]any{c.secret}, params...)
		}
		rawCalls = append(rawCalls, map[string]any{
			"methodName": mc.MethodName,
			"params":     params,
		})
	}

	return c.raw.Multicall(ctx, rawCalls)
}

// rawClient uses jsonrpc.Optional for trailing optional RPC params.
type rawClient struct {
	AddURI      func(context.Context, string, []string, jsonrpc.Optional[map[string]string], jsonrpc.Optional[int]) (string, error)                           `rpc_method:"aria2.addUri"`
	AddTorrent  func(context.Context, string, string, jsonrpc.Optional[[]string], jsonrpc.Optional[map[string]string], jsonrpc.Optional[int]) (string, error) `rpc_method:"aria2.addTorrent"`
	AddMetalink func(context.Context, string, string, jsonrpc.Optional[map[string]string], jsonrpc.Optional[int]) (string, error)                             `rpc_method:"aria2.addMetalink"`

	Remove               func(context.Context, string, string) (string, error)                                                `rpc_method:"aria2.remove"`
	ForceRemove          func(context.Context, string, string) (string, error)                                                `rpc_method:"aria2.forceRemove"`
	Pause                func(context.Context, string, string) (string, error)                                                `rpc_method:"aria2.pause"`
	PauseAll             func(context.Context, string) (string, error)                                                        `rpc_method:"aria2.pauseAll"`
	ForcePause           func(context.Context, string, string) (string, error)                                                `rpc_method:"aria2.forcePause"`
	ForcePauseAll        func(context.Context, string) (string, error)                                                        `rpc_method:"aria2.forcePauseAll"`
	Unpause              func(context.Context, string, string) (string, error)                                                `rpc_method:"aria2.unpause"`
	UnpauseAll           func(context.Context, string) (string, error)                                                        `rpc_method:"aria2.unpauseAll"`
	TellStatus           func(context.Context, string, string, jsonrpc.Optional[[]string]) (*Status, error)                   `rpc_method:"aria2.tellStatus"`
	GetURIs              func(context.Context, string, string) ([]URIStatus, error)                                           `rpc_method:"aria2.getUris"`
	GetFiles             func(context.Context, string, string) ([]FileInfo, error)                                            `rpc_method:"aria2.getFiles"`
	GetPeers             func(context.Context, string, string) ([]PeerInfo, error)                                            `rpc_method:"aria2.getPeers"`
	GetServers           func(context.Context, string, string) ([]ServerInfo, error)                                          `rpc_method:"aria2.getServers"`
	TellActive           func(context.Context, string, jsonrpc.Optional[[]string]) ([]*Status, error)                         `rpc_method:"aria2.tellActive"`
	TellWaiting          func(context.Context, string, int, int, jsonrpc.Optional[[]string]) ([]*Status, error)               `rpc_method:"aria2.tellWaiting"`
	TellStopped          func(context.Context, string, int, int, jsonrpc.Optional[[]string]) ([]*Status, error)               `rpc_method:"aria2.tellStopped"`
	ChangePosition       func(context.Context, string, string, int, string) (int, error)                                      `rpc_method:"aria2.changePosition"`
	ChangeURI            func(context.Context, string, string, int, []string, []string, jsonrpc.Optional[int]) ([]int, error) `rpc_method:"aria2.changeUri"`
	GetOption            func(context.Context, string, string) (map[string]string, error)                                     `rpc_method:"aria2.getOption"`
	ChangeOption         func(context.Context, string, string, map[string]string) (string, error)                             `rpc_method:"aria2.changeOption"`
	GetGlobalOption      func(context.Context, string) (map[string]string, error)                                             `rpc_method:"aria2.getGlobalOption"`
	ChangeGlobalOption   func(context.Context, string, map[string]string) (string, error)                                     `rpc_method:"aria2.changeGlobalOption"`
	GetGlobalStat        func(context.Context, string) (*GlobalStat, error)                                                   `rpc_method:"aria2.getGlobalStat"`
	PurgeDownloadResult  func(context.Context, string) (string, error)                                                        `rpc_method:"aria2.purgeDownloadResult"`
	RemoveDownloadResult func(context.Context, string, string) (string, error)                                                `rpc_method:"aria2.removeDownloadResult"`
	GetVersion           func(context.Context, string) (*VersionInfo, error)                                                  `rpc_method:"aria2.getVersion"`
	GetSessionInfo       func(context.Context, string) (*SessionInfo, error)                                                  `rpc_method:"aria2.getSessionInfo"`
	Shutdown             func(context.Context, string) (string, error)                                                        `rpc_method:"aria2.shutdown"`
	ForceShutdown        func(context.Context, string) (string, error)                                                        `rpc_method:"aria2.forceShutdown"`
	SaveSession          func(context.Context, string) (string, error)                                                        `rpc_method:"aria2.saveSession"`

	Multicall         func(context.Context, []map[string]any) ([][]string, error) `rpc_method:"system.multicall"`
	ListMethods       func(context.Context) ([]string, error)                     `rpc_method:"system.listMethods"`
	ListNotifications func(context.Context) ([]string, error)                     `rpc_method:"system.listNotifications"`
}

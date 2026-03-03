package aria2rpc

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

const secret = "abc123"

func TestAria2_AddURIAndTellStatus(t *testing.T) {
	ctx := context.Background()

	c := startAria2ForTest(t, secret, "http")
	defer c.Close()

	payload := []byte("hello from aria2rpc integration test")
	src := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(payload)
	}))
	defer src.Close()

	opts := map[string]string{"out": "it.txt"}
	pos := 0
	gid, err := c.AddURI(ctx, []string{src.URL}, opts, &pos)
	if err != nil {
		t.Fatalf("AddURI failed: %v", err)
	}
	if gid == "" {
		t.Fatal("AddURI returned empty gid")
	}

	st := waitStatusDone(t, c, gid, 8*time.Second)
	if st == nil {
		t.Fatal("nil status")
	}
	if st.GID != gid {
		t.Fatalf("status gid mismatch: got %q want %q", st.GID, gid)
	}

	st2, err := c.TellStatus(ctx, gid, "gid", "status")
	if err != nil {
		t.Fatalf("TellStatus(keys...) failed: %v", err)
	}
	if st2 == nil || st2.GID == "" {
		t.Fatalf("TellStatus(keys...) returned invalid status: %#v", st2)
	}

	files, err := c.GetFiles(ctx, gid)
	if err != nil {
		t.Fatalf("GetFiles failed: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("GetFiles returned empty list")
	}
	if got := filepath.Base(files[0].Path); got != "it.txt" {
		t.Fatalf("unexpected output file name: %q", got)
	}
}

func TestAria2WS_OnDownloadCallbacks(t *testing.T) {
	ctx := context.Background()

	startCh := make(chan string, 4)
	completeCh := make(chan string, 4)

	c := startAria2ForTest(t, secret, "ws", WithNotificationCallbacks(NotificationCallbacks{
		OnDownloadStart: func(_ context.Context, event DownloadEvent) {
			startCh <- event.GID
		},
		OnDownloadComplete: func(_ context.Context, event DownloadEvent) {
			completeCh <- event.GID
		},
	}))
	defer c.Close()

	payload := []byte("hello from aria2rpc ws callback test")
	src := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(payload)
	}))
	defer src.Close()

	gid, err := c.AddURI(ctx, []string{src.URL}, map[string]string{"out": "ws-it.txt"}, nil)
	if err != nil {
		t.Fatalf("AddURI failed: %v", err)
	}
	waitStatusDone(t, c, gid, 8*time.Second)

	waitEventGID(t, startCh, gid, "onDownloadStart", 4*time.Second)
	waitEventGID(t, completeCh, gid, "onDownloadComplete", 4*time.Second)
}

var torrentURLs = []string{
	"https://releases.ubuntu.com/25.10/ubuntu-25.10-live-server-amd64.iso.torrent",
	"https://cdimage.debian.org/debian-cd/current/amd64/bt-cd/debian-13.1.0-amd64-netinst.iso.torrent",
	"https://cdimage.debian.org/debian-cd/current/amd64/bt-cd/debian-13.1.0-amd64-xfce-CD-1.iso.torrent",
}

func TestAria2WS_BTFromPublicTorrents(t *testing.T) {
	ctx := context.Background()

	startCh := make(chan string, 8)
	c := startAria2ForTest(t, secret, "ws", WithNotificationCallbacks(NotificationCallbacks{
		OnDownloadStart: func(_ context.Context, event DownloadEvent) {
			startCh <- event.GID
		},
	}))
	defer c.Close()

	var lastErr error
	for _, torrentURL := range torrentURLs {
		content, err := fetchTorrent(ctx, torrentURL)
		if err != nil {
			lastErr = fmt.Errorf("fetch %s: %w", torrentURL, err)
			continue
		}
		torrentB64 := base64.StdEncoding.EncodeToString(content)

		gid, err := c.AddTorrent(ctx, torrentB64, nil, map[string]string{"seed-time": "0"}, nil)
		if err != nil {
			lastErr = fmt.Errorf("add torrent %s: %w", torrentURL, err)
			continue
		}

		waitEventGID(t, startCh, gid, "onDownloadStart", 20*time.Second)
		st, err := c.TellStatus(ctx, gid, "gid", "status", "bittorrent", "errorCode", "errorMessage")
		if err != nil {
			t.Fatalf("TellStatus after add torrent failed: %v", err)
		}
		if st == nil {
			t.Fatal("TellStatus returned nil status for torrent")
		}
		if st.GID != gid {
			t.Fatalf("torrent status gid mismatch: got %q want %q", st.GID, gid)
		}
		return
	}

	t.Fatalf("all public torrent sources failed, last error: %v", lastErr)
}

func TestAria2WS_OnBtDownloadComplete(t *testing.T) {
	if os.Getenv("I_HAVE_A_LOT_OF_MEMORY_AND_TIME") != "1" {
		t.Skip("skipping slow bt-complete test, set I_HAVE_A_LOT_OF_MEMORY_AND_TIME=1 to run")
	}

	ctx := context.Background()

	btCompleteCh := make(chan string, 8)
	c := startAria2ForTest(t, secret, "ws", WithNotificationCallbacks(NotificationCallbacks{
		OnBtDownloadComplete: func(_ context.Context, event DownloadEvent) {
			btCompleteCh <- event.GID
		},
	}))
	defer c.Close()

	var lastErr error
	for _, torrentURL := range torrentURLs {
		content, err := fetchTorrent(ctx, torrentURL)
		if err != nil {
			lastErr = fmt.Errorf("fetch %s: %w", torrentURL, err)
			continue
		}
		torrentB64 := base64.StdEncoding.EncodeToString(content)

		gid, err := c.AddTorrent(ctx, torrentB64, nil, map[string]string{"seed-time": "0"}, nil)
		if err != nil {
			lastErr = fmt.Errorf("add torrent %s: %w", torrentURL, err)
			continue
		}

		if waitEventGIDWithTimeout(btCompleteCh, gid, 15*time.Minute) {
			return
		}

		_, _ = c.ForceRemove(ctx, gid)
		lastErr = fmt.Errorf("timeout waiting onBtDownloadComplete for %s", torrentURL)
	}

	t.Fatalf("onBtDownloadComplete not observed from all public torrents, last error: %v", lastErr)
}

func startAria2ForTest(t *testing.T, secret, scheme string, opts ...Option) *Client {
	t.Helper()

	if _, err := exec.LookPath("aria2c"); err != nil {
		t.Skip("aria2c not found in PATH")
	}

	port := pickFreePort(t)
	dir := t.TempDir()

	cmd := exec.Command(
		"aria2c",
		"--enable-rpc=true",
		"--rpc-listen-all=false",
		"--rpc-listen-port="+strconv.Itoa(port),
		"--rpc-secret="+secret,
		"--rpc-allow-origin-all=true",
		"--max-concurrent-downloads=1",
		"--summary-interval=0",
		"--check-certificate=false",
		"-d", dir,
	)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		t.Fatalf("stderr pipe: %v", err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatalf("start aria2c: %v", err)
	}
	t.Cleanup(func() {
		_ = cmd.Process.Kill()
		_, _ = cmd.Process.Wait()
	})

	addr := fmt.Sprintf("%s://127.0.0.1:%d/jsonrpc", scheme, port)
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()

	var c *Client
	for {
		if ctx.Err() != nil {
			b, _ := io.ReadAll(stderr)
			t.Fatalf("aria2 rpc not ready: %v, stderr=%s", ctx.Err(), string(b))
		}

		baseOpts := []Option{WithSecret(secret)}
		baseOpts = append(baseOpts, opts...)
		c, err = New(context.Background(), addr, baseOpts...)
		if err == nil {
			if _, e := c.GetVersion(context.Background()); e == nil {
				break
			}
			c.Close()
		}
		time.Sleep(150 * time.Millisecond)
	}

	t.Cleanup(func() {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer shutdownCancel()
		_, _ = c.ForceShutdown(shutdownCtx)
	})

	return c
}

func fetchTorrent(ctx context.Context, rawURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{Timeout: 12 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}
	return io.ReadAll(io.LimitReader(resp.Body, 2<<20))
}

func pickFreePort(t *testing.T) int {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen on random port: %v", err)
	}
	defer ln.Close()
	addr, ok := ln.Addr().(*net.TCPAddr)
	if !ok {
		t.Fatalf("unexpected addr type: %T", ln.Addr())
	}
	return addr.Port
}

func waitStatusDone(t *testing.T, c *Client, gid string, timeout time.Duration) *Status {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for {
		st, err := c.TellStatus(context.Background(), gid, "gid", "status", "errorCode", "errorMessage")
		if err == nil {
			if st.Status == "complete" {
				return st
			}
			if st.Status == "error" || st.Status == "removed" {
				t.Fatalf("download failed: status=%s code=%s msg=%s", st.Status, st.ErrorCode, st.ErrorMessage)
			}
		}

		if ctx.Err() != nil {
			if err != nil && !errors.Is(err, context.DeadlineExceeded) {
				t.Fatalf("timeout waiting status, last tellStatus err=%v", err)
			}
			t.Fatalf("timeout waiting status complete for gid=%s", gid)
		}
		time.Sleep(120 * time.Millisecond)
	}
}

func waitEventGID(t *testing.T, ch <-chan string, wantGID, event string, timeout time.Duration) {
	t.Helper()
	deadline := time.NewTimer(timeout)
	defer deadline.Stop()

	for {
		select {
		case got := <-ch:
			if got == wantGID {
				return
			}
		case <-deadline.C:
			t.Fatalf("timeout waiting %s callback for gid=%s", event, wantGID)
		}
	}
}

func waitEventGIDWithTimeout(ch <-chan string, wantGID string, timeout time.Duration) bool {
	deadline := time.NewTimer(timeout)
	defer deadline.Stop()

	for {
		select {
		case got := <-ch:
			if got == wantGID {
				return true
			}
		case <-deadline.C:
			return false
		}
	}
}

package gateway

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/meln5674/minimux"

	"github.com/gorilla/websocket"

	"github.com/meln5674/vnc-ws-gateway/pkg/gateway/static"
)

type Config struct {
	PasswordFile string
	VNCArgs      []string
}

type Server struct {
	Config
	mux minimux.Mux
}

func New(cfg Config) *Server {
	srv := Server{Config: cfg}

	srv.mux = minimux.Mux{
		DefaultHandler: minimux.NotFound,
		PreProcess:     minimux.LogPendingRequest(os.Stderr),
		PostProcess:    minimux.LogCompletedRequestWithPanicTraces(os.Stderr),
		Routes: []minimux.Route{
			minimux.LiteralPath("/").IsHandledByFunc(srv.homepage),
			minimux.LiteralPath("/api/v1/vnc").IsHandledByFunc(srv.vnc),
			minimux.
				PathWithVars("/static/(.+)", "path").
				WithMethods(http.MethodGet).
				IsHandledBy(static.Handler),
		},
	}

	return &srv
}

func (s *Server) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	s.mux.ServeHTTP(resp, req)
}

func (s *Server) homepage(ctx context.Context, w http.ResponseWriter, req *http.Request, _ map[string]string, _ error) error {
	http.Redirect(w, req, "/static/html/index.html", http.StatusPermanentRedirect)
	return nil
}

func (s *Server) vnc(ctx context.Context, w http.ResponseWriter, req *http.Request, _ map[string]string, _ error) error {
	tmpDir, err := os.MkdirTemp("", "vnc-ws-gateway-*")
	if err != nil {
		http.Error(w, "could not create temp directory", http.StatusInternalServerError)
		return fmt.Errorf("could not create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	socketPath := filepath.Join(tmpDir, "vnc.sock")
	args := append([]string{"-fg", "-rfbunixpath", socketPath, "-rfbauth", s.PasswordFile}, s.VNCArgs...)
	vncServer := exec.Command("tigervncserver", args...)
	vncServer.Stdout = os.Stdout
	vncServer.Stderr = os.Stderr
	err = vncServer.Start()
	if err != nil {
		http.Error(w, "failed to start VNC", http.StatusInternalServerError)
		return fmt.Errorf("failed to start VNC: %w", err)
	}
	defer vncServer.Process.Signal(os.Interrupt)

	// TODO: Configurable
	retries := 3
	retryPeriod := 1 * time.Second
	var upstream net.Conn
	for retry := range retries {
		upstream, err = net.Dial("unix", socketPath)
		if err == nil || retry == retries-1 {
			break
		}
		slog.Info("vnc server not ready yet", "error", err)
		time.Sleep(retryPeriod)
	}
	if err != nil {
		http.Error(w, "timed out waiting for VNC server to start", http.StatusInternalServerError)
		return fmt.Errorf("timed out waiting for VNC server to start: %w", err)
	}
	slog.Info("Connected to VNC", "socket-path", socketPath)

	var upgrader websocket.Upgrader
	conn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		return fmt.Errorf("failed to upgrade to websocket: %w", err)
	}
	defer func() {
		slog.Info("stopping ws conn")
		conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		conn.Close()
	}()
	go sendWebsocketPings(ctx, conn, 5*time.Second)

	var closeUpstream func() error
	upstreamCloser, ok := upstream.(HalfCloser)
	if !ok {
		closeUpstream = upstream.Close
	} else {
		closeUpstream = upstreamCloser.CloseWrite
	}

	downstream := &WebsocketConn{Conn: conn}
	var done sync.WaitGroup
	done.Add(1)
	go func() {
		defer done.Done()
		defer downstream.Close()
		var err error
		var n int64
		defer func() { slog.Info("done copying to downstream", "n", n, "error", err) }()
		n, err = io.Copy(downstream, upstream)
	}()
	done.Add(1)
	go func() {
		defer done.Done()
		defer closeUpstream()
		var err error
		var n int64
		defer func() { slog.Info("done copying to upstream", "n", n, "error", err) }()
		n, err = io.Copy(upstream, downstream)
	}()
	done.Wait()

	return nil
}

func sendWebsocketPings(ctx context.Context, conn *websocket.Conn, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			conn.WriteControl(websocket.PingMessage, []byte{}, time.Time{})
		case <-ctx.Done():
			return
		}
	}
}

// WebSocketConn wraps a websocket.Conn to implement the missing methods for net.Conn
type WebsocketConn struct {
	*websocket.Conn
	leftovers []byte
}

func (w *WebsocketConn) Read(b []byte) (n int, err error) {
	if len(w.leftovers) == 0 {
		var typ int
		typ, w.leftovers, err = w.ReadMessage()
		var closeErr *websocket.CloseError
		if errors.As(err, &closeErr) && closeErr.Code == websocket.CloseNormalClosure {
			return 0, io.EOF
		}
		if err != nil {
			return
		}
		switch typ {
		case websocket.CloseMessage:
			return 0, io.EOF
		case websocket.BinaryMessage, websocket.TextMessage:
			return
		default:
			err = fmt.Errorf("Unrecognized message type %d", typ)
			return
		}
	}
	n = copy(b, w.leftovers)
	w.leftovers = w.leftovers[n:]
	return
}

func (w *WebsocketConn) Write(b []byte) (n int, err error) {
	err = w.WriteMessage(websocket.BinaryMessage, b)
	if err == nil {
		n = len(b)
	}
	return
}

func (w *WebsocketConn) Close() error {
	err1 := w.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	err2 := w.Conn.Close()
	return errors.Join(err1, err2)
}

func (w *WebsocketConn) SetDeadline(t time.Time) error {
	errR := w.SetReadDeadline(t)
	errW := w.SetWriteDeadline(t)
	return errors.Join(errR, errW)
}

// A HalfCloser implements the half-closing semantics of e.g. TCP and Unix Sockets
type HalfCloser interface {
	CloseRead() error
	CloseWrite() error
}

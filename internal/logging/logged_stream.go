package logging

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"go.lsp.dev/jsonrpc2"
)

// Majority of this logging has been lifted from: github.com/golang/tools/internal/lsp/protocol/log.go

type LoggedStream struct {
	Label string
	jsonrpc2.Stream
	Logger zerolog.Logger
}

var _ jsonrpc2.Stream = (*LoggedStream)(nil)

type req struct {
	method string
	start  time.Time
}

type mapped struct {
	mu          sync.Mutex
	clientCalls map[string]req
	serverCalls map[string]req
}

var maps = &mapped{
	sync.Mutex{},
	make(map[string]req),
	make(map[string]req),
}

func (s LoggedStream) logMsg(msg jsonrpc2.Message, isRead bool) {
	direction, pastTense := "Received", "Received"
	get, set := maps.client, maps.setServer
	if isRead {
		direction, pastTense = "Sending", "Sent"
		get, set = maps.server, maps.setClient
	}
	if msg == nil {
		return
	}
	tm := time.Now()

	var logMsg string
	args := map[string]any{}

	switch msg := msg.(type) {
	case *jsonrpc2.Call:
		id := fmt.Sprint(msg.ID())
		logMsg = fmt.Sprintf("[%s] %s request '%s - (%s)'.", s.Label, direction, msg.Method(), id)
		args["params"] = string(msg.Params())
		set(id, req{method: msg.Method(), start: tm})
	case *jsonrpc2.Notification:
		logMsg = fmt.Sprintf("[%s] %s notification '%s'.", s.Label, direction, msg.Method())
		args["params"] = string(msg.Params())
	case *jsonrpc2.Response:
		id := fmt.Sprint(msg.ID())
		if err := msg.Err(); err != nil {
			s.Logger.Error().Err(err).Msgf("[%s] %s #%s", s.Label, pastTense, id)
			return
		}
		cc := get(id)
		elapsed := tm.Sub(cc.start)
		logMsg = fmt.Sprintf("[%s] %s response '%s - (%s)' in %dms.", s.Label, direction, cc.method, id, elapsed/time.Millisecond)
		args["result"] = string(msg.Result())
	}
	s.Logger.Info().Fields(args).Msg(logMsg)
}

func (s LoggedStream) Read(ctx context.Context) (jsonrpc2.Message, int64, error) {
	msg, count, err := s.Stream.Read(ctx)
	s.logMsg(msg, true)
	return msg, count, err
}

func (s LoggedStream) Write(ctx context.Context, msg jsonrpc2.Message) (int64, error) {
	count, err := s.Stream.Write(ctx, msg)
	s.logMsg(msg, false)
	return count, err
}

// these 4 methods are each used exactly once, but it seemed
// better to have the encapsulation rather than ad hoc mutex
// code in 4 places
func (m *mapped) client(id string) req {
	m.mu.Lock()
	defer m.mu.Unlock()
	v := m.clientCalls[id]
	delete(m.clientCalls, id)
	return v
}

func (m *mapped) server(id string) req {
	m.mu.Lock()
	defer m.mu.Unlock()
	v := m.serverCalls[id]
	delete(m.serverCalls, id)
	return v
}

func (m *mapped) setClient(id string, r req) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.clientCalls[id] = r
}

func (m *mapped) setServer(id string, r req) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.serverCalls[id] = r
}

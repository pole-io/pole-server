/**
 * Tencent is pleased to support the open source community by making Polaris available.
 *
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *
 * Licensed under the BSD 3-Clause License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * https://opensource.org/licenses/BSD-3-Clause
 *
 * Unless required by applicable law or agreed to in writing, software distributed
 * under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR
 * CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package remote

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/stats"

	"github.com/pole-io/pole-server/pkg/common/eventhub"
	"github.com/pole-io/pole-server/pkg/common/syncs/container"
	"github.com/pole-io/pole-server/pkg/common/utils"
	commontime "github.com/pole-io/pole-server/pkg/common/utils/time"
	nacospb "github.com/pole-io/pole-server/plugin/apiserver/nacosserver/v2/pb"
)

// Define a custom type for context keys to avoid collisions
type contextKey string

const (
	ConnIDKey         contextKey = "nacos_client_conn_id"
	ClientIPKey       contextKey = "nacos_client_ip"
	ConnectionInfoKey contextKey = "nacos_client_conn_meta"
)

const (
	ClientConnectionEvent = "ClientConnectionEvent"

	_ EventType = iota
	EventClientConnected
	EventClientDisConnected
)

type (
	EventType int32

	// Client
	Client struct {
		ID             string             `json:"id"`
		Addr           *net.TCPAddr       `json:"addr"`
		ConnMeta       ConnectionMeta     `json:"conn_meta"`
		refreshTimeRef atomic.Value
		streamRef      atomic.Value       `json:"-"`
		cancel         context.CancelFunc `json:"-"`
	}

	// ConnectionEvent
	ConnectionEvent struct {
		EventType EventType
		ConnID    string
		Client    *Client
	}

	// ConnectionMeta
	ConnectionMeta struct {
		ConnectType      string
		ClientIp         string
		RemoteIp         string
		RemotePort       int
		LocalPort        int
		Version          string
		ConnectionId     string
		CreateTime       time.Time
		AppName          string
		Tenant           string
		Labels           map[string]string
		ClientAttributes nacospb.ClientAbilities
	}

	// SyncServerStream
	SyncServerStream struct {
		lock   sync.Mutex
		Stream grpc.ServerStream
	}
)

func (c *Client) SetStreamRef(stream *SyncServerStream) {
	c.streamRef.Store(stream)
}

func (c *Client) LoadStream() (*SyncServerStream, bool) {
	return c.loadStream()
}

func (c *Client) loadStream() (*SyncServerStream, bool) {
	val := c.streamRef.Load()
	if val == nil {
		return nil, false
	}
	stream, ok := val.(*SyncServerStream)
	return stream, ok
}

func (c *Client) loadRefreshTime() time.Time {
	return c.refreshTimeRef.Load().(time.Time)
}

func (c *Client) Close() {
	c.cancel()
}

// Context returns the context for this stream.
func (s *SyncServerStream) Context() context.Context {
	return s.Stream.Context()
}

func (s *SyncServerStream) SendMsg(m interface{}) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.Stream.SendMsg(m)
}

type ConnectionManager struct {
	inFlights   *InFlights
	tcpConns    *container.SyncMap[string, net.Conn] // Replaces sync.Map
	clients     *container.SyncMap[string, *Client]  // Replaces sync.Map
	connections *container.SyncMap[string, *Client]  // Replaces sync.Map
	cancel      context.CancelFunc
}

func NewConnectionManager() *ConnectionManager {
	ctx, cancel := context.WithCancel(context.Background())
	mgr := &ConnectionManager{
		connections: container.NewSyncMap[string, *Client](),
		clients:     container.NewSyncMap[string, *Client](),
		tcpConns:    container.NewSyncMap[string, net.Conn](),
		inFlights:   NewInFlights(ctx),
		cancel:      cancel,
	}
	go mgr.doEject(ctx)
	return mgr
}

func (h *ConnectionManager) InFlights() *InFlights {
	return h.inFlights
}

// OnAccept call when net.Conn accept
func (h *ConnectionManager) OnAccept(conn net.Conn) {
	addr := conn.RemoteAddr().(*net.TCPAddr)
	h.tcpConns.Store(addr.String(), conn)
}

// OnRelease call when net.Conn release
func (h *ConnectionManager) OnRelease(conn net.Conn) {
	addr := conn.RemoteAddr().(*net.TCPAddr)
	h.tcpConns.Delete(addr.String())
}

// OnClose call when net.Listener close
func (h *ConnectionManager) OnClose() {
	// do nothing
}

func (h *ConnectionManager) RegisterConnection(ctx context.Context, payload *nacospb.Payload,
	req *nacospb.ConnectionSetupRequest) error {

	connID := ValueConnID(ctx)

	connMeta := ConnectionMeta{
		ClientIp:         payload.GetMetadata().GetClientIp(),
		Version:          "",
		ConnectionId:     connID,
		CreateTime:       time.Now(),
		AppName:          "-",
		Tenant:           req.Tenant,
		Labels:           req.Labels,
		ClientAttributes: req.ClientAbilities,
	}
	if val, ok := req.Labels["AppName"]; ok {
		connMeta.AppName = val
	}

	client, ok := h.clients.Load(connID)
	if !ok {
		return errors.New("connection register fail, not found target client")
	}

	client.ConnMeta = connMeta

	return nil
}

func (h *ConnectionManager) UnRegisterConnection(connID string) {
	client, ok := h.clients.Load(connID)
	if ok {
		_ = eventhub.Publish(ClientConnectionEvent, &ConnectionEvent{
			EventType: EventClientDisConnected,
			ConnID:    connID,
			Client:    client,
		})

		h.clients.Delete(connID)
		h.connections.Delete(client.Addr.String())

		if tcpConn, ok := h.tcpConns.Load(client.Addr.String()); ok {
			_ = tcpConn.Close()
		}
	}
}

func (h *ConnectionManager) GetClient(id string) (*Client, bool) {
	client, ok := h.clients.Load(id)
	return client, ok
}

func (h *ConnectionManager) GetClientByAddr(addr string) (*Client, bool) {
	client, ok := h.connections.Load(addr)
	return client, ok
}

// TagRPC can attach some information to the given context.
// The context used for the rest lifetime of the RPC will be derived from
// the returned context.
func (h *ConnectionManager) TagRPC(ctx context.Context, _ *stats.RPCTagInfo) context.Context {
	// do nothing
	return ctx
}

// HandleRPC processes the RPC stats.
func (h *ConnectionManager) HandleRPC(ctx context.Context, s stats.RPCStats) {
	// do nothing
}

// TagConn can attach some information to the given context.
// The returned context will be used for stats handling.
// For conn stats handling, the context used in HandleConn for this
// connection will be derived from the context returned.
// For RPC stats handling,
//   - On server side, the context used in HandleRPC for all RPCs on this
//
// connection will be derived from the context returned.
//   - On client side, the context is not derived from the context returned.
func (h *ConnectionManager) TagConn(ctx context.Context, connInfo *stats.ConnTagInfo) context.Context {
	clientAddr := connInfo.RemoteAddr.(*net.TCPAddr)

	h.connections.ComputeIfAbsent(clientAddr.String(), func(_ string) *Client {
		connId := fmt.Sprintf("%d_%s_%d_%s", commontime.CurrentMillisecond(), clientAddr.IP, clientAddr.Port,
			utils.LocalHost)
		client := &Client{
			ID:             connId,
			Addr:           clientAddr,
			refreshTimeRef: atomic.Value{},
			streamRef:      atomic.Value{},
		}
		client.refreshTimeRef.Store(time.Now())
		h.clients.Store(connId, client)
		return client
	})

	client, _ := h.connections.Load(clientAddr.String())
	return context.WithValue(ctx, ConnIDKey, client.ID)
}

// HandleConn processes the Conn stats.
func (h *ConnectionManager) HandleConn(ctx context.Context, s stats.ConnStats) {
	switch s.(type) {
	case *stats.ConnBegin:
		connID, _ := ctx.Value(ConnIDKey).(string)
		nacoslog.Info("[NACOS-V2][ConnMgr] grpc conn begin", zap.String("conn-id", connID))
		client, _ := h.clients.Load(connID)
		_ = eventhub.Publish(ClientConnectionEvent, &ConnectionEvent{
			EventType: EventClientConnected,
			ConnID:    connID,
			Client:    client,
		})
	case *stats.ConnEnd:
		connID, _ := ctx.Value(ConnIDKey).(string)
		nacoslog.Info("[NACOS-V2][ConnMgr] grpc conn end", zap.String("conn-id", connID))
		h.UnRegisterConnection(connID)
	}
}

func (h *ConnectionManager) RefreshClient(ctx context.Context) {
	connID := ValueConnID(ctx)
	if client, ok := h.clients.Load(connID); ok {
		client.refreshTimeRef.Store(time.Now())
	}
}

func (h *ConnectionManager) GetStream(connID string) *SyncServerStream {
	if client, ok := h.clients.Load(connID); ok {
		stream, _ := client.loadStream()
		return stream
	}
	return nil
}

func (h *ConnectionManager) ListConnections() map[string]*Client {
	return h.listConnections()
}

func (h *ConnectionManager) listConnections() map[string]*Client {
	ret := make(map[string]*Client)
	h.clients.ReadRange(func(connID string, client *Client) {
		ret[connID] = client
	})
	return ret
}

func (h *ConnectionManager) doEject(ctx context.Context) {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			h.ejectOutdateConnection()
			h.ejectOverLimitConnection()
		case <-ctx.Done():
			return
		}
	}
}

func (h *ConnectionManager) ejectOverLimitConnection() {
	// TODO: it need impl ?
}

func (h *ConnectionManager) ejectOutdateConnection() {

	keepAliveTime := 4 * 5 * time.Second
	now := time.Now()
	outDatedConnections := map[string]*Client{}
	connIds := make([]string, 0, 4)

	h.clients.Range(func(connID string, conn *Client) {
		if now.Sub(conn.loadRefreshTime()) >= keepAliveTime {
			outDatedConnections[connID] = conn
			connIds = append(connIds, connID)
		}
	})

	if len(outDatedConnections) != 0 {
		nacoslog.Info("[NACOS-V2][ConnectionManager] out dated connection",
			zap.Int("size", len(outDatedConnections)), zap.Strings("conn-ids", connIds))
	}

	successConnections := container.NewSyncMap[string, struct{}]()
	wait := &sync.WaitGroup{}
	wait.Add(len(outDatedConnections))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	for connID := range outDatedConnections {
		req := nacospb.NewClientDetectionRequest()
		req.RequestId = utils.NewUUID()

		outDateConnectionId := connID
		outDateConnection := outDatedConnections[outDateConnectionId]
		// add inflight first
		_ = h.inFlights.AddInFlight(&InFlight{
			ConnID:     connID,
			RequestID:  req.RequestId,
			ExpireTime: time.Now().Add(5 * time.Second),
			Callback: func(attachment map[string]interface{}, resp nacospb.BaseResponse, err error) {
				defer wait.Done()
				select {
				case <-ctx.Done():
					// 已经结束不作处理
					return
				default:
					if resp != nil && resp.IsSuccess() {
						outDateConnection.refreshTimeRef.Store(time.Now())
						successConnections.Store(outDateConnectionId, struct{}{})
					}
				}
			},
		})
	}
	go func() {
		defer cancel()
		wait.Wait()
	}()
	<-ctx.Done()
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		// TODO log
	}
	for connID := range outDatedConnections {
		if _, ok := successConnections.Load(connID); !ok {
			h.UnRegisterConnection(connID)
		}
	}
}

func ValueConnID(ctx context.Context) string {
	ret, _ := ctx.Value(ConnIDKey).(string)
	return ret
}

func ValueClientIP(ctx context.Context) string {
	ret, _ := ctx.Value(ClientIPKey).(string)
	return ret
}

func ValueConnMeta(ctx context.Context) ConnectionMeta {
	ret, _ := ctx.Value(ConnectionInfoKey).(ConnectionMeta)
	return ret
}

// Copyright 2016 IBM Corporation
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package replication

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/amalgam8/amalgam8/pkg/auth"
	"github.com/amalgam8/amalgam8/registry/cluster"
	"github.com/amalgam8/amalgam8/registry/utils/channels"
	"github.com/amalgam8/amalgam8/registry/utils/health"
	"github.com/amalgam8/amalgam8/registry/utils/logging"
)

const (
	module         string = "REPLICATION"
	version        string = "v1"
	repContext     string = "replication"
	syncContext    string = "sync"
	headerMemberID string = "Member-ID"
	repTimeout            = time.Duration(7) * time.Second
)

type server struct {
	listener   net.Listener
	httpclient *http.Client
	transport  *http.Transport

	// Replication outgoing messages
	broadcast channels.ChannelTimeout
	repair    channels.ChannelTimeout

	// New remote peers connections
	newPeers chan *peer

	// Closed peers connection connections
	closingPeers chan *peer

	// Peers connections registry
	peers map[cluster.MemberID]*peer

	// Holds the server http multiplexer , required to be able to add routing patterns and handlers
	mux *http.ServeMux

	// replicators
	replicators     map[auth.Namespace]*replicator
	replicatorsLock sync.RWMutex

	// Local client per Remote peer
	clients     map[cluster.MemberID]*client
	clientsLock sync.Mutex

	// Cluster info
	selfID      cluster.MemberID
	membership  cluster.Membership
	registrator cluster.Registrator

	// Receive channel for incoming messages used to notify external listeners (e.g. Registry)
	notifyChannel chan *InMessage

	// Sync channel for incoming sync requests
	syncReqChannel chan chan []byte

	// Health checker
	health *healthChecker

	done chan struct{}

	logger *log.Entry
}

type peer struct {
	memberID   cluster.MemberID
	msgChannel chan *sse
}

type gzipResponseWrapper struct {
	io.Writer
	http.ResponseWriter
}

func (w gzipResponseWrapper) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func (w gzipResponseWrapper) Header() http.Header {
	return w.ResponseWriter.Header()
}

func (w gzipResponseWrapper) Flush() {
	w.Writer.(*gzip.Writer).Flush()
	w.ResponseWriter.(http.Flusher).Flush()
}

// New - Create a new replication instance
func New(conf *Config) (Replication, error) {
	var lentry = logging.GetLogger(module)

	if conf == nil {
		err := fmt.Errorf("Nil conf")
		lentry.WithFields(log.Fields{
			"error": err,
		}).Error("Failed to create replication server")

		return nil, err
	}

	if conf.Membership == nil || conf.Registrator == nil {
		err := fmt.Errorf("Nil cluster membership and/or registrator")
		lentry.WithFields(log.Fields{
			"error": err,
		}).Error("Failed to create replication server")

		return nil, err
	}

	// Make sure that the listening port is free
	address := fmt.Sprintf("%s:%d", conf.Registrator.Self().IP(), conf.Registrator.Self().Port())
	listener, err := net.Listen("tcp", address)
	if err != nil {
		lentry.WithFields(log.Fields{
			"error": err,
		}).Error("Failed to create replication server")

		return nil, err
	}

	tr := &http.Transport{MaxIdleConnsPerHost: 1}
	hc := &http.Client{Transport: tr}
	logger := lentry.WithFields(log.Fields{"Member-ID": conf.Registrator.Self().ID()})

	// Instantiate a server
	s := &server{
		listener:       listener,
		httpclient:     hc,
		transport:      tr,
		broadcast:      channels.NewChannelTimeout(512),
		repair:         channels.NewChannelTimeout(512),
		newPeers:       make(chan *peer),
		closingPeers:   make(chan *peer, 8),
		peers:          make(map[cluster.MemberID]*peer),
		mux:            http.NewServeMux(),
		replicators:    make(map[auth.Namespace]*replicator),
		clients:        make(map[cluster.MemberID]*client),
		selfID:         conf.Registrator.Self().ID(),
		membership:     conf.Membership,
		registrator:    conf.Registrator,
		notifyChannel:  make(chan *InMessage, 512),
		syncReqChannel: make(chan chan []byte, 8),
		health:         newHealthChecker(),
		done:           make(chan struct{}),
		logger:         logger,
	}

	health.Register(module, s.health)

	logger.Info("Replication server created")
	return s, nil
}

func (s *server) GetReplicator(namespace auth.Namespace) (Replicator, error) {
	rep := newReplicator(namespace, s.broadcast, s.repair)
	s.logger.WithFields(log.Fields{
		"namespace": rep.Namespace,
	}).Info("Add a replicator")

	s.replicatorsLock.Lock()
	defer s.replicatorsLock.Unlock()

	if existRep, exists := s.replicators[rep.Namespace]; exists {
		s.logger.WithFields(log.Fields{
			"namespace": rep.Namespace,
			"error":     "replicator already exists",
		}).Error("Failed to add a replicator")

		return existRep, fmt.Errorf("Replicator %s already exists", rep.Namespace)
	}

	s.replicators[rep.Namespace] = rep
	return rep, nil
}

func (s *server) Notification() <-chan *InMessage {
	return s.notifyChannel
}

func (s *server) Sync(waitTime time.Duration) <-chan *InMessage {
	s.logger.Info("Starting synchronization")
	syncChan := make(chan *InMessage)
	go s.doSync(waitTime, syncChan)
	return syncChan
}

func (s *server) SyncRequest() <-chan chan []byte {
	return s.syncReqChannel
}

func (s *server) Stop() {
	s.logger.Info("Stopping replication server")
	if err := s.listener.Close(); err != nil {
		s.logger.WithFields(log.Fields{
			"error": err,
		}).Warn("Failed to close listener")
	}
	close(s.done)
	s.transport.CloseIdleConnections()
}

func (s *server) doSync(waitTime time.Duration, syncChan chan *InMessage) {
	timeout := time.Now().Add(waitTime)
	// Search for a member to sync with
	for time.Now().Before(timeout) {
		for m := range s.membership.Members() {
			if s.selfID != m.ID() {
				if err := newSyncClient(s.selfID, m, s.httpclient, syncChan, s.logger); err == nil {
					goto syncok
				}
			}
		}
		time.Sleep(time.Duration(200) * time.Millisecond)
	}
	s.logger.Info("No active member found. Synchronization was skipped")

syncok:
	close(syncChan)
	s.startService()
	s.logger.Info("Ending synchronization")
}

func (s *server) startService() {
	// Start server active end point
	address := fmt.Sprintf("%s:%d", s.registrator.Self().IP(), s.registrator.Self().Port())
	s.mux.HandleFunc("/"+version+"/"+repContext, s.serveRep)
	s.mux.HandleFunc("/"+version+"/"+syncContext, s.serveSync)

	s.logger.Infof("Starting replication service on %s", address)

	go func() {
		if err := http.Serve(s.listener, s.mux); err != nil {
			s.logger.WithFields(log.Fields{
				"error": err,
			}).Info("HTTP listener has stopped")
		}
	}()

	// Set it running - listening and broadcasting events
	go s.listen()

	if err := s.registrator.Join(); err != nil {
		s.logger.WithFields(log.Fields{
			"error": err,
		}).Error("Failed to join")
		return
	}

	s.membership.RegisterListener(s)
	// Launch client for existing members
	for key := range s.membership.Members() {
		if s.selfID != key.ID() {
			s.logger.Infof("Add an existing member %s", key)
			s.OnJoin(key)
		}
	}
}

func (s *server) serveRep(rw http.ResponseWriter, req *http.Request) {
	memberID := s.validateConnection(rw, req)
	if memberID == "" {
		return
	}
	s.logger.Infof("Peer Member %s connected from address %s", memberID, req.RemoteAddr)

	peer := &peer{memberID: cluster.MemberID(memberID), msgChannel: make(chan *sse)}

	// Signal that we have a new connection
	s.newPeers <- peer

	// Remove this client from the map of connected clients
	// when this handler exits.
	defer func() {
		s.closingPeers <- peer
	}()

	// Listen to connection close and un-register messageChan
	notify := rw.(http.CloseNotifier).CloseNotify()
	go func() {
		<-notify
		s.closingPeers <- peer
		s.logger.Infof("Peer member %s has been disconnected", memberID)
	}()

	// Send the response header to client side
	flusher := rw.(http.Flusher)

	gzipped := strings.Contains(req.Header.Get("Accept-Encoding"), "gzip")

	var gzipWriter *gzip.Writer
	var err error

	if gzipped {
		// override the response writer to be a gzipped writer
		if gzipWriter, err = gzip.NewWriterLevel(rw, gzip.BestSpeed); err == nil {
			rw = gzipResponseWrapper{Writer: gzipWriter, ResponseWriter: rw}
			rw.Header().Set("Content-Encoding", "gzip")
			flusher = rw.(http.Flusher)
			defer func() {
				gzipWriter.Close()
			}()
		} else {
			s.logger.Warnf("Gzip wrapper creation for %s from %s failed: %v. Falling back to uncompressed HTTP", memberID, req.RemoteAddr, err)
		}
	}

	encoder := newEncoder(rw)
	ticker := time.NewTicker(time.Millisecond * 100).C
	for {
		select {
		case ev, ok := <-peer.msgChannel:
			if !ok {
				return
			}
			// Write to the ResponseWriter
			// Server Sent Events compatible
			if err := encoder.Encode(ev); err != nil {
				s.logger.WithFields(log.Fields{
					"error": err,
				}).Errorf("Failed to encode replication message to member %s", memberID)
				break
			}
		case <-ticker:
			flusher.Flush()
		}
	}
}

func (s *server) serveSync(rw http.ResponseWriter, req *http.Request) {
	memberID := s.validateConnection(rw, req)
	if memberID == "" {
		return
	}
	s.logger.Infof("Peer Member %s started synchronization from address %s", memberID, req.RemoteAddr)

	// Create a new channel for the new member synchronization
	syncChan := make(chan []byte)
	s.syncReqChannel <- syncChan

	// Send the response header to client side
	flusher := rw.(http.Flusher)

	gzipped := strings.Contains(req.Header.Get("Accept-Encoding"), "gzip")

	var gzipWriter *gzip.Writer
	var err error

	if gzipped {
		// override the response writer to be a gzipped writer
		if gzipWriter, err = gzip.NewWriterLevel(rw, gzip.DefaultCompression); err == nil {
			rw = gzipResponseWrapper{Writer: gzipWriter, ResponseWriter: rw}
			rw.Header().Set("Content-Encoding", "gzip")
			flusher = rw.(http.Flusher)
			defer func() {
				gzipWriter.Close()
			}()
		} else {
			s.logger.Warnf("Gzip wrapper creation for %s from %s failed: %v. Falling back to uncompressed HTTP", memberID, req.RemoteAddr, err)
		}
	}

	encoder := newEncoder(rw)
	var msgID uint64
	for data := range syncChan {
		ev := &sse{id: strconv.FormatUint(msgID, 10), event: "SYNC", data: string(data)}
		// Write to the ResponseWriter
		// Server Sent Events compatible
		if err := encoder.Encode(ev); err != nil {
			s.logger.WithFields(log.Fields{
				"error": err,
			}).Errorf("Failed to encode sync message to peer member %s", memberID)
			return
		}
		msgID++
	}
	flusher.Flush()
	s.logger.Infof("Peer Member %s synchronization has been completed successfully", memberID)
}

func (s *server) validateConnection(rw http.ResponseWriter, req *http.Request) cluster.MemberID {
	// Make sure that the writer supports flushing.
	_, ok := rw.(http.Flusher)

	if !ok {
		s.logger.Errorf("Streaming unsupported. Member %s", req.RemoteAddr)
		http.Error(rw, "Streaming unsupported!", http.StatusInternalServerError)
		return ""
	}

	rw.Header().Set("Content-Type", "text/event-stream")
	rw.Header().Set("Cache-Control", "no-cache")
	rw.Header().Set("Connection", "keep-alive")

	memberID := req.Header.Get(headerMemberID)
	if memberID == "" {
		s.logger.Errorf("Missing header Memeber-ID from connection address %s", req.RemoteAddr)
		http.Error(rw, "Missing header Member-ID", http.StatusBadRequest)
		return ""
	}

	if cluster.MemberID(memberID) == s.registrator.Self().ID() {
		s.logger.Errorf("Header Member-ID %s conflicts with self on connection %s", memberID, req.RemoteAddr)
		http.Error(rw, "Wrong Member-ID", http.StatusBadRequest)
		return ""
	}
	return cluster.MemberID(memberID)
}

func (s *server) listen() {
	var msgID uint64 = 1

	for {
		select {
		case peer := <-s.newPeers:
			// A new client has connected.
			// Register their message channel
			if p, exist := s.peers[peer.memberID]; exist {
				s.logger.Infof("Duplicated connected peer %s found. Closing the old one", peer.memberID)
				close(p.msgChannel)
			}
			s.peers[peer.memberID] = peer
		case peer := <-s.closingPeers:
			// A client has detached and we want to
			// stop sending them messages.
			if p, exist := s.peers[peer.memberID]; exist && p.msgChannel == peer.msgChannel {
				close(peer.msgChannel)
				delete(s.peers, peer.memberID)
			}
		case msg := <-s.broadcast.Channel():
			// Send Server Event (SSE) formatted event message
			data, _ := json.Marshal(msg)
			ev := &sse{id: strconv.FormatUint(msgID, 10), event: "REP", data: string(data)}
			msgID++
			// We got a new event from the outside!
			// Send event to all connected clients
			for _, peer := range s.peers {
				peer.msgChannel <- ev
			}
		case msg := <-s.repair.Channel():
			// Send Server Event (SSE) formatted event message
			outMsg := msg.(*outMessage)
			data, _ := json.Marshal(outMsg)
			ev := &sse{id: "0", event: "REPAIR", data: string(data)}
			// We got a new event from the outside!
			// Send event to all connected clients
			if peer, exists := s.peers[outMsg.memberID]; !exists {
				s.logger.Warnf("Failed send event to member %s, data %s", outMsg.memberID, outMsg.Data)
			} else {
				peer.msgChannel <- ev
			}
		case <-s.done:
			if err := s.registrator.Leave(); err != nil {
				s.logger.WithFields(log.Fields{
					"error": err,
				}).Warn("Failed to leave the cluster")
			}
			close(s.newPeers)
			s.logger.Info("Replication server has stopped")
			return
		}
	}
}

// Invoked when a member joins a cluster.
func (s *server) OnJoin(m cluster.Member) {
	if s.selfID == m.ID() {
		return
	}

	s.clientsLock.Lock()
	defer s.clientsLock.Unlock()

	s.logger.Infof("Peer Member %s joined the cluster", m)

	if client, exists := s.clients[m.ID()]; exists {
		client.close()
		delete(s.clients, m.ID())
		s.health.RemoveClient(m.ID())
	}

	client, err := newClient(s.selfID, m, s.httpclient, s.notifyChannel, s.logger)
	if err != nil {
		s.logger.WithFields(log.Fields{
			"error": err,
		}).Errorf("Failed to add the member %s", m)

		return
	}
	s.clients[m.ID()] = client
	s.health.AddClient(m.ID(), client)
}

// Invoked when a member leaves a cluster.
func (s *server) OnLeave(m cluster.Member) {
	if s.selfID == m.ID() {
		return
	}

	s.clientsLock.Lock()
	defer s.clientsLock.Unlock()

	s.logger.Infof("Member %s left the cluster", m)

	if client, exists := s.clients[m.ID()]; exists {
		client.close()
		delete(s.clients, m.ID())
		s.health.RemoveClient(m.ID())
	}
}

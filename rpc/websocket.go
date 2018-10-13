// Copyright 2015 The aquachain Authors
// This file is part of the aquachain library.
//
// The aquachain library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The aquachain library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the aquachain library. If not, see <http://www.gnu.org/licenses/>.

package rpc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"golang.org/x/net/websocket"

	set "github.com/deckarep/golang-set"
	"gitlab.com/aquachain/aquachain/common/log"
)

// websocketJSONCodec is a custom JSON codec with payload size enforcement and
// special number parsing.
var websocketJSONCodec = websocket.Codec{
	// Marshal is the stock JSON marshaller used by the websocket library too.
	Marshal: func(v interface{}) ([]byte, byte, error) {
		msg, err := json.Marshal(v)
		return msg, websocket.TextFrame, err
	},
	// Unmarshal is a specialized unmarshaller to properly convert numbers.
	Unmarshal: func(msg []byte, payloadType byte, v interface{}) error {
		dec := json.NewDecoder(bytes.NewReader(msg))
		dec.UseNumber()

		return dec.Decode(v)
	},
}

// WebsocketHandler returns a handler that serves JSON-RPC to WebSocket connections.
//
// allowedOrigins should be a comma-separated list of allowed origin URLs.
// To allow connections with any origin, pass "*".
func (srv *Server) WebsocketHandler(allowedOrigins []string, allowedIP []string) http.Handler {
	return websocket.Server{
		Handshake: wsHandshakeValidator(allowedOrigins, allowedIP),
		Handler: func(conn *websocket.Conn) {

			// Create a custom encode/decode pair to enforce payload size and number encoding
			conn.MaxPayloadBytes = maxHTTPRequestContentLength

			encoder := func(v interface{}) error {
				return websocketJSONCodec.Send(conn, v)
			}
			decoder := func(v interface{}) error {
				return websocketJSONCodec.Receive(conn, v)
			}
			srv.ServeCodec(NewCodec(conn, encoder, decoder), OptionMethodInvocation|OptionSubscriptions)
		},
	}
}

// NewWSServer creates a new websocket RPC server around an API provider.
//
// Deprecated: use Server.WebsocketHandler
func NewWSServer(allowedOrigins []string, allowedIP []string, srv *Server) *http.Server {
	return &http.Server{Handler: srv.WebsocketHandler(allowedOrigins, allowedIP)}
}

// wsHandshakeValidator returns a handler that verifies the origin during the
// websocket upgrade process. When a '*' is specified as an allowed origins all
// connections are accepted.
func wsHandshakeValidator(allowedOrigins, allowedIP []string) func(*websocket.Config, *http.Request) error {
	origins := set.NewSet()
	allowedIPset := set.NewSet()
	allowAllOrigins := false
	allowAllIP := false
	for _, origin := range allowedOrigins {
		if origin == "*" {
			allowAllOrigins = true
		}
		if origin != "" {
			origins.Add(strings.ToLower(origin))
		}
	}
	for _, ip := range allowedIP {
		if ip == "*" {
			allowAllIP = true
			break
		}
		if ip != "" {
			allowedIPset.Add(strings.ToLower(ip))
		}
	}

	// allow localhost if no allowedOrigins are specified.
	if len(origins.ToSlice()) == 0 {
		origins.Add("http://localhost")
		if hostname, err := os.Hostname(); err == nil {
			origins.Add("http://" + strings.ToLower(hostname))
		}
	}

	log.Debug(fmt.Sprintf("Allowed origin(s) for WS RPC interface %v\n", origins.ToSlice()))

	f := func(cfg *websocket.Config, req *http.Request) error {
		if !allowAllIP {
			checkip := func(r *http.Request) error {
				ip := getIP(r)
				if ip == "" {
					ip = strings.Split(r.RemoteAddr, ":")[0]
				}
				if allowedIPset.Contains(ip) {
					return nil
				}
				log.Warn("unwarranted websocket request", "ip", ip)
				return fmt.Errorf("ip not allowed")
			}
			if err := checkip(req); err != nil {
				return err
			}
		}
		origin := strings.ToLower(req.Header.Get("Origin"))
		if allowAllOrigins || origins.Contains(origin) {
			return nil
		}
		log.Warn(fmt.Sprintf("origin '%s' not allowed on WS-RPC interface\n", origin))
		return fmt.Errorf("origin %s not allowed", origin)
	}

	return f
}

var wsPortMap = map[string]string{"ws": "80", "wss": "443"}

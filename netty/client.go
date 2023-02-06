/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package netty

import (
	"fmt"
	gxnet "github.com/AlexStocks/goext/net"
	gxsync "github.com/dubbogo/gost/sync"
	"math/rand"
	"net"
	"sync"
	"time"
)

import (
	"github.com/AlexStocks/getty/transport"
)

import (
	log "github.com/AlexStocks/getty/util"
)

var (
	reqID uint32
	src   = rand.NewSource(time.Now().UnixNano())
)

const (
	WritePkgTimeout = time.Millisecond * 1000
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

////////////////////////////////////////////////////////////////////
// echo client
////////////////////////////////////////////////////////////////////

type EchoClient struct {
	lock        sync.RWMutex
	sessions    []*clientEchoSession
	gettyClient getty.Client
	conf        Config
}

type clientEchoSession struct {
	session getty.Session
	reqNum  int32
}

// StartClient 启用客户端
func StartClient(conf Config) *EchoClient {
	taskPool := gxsync.NewTaskPoolSimple(0)
	clientOpts := []getty.ClientOption{getty.WithServerAddress(gxnet.HostAddress(conf.ServerHost, conf.ServerPort))}
	clientOpts = append(clientOpts, getty.WithClientTaskPool(taskPool))
	if conf.ConnectionNum != 0 {
		clientOpts = append(clientOpts, getty.WithConnectionNumber(conf.ConnectionNum))
	}
	var client = new(EchoClient)
	client.conf = conf
	client.gettyClient = getty.NewTCPClient(clientOpts...)
	messageHandler := NewEchoMessageHandler(client)
	client.gettyClient.RunEventLoop(newSession(conf, messageHandler))
	return client
}

func newSession(conf Config, msgHandler *EchoMessageHandler) func(getty.Session) error {
	return func(session getty.Session) error {
		var (
			ok      bool
			tcpConn *net.TCPConn
		)

		if conf.GettySessionParam.CompressEncoding {
			session.SetCompressType(getty.CompressZip)
		}

		if tcpConn, ok = session.Conn().(*net.TCPConn); !ok {
			panic(fmt.Sprintf("%s, session.conn{%#v} is not tcp connection\n", session.Stat(), session.Conn()))
		}

		tcpConn.SetNoDelay(conf.GettySessionParam.TcpNoDelay)
		tcpConn.SetKeepAlive(conf.GettySessionParam.TcpKeepAlive)
		if conf.GettySessionParam.TcpKeepAlive {
			tcpConn.SetKeepAlivePeriod(conf.GettySessionParam.KeepAlivePeriod2)
		}
		tcpConn.SetReadBuffer(conf.GettySessionParam.TcpRBufSize)
		tcpConn.SetWriteBuffer(conf.GettySessionParam.TcpWBufSize)

		session.SetName(conf.GettySessionParam.SessionName)
		session.SetMaxMsgLen(conf.GettySessionParam.MaxMsgLen)
		session.SetPkgHandler(echoPkgHandler)
		session.SetEventListener(msgHandler)
		session.SetReadTimeout(conf.GettySessionParam.TcpReadTimeout2)
		session.SetWriteTimeout(conf.GettySessionParam.TcpWriteTimeout2)
		session.SetCronPeriod((int)(conf.HeartbeatPeriod2.Nanoseconds() / 1e6))
		session.SetWaitTime(conf.GettySessionParam.WaitTimeout2)
		log.Debug("client new session:%s\n", session.Stat())
		return nil
	}
}

func (c *EchoClient) IsAvailable() bool {
	if c.SelectSession() == nil {
		return false
	}
	return true
}

func (c *EchoClient) Close() {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.gettyClient != nil {
		c.gettyClient.Close()
		c.gettyClient = nil
		for _, s := range c.sessions {
			log.Info("close client session{%s, last active:%s, request number:%d}",
				s.session.Stat(), s.session.GetActive().String(), s.reqNum)
			s.session.Close()
		}
		c.sessions = c.sessions[:0]
	}
}

func (c *EchoClient) SelectSession() getty.Session {
	c.lock.RLock()
	defer c.lock.RUnlock()
	count := len(c.sessions)
	if count == 0 {
		log.Info("client session array is nil...")
		return nil
	}
	return c.sessions[rand.Int31n(int32(count))].session
}

func (c *EchoClient) AddSession(session getty.Session) {
	log.Debug("add session{%s}", session.Stat())
	if session == nil {
		return
	}
	c.lock.Lock()
	c.sessions = append(c.sessions, &clientEchoSession{session: session})
	c.lock.Unlock()
}

func (c *EchoClient) RemoveSession(session getty.Session) {
	if session == nil {
		return
	}
	c.lock.Lock()
	for i, s := range c.sessions {
		if s.session == session {
			c.sessions = append(c.sessions[:i], c.sessions[i+1:]...)
			log.Debug("delete session{%s}, its index{%d}", session.Stat(), i)
			break
		}
	}
	log.Info("after remove session{%s}, left session number:%d", session.Stat(), len(c.sessions))

	c.lock.Unlock()
}

func (c *EchoClient) UpdateSession(session getty.Session) {
	if session == nil {
		return
	}
	c.lock.Lock()
	for i, s := range c.sessions {
		if s.session == session {
			c.sessions[i].reqNum++
			break
		}
	}
	c.lock.Unlock()
}

func (c *EchoClient) GetClientEchoSession(session getty.Session) (clientEchoSession, error) {
	var (
		err         error
		echoSession clientEchoSession
	)
	c.lock.Lock()
	err = errSessionNotExist
	for _, s := range c.sessions {
		if s.session == session {
			echoSession = *s
			err = nil
			break
		}
	}
	c.lock.Unlock()
	return echoSession, err
}

func (c *EchoClient) Heartbeat(session getty.Session) {
	var pkg = EchoPackage{
		B: "ping",
	}
	if _, _, err := session.WritePkg(&pkg, WritePkgTimeout); err != nil {
		log.Warn("session.WritePkg(session{%s}, pkg{%s}) = error{%v}", session.Stat(), pkg, err)
		session.Close()
		c.RemoveSession(session)
	}
}

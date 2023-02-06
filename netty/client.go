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
	WritePkgTimeout = time.Millisecond * 2000
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

////////////////////////////////////////////////////////////////////
// echo client
////////////////////////////////////////////////////////////////////

type EchoClient struct {
	Lock        sync.RWMutex
	Sessions    []*ClientEchoSession
	GettyClient getty.Client
	Conf        Config
}

type ClientEchoSession struct {
	Session getty.Session
	ReqNum  int32
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
	client.Conf = conf
	client.GettyClient = getty.NewTCPClient(clientOpts...)
	// 事件监听器（业务处理）
	messageHandler := NewEchoMessageHandler(client)
	// 解码处理器
	var pkgHandler = NewEchoPackageHandler()
	client.GettyClient.RunEventLoop(newSession(conf, messageHandler, pkgHandler))
	return client
}

func newSession(conf Config, eventHandler getty.EventListener, pkgHandler getty.ReadWriter) func(getty.Session) error {
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
		session.SetPkgHandler(pkgHandler)
		session.SetEventListener(eventHandler)
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
	c.Lock.Lock()
	defer c.Lock.Unlock()
	if c.GettyClient != nil {
		c.GettyClient.Close()
		c.GettyClient = nil
		for _, s := range c.Sessions {
			log.Info("close client session{%s, last active:%s, request number:%d}",
				s.Session.Stat(), s.Session.GetActive().String(), s.ReqNum)
			s.Session.Close()
		}
		c.Sessions = c.Sessions[:0]
	}
}

func (c *EchoClient) SelectSession() getty.Session {
	c.Lock.RLock()
	defer c.Lock.RUnlock()
	count := len(c.Sessions)
	if count == 0 {
		log.Info("client session array is nil...")
		return nil
	}
	return c.Sessions[rand.Int31n(int32(count))].Session
}

func (c *EchoClient) AddSession(session getty.Session) {
	log.Debug("add session{%s}", session.Stat())
	if session == nil {
		return
	}
	c.Lock.Lock()
	c.Sessions = append(c.Sessions, &ClientEchoSession{Session: session})
	c.Lock.Unlock()
}

func (c *EchoClient) RemoveSession(session getty.Session) {
	if session == nil {
		return
	}
	c.Lock.Lock()
	for i, s := range c.Sessions {
		if s.Session == session {
			c.Sessions = append(c.Sessions[:i], c.Sessions[i+1:]...)
			log.Debug("delete session{%s}, its index{%d}", session.Stat(), i)
			break
		}
	}
	log.Info("after remove session{%s}, left session number:%d", session.Stat(), len(c.Sessions))

	c.Lock.Unlock()
}

func (c *EchoClient) UpdateSession(session getty.Session) {
	if session == nil {
		return
	}
	c.Lock.Lock()
	for i, s := range c.Sessions {
		if s.Session == session {
			c.Sessions[i].ReqNum++
			break
		}
	}
	c.Lock.Unlock()
}

func (c *EchoClient) GetClientEchoSession(session getty.Session) (ClientEchoSession, error) {
	var (
		err         error
		echoSession ClientEchoSession
	)
	c.Lock.Lock()
	err = errSessionNotExist
	for _, s := range c.Sessions {
		if s.Session == session {
			echoSession = *s
			err = nil
			break
		}
	}
	c.Lock.Unlock()
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

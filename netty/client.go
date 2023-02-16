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

func init() {
	rand.Seed(time.Now().UnixNano())
}

type EchoClient struct {
	Lock        sync.RWMutex
	Conf        Config
	GettyClient getty.Client
	Session     getty.Session
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
	// 异步执行，如果创建连接失败会重试就会阻塞
	go client.GettyClient.RunEventLoop(newSession(conf, messageHandler, pkgHandler))
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
	if c.Session == nil {
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
		c.Session.Close()
	}
}

func (c *EchoClient) WritePkg(pkg interface{}) (totalBytesLength int, sendBytesLength int, err error) {
	return c.Session.WritePkg(pkg, c.Conf.GettySessionParam.TcpWriteTimeout2)
}

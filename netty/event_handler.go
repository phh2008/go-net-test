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
	"log"
	"time"
)

import (
	"github.com/AlexStocks/getty/transport"
)

type EchoMessageHandler struct {
	Client *EchoClient
}

func NewEchoMessageHandler(client *EchoClient) *EchoMessageHandler {
	return &EchoMessageHandler{
		Client: client,
	}
}

func (h *EchoMessageHandler) OnOpen(session getty.Session) error {
	h.Client.Session = session
	return nil
}

func (h *EchoMessageHandler) OnError(session getty.Session, err error) {
	log.Printf("session{%s} got error{%v}, will be closed.\n", session.Stat(), err)
}

func (h *EchoMessageHandler) OnClose(session getty.Session) {
	log.Printf("session{%s} is closing......\n", session.Stat())
}

func (h *EchoMessageHandler) OnMessage(session getty.Session, pkg interface{}) {
	p, ok := pkg.(*EchoPackage)
	if !ok {
		log.Printf("illegal packge{%#v}\n", pkg)
		return
	}
	log.Printf(">>>>>> get echo package{%s}\n", p)
}

func (h *EchoMessageHandler) OnCron(session getty.Session) {
	conf := h.Client.Conf
	if conf.SessionTimeout2.Nanoseconds() < time.Since(session.GetActive()).Nanoseconds() {
		log.Printf("session{%s} timeout{%s}\n", session.Stat(), time.Since(session.GetActive()).String())

		return
	}
	// 发送心跳包
	var pkg = EchoPackage{
		B: "ping",
	}
	if _, _, err := session.WritePkg(&pkg, conf.GettySessionParam.TcpWriteTimeout2); err != nil {
		log.Printf("session.WritePkg(session{%s}, pkg{%s}) = error{%v}\n", session.Stat(), pkg, err)
		session.Close()
	}
}

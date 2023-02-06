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
	"bytes"
	"errors"
	"fmt"
	getty "github.com/AlexStocks/getty/transport"
	"time"
)

import (
	log "github.com/AlexStocks/getty/util"
)

////////////////////////////////////////////
//  echo command
////////////////////////////////////////////

////////////////////////////////////////////
// EchoPkgHandler
////////////////////////////////////////////

const (
	echoPkgMagic     = 0x20160905
	maxEchoStringLen = 0xff

	echoHeartbeatRequestString  = "ping"
	echoHeartbeatResponseString = "pong"
)

var (
	ErrNotEnoughStream = errors.New("packet stream is not enough")
	ErrTooLargePackage = errors.New("package length is exceed the echo package's legal maximum length")
	ErrIllegalMagic    = errors.New("package magic is not right")
)

type EchoPackage struct {
	B string
}

// String toString
func (p *EchoPackage) String() string {
	return fmt.Sprintf("echo string:%s", p.B)
}

// Marshal 编码
func (p *EchoPackage) Marshal() (*bytes.Buffer, error) {
	var buf = &bytes.Buffer{}
	buf.WriteString(p.B)
	return buf, nil
}

// Unmarshal 解码
func (p *EchoPackage) Unmarshal(buf *bytes.Buffer) (int, error) {
	bt := buf.Bytes()
	p.B = string(bt)
	return len(bt), nil
}

var echoPkgHandler = NewEchoPackageHandler()

type EchoPackageHandler struct{}

func NewEchoPackageHandler() *EchoPackageHandler {
	return &EchoPackageHandler{}
}

func (h *EchoPackageHandler) Read(ss getty.Session, data []byte) (interface{}, int, error) {
	var (
		err error
		len int
		pkg = new(EchoPackage)
		buf *bytes.Buffer
	)

	buf = bytes.NewBuffer(data)
	len, err = pkg.Unmarshal(buf)
	if err != nil {
		if err == ErrNotEnoughStream {
			return nil, 0, nil
		}
		return nil, 0, err
	}
	return pkg, len, nil
}

func (h *EchoPackageHandler) Write(ss getty.Session, pkg interface{}) ([]byte, error) {
	var (
		ok        bool
		err       error
		startTime time.Time
		echoPkg   *EchoPackage
		buf       *bytes.Buffer
	)
	startTime = time.Now()
	if echoPkg, ok = pkg.(*EchoPackage); !ok {
		log.Error("illegal pkg:%+v\n", pkg)
		return nil, errors.New("invalid echo package")
	}
	buf, err = echoPkg.Marshal()
	if err != nil {
		log.Warn("binary.Write(echoPkg{%#v}) = err{%#v}", echoPkg, err)
		return nil, err
	}
	log.Debug("WriteEchoPkgTimeMs = %s", time.Since(startTime).String())
	return buf.Bytes(), nil
}

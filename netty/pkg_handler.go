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
	getty "github.com/AlexStocks/getty/transport"
	"log"
	"strings"
	"time"
)

var (
	ErrNotEnoughStream = errors.New("packet stream is not enough")
)

type EchoPackage struct {
	B string
}

// String toString
func (p *EchoPackage) String() string {
	return p.B
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

type EchoPackageHandler struct{}

func NewEchoPackageHandler() *EchoPackageHandler {
	return &EchoPackageHandler{}
}

func (h *EchoPackageHandler) Read(ss getty.Session, data []byte) (interface{}, int, error) {
	var (
		err    error
		length int
		pkg    = new(EchoPackage)
		buf    *bytes.Buffer
	)
	length = len(data)
	str := string(data)
	if !strings.HasSuffix(str, "\n") {
		return nil, length, nil
	}
	buf = bytes.NewBuffer(data)
	length, err = pkg.Unmarshal(buf)
	if err != nil {
		if err == ErrNotEnoughStream {
			return nil, 0, nil
		}
		return nil, 0, err
	}
	return pkg, length, nil
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
		log.Printf("illegal pkg:%+v\n", pkg)
		return nil, errors.New("invalid echo package")
	}
	buf, err = echoPkg.Marshal()
	if err != nil {
		log.Printf("binary.Write(echoPkg{%#v}\n) = err{%#v}", echoPkg, err)
		return nil, err
	}
	log.Printf("WriteEchoPkgTimeMs = %s\n", time.Since(startTime).String())
	return buf.Bytes(), nil
}

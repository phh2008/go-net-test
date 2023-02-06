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

import "time"

type GettySessionParam struct {
	CompressEncoding bool   `default:"false" yaml:"compress_encoding" mapstructure:"compress_encoding" json:"compress_encoding,omitempty"`
	TcpNoDelay       bool   `default:"true" yaml:"tcp_no_delay" mapstructure:"tcp_no_delay" json:"tcp_no_delay,omitempty"`
	TcpKeepAlive     bool   `default:"true" yaml:"tcp_keep_alive" mapstructure:"tcp_keep_alive" json:"tcp_keep_alive,omitempty"`
	KeepAlivePeriod  string `default:"180s" yaml:"keep_alive_period" mapstructure:"keep_alive_period" json:"keep_alive_period,omitempty"`
	KeepAlivePeriod2 time.Duration
	TcpRBufSize      int    `default:"262144" yaml:"tcp_r_buf_size" mapstructure:"tcp_r_buf_size" json:"tcp_r_buf_size,omitempty"`
	TcpWBufSize      int    `default:"65536" yaml:"tcp_w_buf_size" mapstructure:"tcp_w_buf_size" json:"tcp_w_buf_size,omitempty"`
	TcpReadTimeout   string `default:"1s" yaml:"tcp_read_timeout" mapstructure:"tcp_read_timeout" json:"tcp_read_timeout,omitempty"`
	TcpReadTimeout2  time.Duration
	TcpWriteTimeout  string `default:"5s" yaml:"tcp_write_timeout" mapstructure:"tcp_write_timeout" json:"tcp_write_timeout,omitempty"`
	TcpWriteTimeout2 time.Duration
	WaitTimeout      string `default:"7s" yaml:"wait_timeout" mapstructure:"wait_timeout" json:"wait_timeout,omitempty"`
	WaitTimeout2     time.Duration
	MaxMsgLen        int    `default:"1024" yaml:"max_msg_len" mapstructure:"max_msg_len" json:"max_msg_len,omitempty"`
	SessionName      string `default:"echo-client" yaml:"session_name" mapstructure:"session_name" json:"session_name,omitempty"`
}

// Config holds supported types by the multiconfig package
type Config struct {
	ServerHost        string `default:"127.0.0.1"  yaml:"server_host" mapstructure:"server_host" json:"server_host,omitempty"`
	ServerPort        int    `default:"10000"  yaml:"server_port" mapstructure:"server_port" json:"server_port,omitempty"`
	ConnectionNum     int    `default:"16" yaml:"connection_number" mapstructure:"connection_number" json:"connection_number,omitempty"`
	HeartbeatPeriod   string `default:"15s" yaml:"heartbeat_period" mapstructure:"heartbeat_period" json:"heartbeat_period,omitempty"`
	HeartbeatPeriod2  time.Duration
	SessionTimeout    string `default:"60s" yaml:"session_timeout" mapstructure:"session_timeout" json:"session_timeout,omitempty"`
	SessionTimeout2   time.Duration
	FailFastTimeout   string `default:"5s" yaml:"fail_fast_timeout" mapstructure:"fail_fast_timeout" json:"fail_fast_timeout,omitempty"`
	FailFastTimeout2  time.Duration
	GettySessionParam GettySessionParam `required:"true" yaml:"getty_session_param" mapstructure:"getty_session_param" json:"getty_session_param,omitempty"`
}

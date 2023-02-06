package go_net

import (
	"fmt"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"log"
	"net"
	"net-test/netty"
	"testing"
	"time"
)

func echo(client *netty.EchoClient) {
	var pkg netty.EchoPackage
	pkg.B = "hello,你好！"
	_, _, err := client.WritePkg(&pkg)
	if err != nil {
		log.Printf("session.WritePkg(pkg{%s}) = error{%v}\n", pkg, err)
	}
}

func TestGetty(t *testing.T) {
	// 加载配置
	vp := viper.New()
	vp.SetConfigName("getty-config")
	vp.SetConfigType("yml")
	vp.AddConfigPath(".")
	err := vp.ReadInConfig()
	if err != nil {
		zap.L().Error("加载配置错误")
		panic(err)
	}
	var conf netty.Config
	if err = vp.Unmarshal(&conf); err != nil {
		zap.L().Error("绑定配置出错")
		panic(err)
	}
	conf.FailFastTimeout2, _ = time.ParseDuration(conf.FailFastTimeout)
	conf.SessionTimeout2, _ = time.ParseDuration(conf.SessionTimeout)
	conf.HeartbeatPeriod2, _ = time.ParseDuration(conf.HeartbeatPeriod)
	conf.GettySessionParam.KeepAlivePeriod2, _ = time.ParseDuration(conf.GettySessionParam.KeepAlivePeriod)
	conf.GettySessionParam.TcpReadTimeout2, _ = time.ParseDuration(conf.GettySessionParam.TcpReadTimeout)
	conf.GettySessionParam.TcpWriteTimeout2, _ = time.ParseDuration(conf.GettySessionParam.TcpWriteTimeout)
	conf.GettySessionParam.WaitTimeout2, _ = time.ParseDuration(conf.GettySessionParam.WaitTimeout)

	log.Println(conf)
	client := netty.StartClient(conf)
	if client.IsAvailable() {
		time.Sleep(time.Second * 3)
	}
	for i := 0; i < 100; i++ {
		echo(client)
		time.Sleep(time.Second * 2)
	}
}

func TestTcpServer(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:8085")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Listering on %v\n", listener.Addr())
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Accepted connection to %v from %v\n", conn.LocalAddr(), conn.RemoteAddr())
		go worker(conn)
	}
}

func worker(conn net.Conn) {
	defer conn.Close()
	for {
		var buf [2408]byte
		//接受数据
		n, err := conn.Read(buf[:])
		if err != nil {
			fmt.Printf("read from connect failed, err: %v\n", err)
			break
		}
		log.Printf("**** Received %v bytes, data: %s\n", n, string(buf[:n]))
		_, err = conn.Write([]byte("服务端响应：你好，this is server!"))
		if err != nil {
			fmt.Printf("write to client failed, err: %v\n", err)
			break
		}
	}
}

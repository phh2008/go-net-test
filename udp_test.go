package go_net_test

import (
	"log"
	"net"
	"testing"
)

func TestUdpServer(t *testing.T) {
	//chu
	listen, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(0, 0, 0, 0),
		Port: 30000})
	if err != nil {
		log.Println("err:", err)
		return
	}
	defer listen.Close() //关闭监听
	for true {
		var bf [1024]byte
		udp, addr, err := listen.ReadFromUDP(bf[:]) //接受UDP数据
		if err != nil {
			log.Println("read udp failed,err", err)
			continue
		}
		log.Printf("data:%v,addr:%v", string(bf[:udp]), addr)
		//写入数据
		_, err = listen.WriteToUDP(bf[:udp], addr)
		if err != nil {
			log.Println("write udp error", err)
			continue
		}
	}
}

func TestUdpClient(t *testing.T) {
	//开启通道
	socket, err := net.DialUDP("udp", nil, &net.UDPAddr{IP: net.IPv4(0, 0, 0, 0),
		Port: 30000})
	if err != nil {
		log.Println("err:", err)
		return
	}
	defer socket.Close() //关闭连接
	sendData := []byte("hello server")
	_, err = socket.Write(sendData)
	if err != nil {
		log.Println("发送数据失败", err)
		return
	}
	data := make([]byte, 4096)
	udp, addr, err := socket.ReadFromUDP(data) //接收数据
	if err != nil {
		log.Println("接受数据失败", err)
		return
	}
	log.Printf("recv:%v,addr:%v", string(data[:udp]), addr)
}

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
)

type Config struct {
	Mappings []Mapping `json:"mappings"`
}

type Mapping struct {
	SrcHost string `json:"srchost"`
	DstHost string `json:"dsthost"`
}

var Conf Config
var path = flag.String("path", "./", "app path")

func main() {
	flag.Parse()
	if !LoadConf(*path) {
		fmt.Println("setup.json format error")
		return
	}

	if len(Conf.Mappings) > 0 {

		fmt.Println("======================================")
		fmt.Println("|      Tcp Port Mapping v1.0         |")
		fmt.Println("|   Written By Andy Gu in 2019/7/20  |")
		fmt.Println("|      email:join_gu@sina.com        |")
		fmt.Println("|   github.com:gujunyan/portmapping  |")
		fmt.Println("======================================")

		for i := 0; i < len(Conf.Mappings); i++ {
			go Conf.Mappings[i].Start()
		}

		var c = make(chan string)
		_ = <-c
	} else {
		fmt.Println("please config port mapping relations first")
	}
}

type TcpSession struct {
	TcpConn  *net.TCPConn
	DstTcp   *net.TCPConn
	PortConf Mapping
}

func (m *Mapping) Start() {

	tcpAddr, err := net.ResolveTCPAddr("tcp", m.SrcHost)
	tcpListener, err := net.ListenTCP("tcp", tcpAddr)
	defer tcpListener.Close()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println("mapping ", m.SrcHost, "->", m.DstHost)
	for {
		tcpConn, err1 := tcpListener.AcceptTCP()

		if err1 == nil {
			session := TcpSession{
				TcpConn:  tcpConn,
				PortConf: (*m),
			}
			go session.Start()
		}
	}
}

func (m *TcpSession) Start() {
	var err error
	fmt.Print("connected from ", m.TcpConn.RemoteAddr().String(), " -> ", m.TcpConn.LocalAddr().String())
	tcpAddr, _ := net.ResolveTCPAddr("tcp", m.PortConf.DstHost)
	fmt.Print(" connecting to ", tcpAddr.String(), "...")
	m.DstTcp, err = net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		fmt.Println("error", err.Error())
		m.TcpConn.Close()
		return
	}
	fmt.Println("succ!")
	go m.ClientLoop()
	go m.ServerLoop()

}

func (m *TcpSession) ClientLoop() {
	var err error
	var len int
	var buf = make([]byte, 1024)
	for {
		len, err = m.TcpConn.Read(buf)
		if err != nil {
			m.DstTcp.Close()
			break
		}
		m.DstTcp.Write(buf[:len])
	}
	m.TcpConn.Close()
}

func (m *TcpSession) ServerLoop() {
	var err error
	var len int
	var buf = make([]byte, 1024)
	for {
		len, err = m.DstTcp.Read(buf)
		if err != nil {
			m.TcpConn.Close()
			break
		}
		m.TcpConn.Write(buf[:len])
	}
	m.DstTcp.Close()
}

func LoadConf(p string) bool {
	data, err := ioutil.ReadFile(p + "setup.json")
	if err != nil {
		fmt.Println(p + "setup.json not found")
		return false
	}

	err = json.Unmarshal(data, &Conf)
	if err != nil {
		fmt.Println(p + "setup.json format error:" + err.Error())
		return false
	}
	return true
}

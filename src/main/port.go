package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func (sm *SafeMap) pushIp(port string,service string){
	sm.Lock()
	Porting := sm.Map[port]
	if len(service)>0 && len(Porting.Service) == 0 {
		Porting.Service = service
	}
	Porting.State = "open"
	sm.Map[port] = Porting
	sm.Unlock()
}

func (sm *SafeMap) checkPort(ip net.IP, port int, parallelChan *chan int) {
	tcpAddr := net.TCPAddr{
		IP:   ip,
		Port: port,
	}

	conn, err := net.DialTCP("tcp", nil, &tcpAddr)

	if err == nil {
		//sm.printOpeningPort( port,"tcp" )
		t = time.NewTicker(3 * time.Second)
		service = ""
		sm.pushIp(strconv.Itoa(port)+"/tcp", service)
		fmt.Fprintf(conn,"hello\r\n")
		buf:=make([]byte,0,4096)
		tmp:=make([]byte,256)
		for {
			n,err:=conn.Read(tmp)
			if err!=nil{
				if err!=io.EOF{
					fmt.Println("read error:",err)
				}
				break
			}
			buf=append(buf,tmp[:n]...)
		}
		conn.Close()
		if strings.Contains(string(buf[:]), "SSH"){
			service = "ssh"
		}
		if strings.Contains(string(buf[:]), "HTTP"){
			service = "http"
		}
		sm.pushIp(strconv.Itoa(port)+"/tcp", service)
	}
	<-*parallelChan
}

func (sm *SafeMap) portScan(ctx context.Context,ip net.IP) {
	if port !=0 {
		parallelChan := make(chan int)
		go func() {
			parallelChan <- 1
		}()
		go sm.checkPort(ip, port, &parallelChan)
	} else{
		matched, _ := regexp.Match(`^\d+~\d+$`, []byte(portRange))
		if !matched {
			flag.Usage()
		} else {
			portSecs := strings.Split(portRange, "~")
			startPort, err1 := strconv.Atoi(portSecs[0])
			endPort, err2 := strconv.Atoi(portSecs[1])
			if err1 != nil || err2 != nil || startPort < 1 || endPort < 2 || endPort <= startPort || parallelCounts < 1 {
				flag.Usage()
			} else {
				parallelChan := make(chan int, parallelCounts)
				for i := startPort; i <= endPort; i++ {
					parallelChan <- 1
					go sm.checkPort(ip, i, &parallelChan)
				//sm.PrintDataip()
				}
			}
		}
	}
	for {
		select {
		case <-ctx.Done():
			return
		default:

		}
	}

}

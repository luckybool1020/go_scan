package main

import (
    "context"
    "flag"
    "fmt"
    "github.com/Sirupsen/logrus"
    manuf "github.com/timest/gomanuf"
    "net"
    "os"
    "sort"
    "strings"
    "sync"
    "time"
)

var log = logrus.New()
// ipNet 存放 IP地址和子网掩码
var ipNet *net.IPNet
// 本机的mac地址，发以太网包需要用到
var localHaddr net.HardwareAddr
var iface string
// 存放最终的数据，key[string] 存放的是IP地址
var data map[string]Info
var dataip map[string]Porting
// 计时器，在一段时间没有新的数据写入data中，退出程序，反之重置计时器
var t *time.Ticker
var do chan string
//var doip chan string
var port int
var portRange string
var service string

var parallelCounts int

const (
    // 3秒的计时器
    START = "start"
    END = "end"
)

type Info struct {
    // IP地址
    // Mac地址
    Mac net.HardwareAddr
    // 主机名
    Hostname string
    // 厂商信息
    Manuf   string
}

type Porting struct {
    // IP地址
    // Mac地址
    // 主机名
    State string
    // 厂商信息
    Service string
}

type SafeMap struct {
    sync.RWMutex
    Map map[string]Porting
}

// 格式化输出结果
// xxx.xxx.xxx.xxx  xx:xx:xx:xx:xx:xx  hostname  manuf
// xxx.xxx.xxx.xxx  xx:xx:xx:xx:xx:xx  hostname  manuf

func PrintDataip(dataip map[string]Porting){
    //var keys string
    //for k := range dataip {
    //    keys = append(keys, ParseIPString(k))
    //}
    //sort.Sort(keys)
    fmt.Printf("%-15s %-17s %-30s \n", "Port", "State", "Service")
    for k := range dataip{
        d := dataip[k]
        //fmt.Printf("%-15s %-17s %-30s %-10s\n", "Ip", "Mac", "Hostname", "Manuf")
        fmt.Printf("%-15s %-17s %-30s \n",  k, d.State, d.Service)
    }

}

func PrintData() {
    var keys IPSlice
    for k := range data {
        keys = append(keys, ParseIPString(k))
    }
    sort.Sort(keys)
    fmt.Printf("%-15s %-17s %-30s %-10s\n", "Ip", "Mac", "Hostname", "Manuf")
    for _, k := range keys {
        d := data[k.String()]
        mac := ""
        if d.Mac != nil {
            mac = d.Mac.String()
        }
        //fmt.Printf("%-15s %-17s %-30s %-10s\n", "Ip", "Mac", "Hostname", "Manuf")
        fmt.Printf("%-15s %-17s %-30s %-10s\n", k.String(), mac, d.Hostname, d.Manuf)
    }
}

// 将抓到的数据集加入到data中，同时重置计时器
func pushData(ip string, mac net.HardwareAddr, hostname, manuf string) {
    // 停止计时器c
    do <- START
    var mu sync.RWMutex
    mu.RLock()
    defer func() {
        // 重置计时器
        do <- END
        mu.RUnlock()
    }()
    if _, ok := data[ip]; !ok {
        data[ip] = Info{Mac: mac, Hostname: hostname, Manuf: manuf}
        return
    }
    info := data[ip]
    if len(hostname) > 0 && len(info.Hostname) == 0 {
        info.Hostname = hostname
    }
    if len(manuf) > 0 && len(info.Manuf) == 0 {
        info.Manuf = manuf
    }
    if mac != nil {
        info.Mac = mac
    }
    data[ip] = info
}

func setupNetInfo(f string) {
    var ifs []net.Interface
    var err error
    if f == "" {
        ifs, err = net.Interfaces()
    } else {
        // 已经选择iface
        var it *net.Interface
        it, err = net.InterfaceByName(f)
        if err == nil {
            ifs = append(ifs, *it)
        }
    }
    if err != nil {
        log.Fatal("无法获取本地网络信息:", err)
    }
    for _, it := range ifs {
        addr, _ := it.Addrs()
        for _, a := range addr {
            if ip, ok := a.(*net.IPNet); ok && !ip.IP.IsLoopback() {
                if ip.IP.To4() != nil {
                    ipNet = ip
                    localHaddr = it.HardwareAddr
                    iface = it.Name
                    goto END
                }
            }
        }
    }
    END:
    if ipNet == nil || len(localHaddr) == 0 {
        log.Fatal("无法获取本地网络信息")
    }
}

func localHost() {
    host, _ := os.Hostname()
    data[ipNet.IP.String()] = Info{Mac: localHaddr, Hostname: strings.TrimSuffix(host, ".local"), Manuf: manuf.Search(localHaddr.String())}
}

func sendARP() {
    // ips 是内网IP地址集合
    ips := Table(ipNet)
    for _, ip := range ips {
        go sendArpPackage(ip)
    }
}

func newSafeMap() *SafeMap {
    sm := new(SafeMap)
    sm.Map = make(map[string]Porting)
    return sm

}

func main() {
    // allow non root user to execute by compare with euid
    if os.Geteuid() != 0 {
        log.Fatal("goscan must run as root.")
    }
    flag.StringVar(&iface, "I", "", "Network interface name")
    flag.IntVar(&port, "p", 0, "port")
    flag.StringVar(&portRange, "r", "22~8080", "range ports. format is <from>~<to>. eg. 100~200")
    flag.IntVar(&parallelCounts, "n", 64, "parallel counts")

    flag.Usage = func() {
        fmt.Fprintf(os.Stderr, "\nUsage: %s [Options] <IP>\n\nOptions:\n\n", os.Args[0])
        flag.PrintDefaults()
    }
    flag.Parse()
    args := flag.Args()

    if len(args) > 1{
        flag.Usage()
    } else if len(args) == 1 {
        //ch := make(chan bool)
        //doip = make(chan string)
        //dataip = make(map[string]Porting)
        ip := net.ParseIP(flag.Arg(0))
        fmt.Println("scan report for",flag.Arg(0))
        safeMap := newSafeMap()
        //ctx, _ := context.WithTimeout(context.Background(), timeout)
        ctx, cancel := context.WithCancel(context.Background())
        go safeMap.portScan(ctx,ip)
        //t = time.NewTicker(4 * time.Second)
        time.Sleep(6 * time.Second)
        PrintDataip(safeMap.Map)
        cancel()
    } else {

        // 初始化 data
        data = make(map[string]Info)
        do = make(chan string)
        // 初始化 网络信息
        setupNetInfo(iface)

        ctx, cancel := context.WithCancel(context.Background())
        go listenARP(ctx)
        go listenMDNS(ctx)
        go listenNBNS(ctx)
        go sendARP()
        go localHost()

        t = time.NewTicker(4 * time.Second)

        for {
            select {
            case <-t.C:
                PrintData()
                cancel()
                goto END1
            case d := <-do:
                switch d {
                case START:
                    t.Stop()
                case END:
                    // 接收到新数据，重置2秒的计数器
                    t = time.NewTicker(2 * time.Second)
                }
            }
        }
    END1:
    }
    
}


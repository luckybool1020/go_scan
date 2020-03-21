# goscan

**程序思路：**
 * 根据ip和子网掩码计算出ip范围
 * 广播ARP Request
 * 监听并抓取ARP Response包，记录IP和Mac地址
 * 向活跃IP发送MDNS和NBNS包，并监听和解析Hostname
 * 利用MAC地址计算设备的厂商信息
 * 多线程嗅探端口
 * 多线程嗅探端口
 
 **端口扫描：**
 * TCP:
 ```
1> 根据端口的输入参数范围开启扫描端口的goroutine，每个goroutine用一个chan来进行通信
2> 利用net内置的函数DialTCP对TCP端口状态进行判断，若开启DialTCP返回的err为nil，若关闭则err不为nil
3> 若相应的端口err为nil，则对该端口进行通信获取service信息。
 ```
 * UDP:
 ```

 ```
 
### Usage: ###

```sh
# install dependencies
$ go get github.com/Sirupsen/logrus
$ go get github.com/timest/gomanuf
$ go get github.com/google/gopacket

# build
$ go build

# 扫描网络信息
$ sudo ./main  
# 指定网卡扫描
$ sudo ./main -I en0
# 指定ip扫描,默认扫描22～8080端口状态
$ sudo ./main 10.11.11.150
# 用参数-p对ip指定某端口扫描
$ sudo ./main -p 22 10.11.11.150
# 用参数-p对ip指定端口范围扫描
$ sudo ./main -r 22~27017 10.11.11.150
# 用参数-n指定并发数量
$ sudo ./main -n 8 -r 22~27017 10.11.11.150
```

扫描器必须以**root**运行.


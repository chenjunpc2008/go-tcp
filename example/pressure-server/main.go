package main

import (
    "flag"
    "log"
    "runtime"
    "strings"
    "time"

    "github.com/chenjunpc2008/go-tcp/tcpserver"
    "github.com/chenjunpc2008/go/util/onlinepprof"
)

const (
    dataSize1k = "1k"
    dataSize2k = "2k"
    dataSize3k = "3k"
    dataSize4k = "4k"
)

var (
    gPprofPort uint
    gPort      uint

    gDataSize string

    gChExit chan int

    gTotalReceivedPkg uint64
    gTotalSendedPkg   uint64

    gServer *tcpserver.Ctcpsvr
)

func init() {
    flag.UintVar(&gPprofPort, "pprof-port", 10010, "net pprof listen port")
    flag.UintVar(&gPort, "port", 8019, "server listen port")
    flag.StringVar(&gDataSize, "data-size", "1k", "send data size")

    gChExit = make(chan int)
}

func main() {

    var err error

    runtime.GOMAXPROCS(200)
    log.Println("runtime.GOMAXPROCS", 200)

    // pprof
    _, err = onlinepprof.StartOnlinePprof(true, uint16(gPprofPort), true)
    if nil != err {
        log.Panicln("StartServer failed", err)
    }

    appHdl := &appHandler{}
    cnf := tcpserver.DefaultConfig()

    gServer = tcpserver.NewTCPSvr(appHdl, cnf)

    err = gServer.StartServer(uint16(gPort))
    if nil != err {
        log.Panicln("StartServer failed", err)
    }

    sysMonitor()
}

func sysMonitor() {

    var (
        timeout = 10 * time.Second

        lastTotalReceivedPkg uint64
        lastTotalSendedPkg   uint64
        nowTotalReceivedPkg  uint64
        nowTotalSendedPkg    uint64
    )

    for {
        select {
        case <-gChExit:

        case <-time.After(timeout):
            nowTotalReceivedPkg = gTotalReceivedPkg
            nowTotalSendedPkg = gTotalSendedPkg

            log.Printf("\n TotalReceivedPkg:%d, tps:[%d],\n TotalSendedPkg:%d, tps:[%d]\n",
                nowTotalReceivedPkg, (nowTotalReceivedPkg-lastTotalReceivedPkg)/10,
                nowTotalSendedPkg, (nowTotalSendedPkg-lastTotalSendedPkg)/10)

            lastTotalReceivedPkg = nowTotalReceivedPkg
            lastTotalSendedPkg = nowTotalSendedPkg
        }
    }
}

func prepareData(dataSize string) string {
    var (
        buff strings.Builder
        msg  string
    )

    switch dataSize {
    case dataSize1k:
        for i := 0; i < 100; i++ {
            buff.WriteString("1234567890")
        }

    case dataSize2k:
        for i := 0; i < 203; i++ {
            buff.WriteString("1234567890")
        }

    case dataSize3k:
        for i := 0; i < 306; i++ {
            buff.WriteString("1234567890")
        }

    case dataSize4k:
        for i := 0; i < 408; i++ {
            buff.WriteString("1234567890")
        }

    default:
        // 1k
        for i := 0; i < 100; i++ {
            buff.WriteString("1234567890")
        }
    }

    msg = buff.String()
    return msg
}

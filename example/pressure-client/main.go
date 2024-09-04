package main

import (
	"flag"
	"log"
	"runtime"
	"strings"
	"time"

	"github.com/chenjunpc2008/go-tcp/tcpclient"
	"github.com/chenjunpc2008/go/util/onlinepprof"
)

const (
	dataSize1k = "1k"
	dataSize2k = "2k"
	dataSize3k = "3k"
	dataSize4k = "4k"
)

var (
	gPprofPort  uint
	gRemoteIP   string
	gRemotePort uint

	gClientNums int
	gSendRate   int
	gDataSize   string

	gChExit           chan int
	gTotalReceivedPkg uint64
	gTotalSendedPkg   uint64
)

func init() {
	flag.UintVar(&gPprofPort, "pprof-port", 10012, "net pprof listen port")
	flag.StringVar(&gRemoteIP, "rip", "127.0.0.1", "remote server ip")
	flag.UintVar(&gRemotePort, "rport", 8019, "remote server port")

	flag.IntVar(&gClientNums, "clients", 100, "number of clients")
	flag.IntVar(&gSendRate, "send-rate", 1000, "send rate per client per second")

	flag.StringVar(&gDataSize, "data-size", "1k", "send data size")

	gChExit = make(chan int)
}

func main() {

	runtime.GOMAXPROCS(200)
	log.Println("runtime.GOMAXPROCS", 200)

	var err error

	// pprof
	_, err = onlinepprof.StartOnlinePprof(true, uint16(gPprofPort), true)
	if nil != err {
		log.Panicln("StartServer failed", err)
	}

	for i := 0; i < gClientNums; i++ {
		go coPressureSender()
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

func coPressureSender() {
	const ftag = "coPressureSender()"

	time.Sleep(2 * time.Second)

	var (
		timeout = time.Duration(1) * time.Second
		msg     string
		err     error
		busy    bool
	)

	appHdl := &appHandler{}
	cnf := tcpclient.DefaultConfig()

	client := tcpclient.New(appHdl, cnf)

	cnct_to := 3 * time.Second
	err = client.ConnectToServer_Timeout(gRemoteIP, uint16(gRemotePort), cnct_to)
	if nil != err {
		log.Panicln("ConnectToServer_Timeout failed", err)
	}

	for {
		select {
		case <-gChExit:
			// exit
			return

		case <-time.After(timeout):
			// sleep gap
		}

		//
		for i := 0; i < gSendRate; i++ {
			msg = prepareData(gDataSize)

			busy, err = client.SendToServer(msg)
			if nil != err {
				log.Println("SendToServer failed", err)
				break
			}

			if busy {
				log.Println("SendToServer failed because busy")
				break
			}
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

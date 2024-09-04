package tcpclient

import (
	"errors"
	"net"
	"sync"
	"time"
)

const (
	// DefaultSendBuffSize default send buff size
	DefaultSendBuffSize = 1 * 10000
)

// Config extra config
type Config struct {
	// tcp send buff size
	SendBuffsize int

	// after recieve a whole package, the receive callback will go async or sync
	AsyncReceive bool

	// after data write to tcp sys buff, do OnSendedData() call back require
	RequireSendedCb bool
	// sended callback if async or sync
	AsyncSended bool
}

// DefaultConfig default Config
func DefaultConfig() Config {
	return Config{
		SendBuffsize:    DefaultSendBuffSize,
		AsyncReceive:    true,
		RequireSendedCb: true,
		AsyncSended:     true,
	}
}

// CtcpCli tcp client
type CtcpCli struct {
	svrIP   string
	svrPort uint16
	cnf     Config

	conn         net.Conn
	bIsConnected bool
	handler      EventHandler
	chExit       chan int // notify all goroutines to shutdown

	lock           sync.Mutex // lock for below values
	chMsgsToBeSend chan interface{}
}

// New new tcp client
func New(eventCb EventHandler, cnf Config) *CtcpCli {

	var cli = &CtcpCli{bIsConnected: false}

	cli.chExit = make(chan int)
	cli.handler = eventCb
	cli.cnf = cnf
	cli.chMsgsToBeSend = make(chan interface{}, cli.cnf.SendBuffsize)

	return cli
}

// ConnectToServer connect to remote server
func (cli *CtcpCli) ConnectToServer(ip string, port uint16) error {
	return connectToServer(ip, port, cli)
}

// ConnectToServer connect to remote server
func (cli *CtcpCli) ConnectToServer_Timeout(ip string, port uint16, timeout time.Duration) error {
	return connectToServer_Timeout(ip, port, timeout, cli)
}

// Close close connection
func (cli *CtcpCli) Close() {
	// lock
	cli.lock.Lock()
	defer cli.lock.Unlock()

	if nil != cli.conn {
		close(cli.chExit)
		cli.conn.Close()
		cli.conn = nil
	}

	cli.svrIP = ""
	cli.svrPort = 0
	cli.bIsConnected = false

	if nil != cli.chMsgsToBeSend {
		close(cli.chMsgsToBeSend)
		cli.chMsgsToBeSend = nil
	}
}

// new conn
func (cli *CtcpCli) addNewConnection(conn net.Conn, ip string, port uint16) {
	// lock
	cli.lock.Lock()
	defer cli.lock.Unlock()

	cli.conn = conn
	cli.svrIP = ip
	cli.svrPort = port
	cli.bIsConnected = true
	cli.chMsgsToBeSend = make(chan interface{}, cli.cnf.SendBuffsize)
}

// disconnected info update
func (cli *CtcpCli) disconnected() {
	// lock
	cli.lock.Lock()
	defer cli.lock.Unlock()

	cli.conn = nil
	cli.svrIP = ""
	cli.svrPort = 0
	cli.bIsConnected = false

	if nil != cli.chMsgsToBeSend {
		close(cli.chMsgsToBeSend)
		cli.chMsgsToBeSend = nil
	}
}

/*
SendToServer send message to server

@return busy bool : true -- buff is full, you may need to try again
*/
func (cli *CtcpCli) SendToServer(msg interface{}) (busy bool, retErr error) {
	// lock
	cli.lock.Lock()

	if !cli.bIsConnected {
		// unlock
		cli.lock.Unlock()
		return false, errors.New("not connected")
	}

	if nil == cli.chMsgsToBeSend {
		// unlock
		cli.lock.Unlock()
		return false, errors.New("nil chMsgsToBeSend")
	}

	curBuffSize := len(cli.chMsgsToBeSend)

	if curBuffSize >= cli.cnf.SendBuffsize-1 {
		// unlock
		cli.lock.Unlock()
		return true, nil
	}

	// push
	cli.chMsgsToBeSend <- msg

	// unlock
	cli.lock.Unlock()

	return false, nil
}

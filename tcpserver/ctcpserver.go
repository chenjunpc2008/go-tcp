package tcpserver

import (
    "errors"
    "fmt"
    "net"
    "sync"
)

const (
    // DefaultSendBuffSize default send buff size
    DefaultSendBuffSize = 1 * 10000
)

func init() {

}

// ServerStatus server status
type ServerStatus int

const (
    // ServerStatusClosed closed
    ServerStatusClosed ServerStatus = 0
    // ServerStatusStarting starting
    ServerStatusStarting ServerStatus = 1
    // ServerStatusRunning running
    ServerStatusRunning ServerStatus = 2
    // ServerStatusStopping stopping
    ServerStatusStopping ServerStatus = 3
)

// Config extra config
type Config struct {
    // tcp send buff size
    SendBuffsize int

    // AsyncReceive after recieve a whole package, the receive callback will go async or sync
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

/*
CtcpsvrSt tcp server instance
*/
type Ctcpsvr struct {
    port uint16
    cnf  Config

    listener  *net.TCPListener
    clientID  uint64
    cliSns    clientSnsSt
    handler   EventHandler
    chExit    chan int        // notify all goroutines to shutdown
    wg        *sync.WaitGroup // wait for all goroutines to exit
    svrStatus ServerStatus    // server status
}

/*
NewTCPSvr new ctcpserver object
*/
func NewTCPSvr(eventCb EventHandler, cnf Config) *Ctcpsvr {
    var svr *Ctcpsvr = &Ctcpsvr{clientID: 0}
    svr.cnf = cnf

    svr.cliSns.init()
    svr.chExit = make(chan int)
    svr.wg = &sync.WaitGroup{}

    svr.handler = eventCb

    svr.svrStatus = ServerStatusClosed

    return svr
}

// StartServer start server
func (t *Ctcpsvr) StartServer(iPort uint16) error {

    if ServerStatusClosed != t.svrStatus {
        errMsg := fmt.Sprintf("couldn't start server in status:%d", t.svrStatus)
        return errors.New(errMsg)
    }

    t.svrStatus = ServerStatusStarting

    listener, err := createListen(iPort, t)
    if nil != err {
        return err
    }

    go startServer(iPort, listener, t)

    return nil
}

// StopServer top server
func (t *Ctcpsvr) StopServer() error {

    if ServerStatusRunning != t.svrStatus {
        errMsg := fmt.Sprintf("couldn't stop server in status:%d", t.svrStatus)
        return errors.New(errMsg)
    }

    stopServer(t)

    return nil
}

// SendToClient sent to client
func (t *Ctcpsvr) SendToClient(clientID uint64, msg interface{}) (busy bool, retErr error) {
    if ServerStatusRunning != t.svrStatus {
        errMsg := fmt.Sprintf("couldn't SendToClient, server in status:%d", t.svrStatus)
        return false, errors.New(errMsg)
    }

    cli, ok := t.cliSns.getClientSession(clientID)
    if !ok {
        return false, errors.New("couldn't getClientSession")
    }

    // nil
    if nil == cli {
        return false, errors.New("nil ClientSession")
    }

    busy, retErr = cli.putSendMsg(msg)

    return
}

// SendToAllClients send to all clients
func (t *Ctcpsvr) SendToAllClients(msg interface{}) error {
    if ServerStatusRunning != t.svrStatus {
        errMsg := fmt.Sprintf("couldn't SendToClient, server in status:%d", t.svrStatus)
        return errors.New(errMsg)
    }

    cliIDs := t.cliSns.getAllClientIDs()
    // nil
    if nil == cliIDs {
        return errors.New("nil []cli-ids")
    }

    var (
        err  error
        cli  *clientSessnSt
        ok   bool
        busy bool
    )

    for _, v := range cliIDs {
        cli, ok = t.cliSns.getClientSession(v)
        if !ok {
            t.handler.OnCliErrorStr(v, "", "", "couldn't getClientSession")
            continue
        }

        // nil
        if nil == cli {
            t.handler.OnCliErrorStr(v, "", "", "nil ClientSession")
            continue
        }

        busy, err = cli.putSendMsg(msg)
        if nil != err {
            t.handler.OnCliError(v, "", "", "putSendMsg", err)
            continue
        } else if busy {
            t.handler.OnCliError(v, "", "", "putSendMsg busy", nil)
            continue
        }
    }

    return nil
}

// CloseClient close one client
func (t *Ctcpsvr) CloseClient(clientID uint64, reason string) {
    closeCli(clientID, reason, t)
}

// CloseClients close clients
func (t *Ctcpsvr) CloseClients(clientIDs []uint64, reason string) {
    closeClients(clientIDs, reason, t)
}

// CliDebugInfoSt client debug info
type CliDebugInfoSt struct {
    ClientID     uint64 `json:"ClientID"`
    Addr         string `json:"Addr"`
    SendBuffSize int    `json:"SendBuffSize"`
}

// GetDebugInfo clients debug infos
func (t *Ctcpsvr) GetDebugInfo() (debug []CliDebugInfoSt) {
    if ServerStatusRunning != t.svrStatus {
        return
    }

    debug = t.cliSns.getDebugInfos()

    return
}

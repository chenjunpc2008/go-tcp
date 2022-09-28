package tcpserver

import (
    "errors"
    "net"
    "sync"
)

/*
single client session
*/
type clientSessnSt struct {
    conn net.Conn
    id   uint64
    ip   string
    addr string

    lock           sync.Mutex // lock for below values
    sendBuffsize   int
    chMsgsToBeSend chan interface{}

    closeOnce sync.Once // close the conn, once/per instance
    closed    bool
}

func newClientSessnSt(c net.Conn, clientID uint64, cliIP string, cliAddr string, sendBuffsize int) *clientSessnSt {
    return &clientSessnSt{conn: c, id: clientID,
        ip: cliIP, addr: cliAddr,
        sendBuffsize:   sendBuffsize,
        chMsgsToBeSend: make(chan interface{}, sendBuffsize),
        closed:         false}
}

// close client session
func (sn *clientSessnSt) close() {
    // lock
    sn.lock.Lock()
    defer sn.lock.Unlock()

    sn.closeOnce.Do(func() {
        sn.conn.Close()
        sn.closed = true

        close(sn.chMsgsToBeSend)
        sn.chMsgsToBeSend = nil
    })
}

/*
@return busy bool : true -- buff is full, you may need to try again
*/
func (sn *clientSessnSt) putSendMsg(msg interface{}) (busy bool, retErr error) {
    // lock
    sn.lock.Lock()

    if sn.closed {
        // unlock
        sn.lock.Unlock()
        return false, errors.New("client closed")
    }

    if nil == sn.chMsgsToBeSend {
        // unlock
        sn.lock.Unlock()
        return false, errors.New("nil msgsToBeSend")
    }

    curBuffSize := len(sn.chMsgsToBeSend)

    if curBuffSize >= sn.sendBuffsize-1 {
        // unlock
        sn.lock.Unlock()
        return true, nil
    }

    // push
    sn.chMsgsToBeSend <- msg

    // unlock
    sn.lock.Unlock()

    return false, nil
}

/*
@return busy bool : true -- buff is full, you may need to try again
*/
func (sn *clientSessnSt) getDebugInfo() (debug CliDebugInfoSt) {
    // lock
    sn.lock.Lock()

    if sn.closed {
        // unlock
        sn.lock.Unlock()
        return
    }

    if nil == sn.chMsgsToBeSend {
        // unlock
        sn.lock.Unlock()
        return
    }

    debug.ClientID = sn.id
    debug.Addr = sn.addr
    debug.SendBuffSize = len(sn.chMsgsToBeSend)

    // unlock
    sn.lock.Unlock()

    return
}

/*
client sessions
*/
type clientSnsSt struct {
    initOnce sync.Once // init once

    // lock for values below
    lock sync.Mutex
    // key-clientID
    mapCliSess map[uint64]*clientSessnSt
}

func (sns *clientSnsSt) init() {
    sns.initOnce.Do(func() {
        sns.mapCliSess = make(map[uint64]*clientSessnSt, 0)
    })
}

// add a new client session
func (sns *clientSnsSt) addNewConnection(c net.Conn, clientID uint64, cliIP string, cliAddr string, sendBuffsize int) *clientSessnSt {
    // lock
    sns.lock.Lock()
    defer sns.lock.Unlock()

    var cliSessn = newClientSessnSt(c, clientID, cliIP, cliAddr, sendBuffsize)

    sns.mapCliSess[clientID] = cliSessn

    // fmt.Printf("%v on new connection, client-id:=%v, ip:%v, addr:%v\n", ftag, clientID, cliIP, cliAddr)
    return cliSessn
}

// delete connetion
func (sns *clientSnsSt) delConnect(clientID uint64, cliIP string, cliAddr string) {
    // lock
    sns.lock.Lock()
    defer sns.lock.Unlock()

    delete(sns.mapCliSess, clientID)

    // fmt.Printf("%v delConnect, client-id:=%v, ip:%v, addr:%v\n", ftag, clientID, cliIP, cliAddr)
}

// get all client ids
func (sns *clientSnsSt) getAllClientIDs() []uint64 {
    // lock
    sns.lock.Lock()
    defer sns.lock.Unlock()

    var clientids = make([]uint64, 0)
    for k := range sns.mapCliSess {
        clientids = append(clientids, k)
    }

    return clientids
}

// get client session object
func (sns *clientSnsSt) getClientSession(clientID uint64) (*clientSessnSt, bool) {
    // lock
    sns.lock.Lock()
    defer sns.lock.Unlock()

    cli, ok := sns.mapCliSess[clientID]
    if !ok {
        return nil, false
    }

    return cli, true
}

// get client session object
func (sns *clientSnsSt) getDebugInfos() (infos []CliDebugInfoSt) {
    // lock
    sns.lock.Lock()
    defer sns.lock.Unlock()

    var (
        cliInfo CliDebugInfoSt
    )

    for _, v := range sns.mapCliSess {
        cliInfo = v.getDebugInfo()
        infos = append(infos, cliInfo)
    }

    return
}

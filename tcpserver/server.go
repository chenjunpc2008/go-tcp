package tcpserver

import (
    "fmt"
    "io"
    "net"
    "strconv"
    "strings"
)

// listen
func createListen(iPort uint16, svr *Ctcpsvr) (*net.TCPListener, error) {
    var (
        listener *net.TCPListener
        err      error
    )

    sPort := strconv.FormatInt(int64(iPort), 10)
    tcpAddr, err := net.ResolveTCPAddr("tcp4", ":"+sPort)
    if nil != err {
        svr.handler.OnError("ResolveTCPAddr failed", err)
        return nil, err
    }

    listener, err = net.ListenTCP("tcp", tcpAddr)
    if nil != err {
        svr.handler.OnError("ListenTCP failed", err)
        return nil, err
    }

    return listener, nil
}

/*
start tcp server

@return nil : success
*/
func startServer(iPort uint16, listener *net.TCPListener, svr *Ctcpsvr) error {
    svr.wg.Add(1)
    defer svr.wg.Done()

    svr.listener = listener
    svr.port = iPort
    svr.svrStatus = ServerStatusRunning

    for {
        conn, err := listener.Accept()
        if nil != err {
            select {
            case <-svr.chExit:
                return nil

            default:
                svr.handler.OnError("Accept failed", err)
                return err
            }
        }

        svr.clientID++

        go handleNewIncoming(conn, svr.clientID, svr)
    }

    // server accept loop out
}

// stop server
func stopServer(svr *Ctcpsvr) {

    svr.svrStatus = ServerStatusStopping

    close(svr.chExit)

    if nil != svr.listener {
        svr.listener.Close()
        svr.listener = nil
    }

    // close all clients
    clis := svr.cliSns.getAllClientIDs()
    closeClients(clis, "server shutdown", svr)

    svr.wg.Wait()

    svr.svrStatus = ServerStatusClosed
}

func handleNewIncoming(conn net.Conn, clientID uint64, svr *Ctcpsvr) {
    var (
        sIP      string
        cliSessn *clientSessnSt
    )

    sAddr := conn.RemoteAddr().String()
    arySplits := strings.Split(sAddr, ":")
    if 0 == len(arySplits) || 2 != len(arySplits) {
        sIP = sAddr
    } else {
        sIP = arySplits[0]
    }

    cliSessn = svrNewConnection(conn, clientID, sIP, sAddr, svr)

    go cliLoopRead(conn, clientID, sIP, sAddr, cliSessn, svr, svr.cnf.AsyncReceive)
    go cliLoopSend(conn, clientID, sIP, sAddr, cliSessn.chMsgsToBeSend, cliSessn, svr,
        svr.cnf.RequireSendedCb, svr.cnf.AsyncSended)
}

func closeCliConn(conn net.Conn, clientID uint64, cliIP string, cliAddr string, cliSessn *clientSessnSt, svr *Ctcpsvr) {
    if nil == cliSessn {
        return
    }

    cliSessn.close()
    svrDisconnect(clientID, cliIP, cliAddr, svr)
}

// loop read for one client
func cliLoopRead(conn net.Conn, clientID uint64, cliIP string, cliAddr string, cliSessn *clientSessnSt,
    svr *Ctcpsvr, asyncReceive bool) {
    // const ftag = "cliLoopRead()"

    defer closeCliConn(conn, clientID, cliIP, cliAddr, cliSessn, svr)

    // close count
    svr.wg.Add(1)
    defer svr.wg.Done()

    var (
        allbuf            = make([]byte, 0)
        buffer            = make([]byte, 4096)
        byAfterDepackBuff []byte
        lenRcv            int
        err               error
    )

    for {
        lenRcv, err = conn.Read(buffer)

        if nil != err {
            select {
            case <-svr.chExit:
                // server close
                // fmt.Printf("%v closing of client chExit-1, clientID:%v, ip:%v, addr:%v\n", ftag, clientID, cliIP, cliAddr)

                // manually empty
                allbuf = make([]byte, 0)
                return

            default:
            }

            if io.EOF == err {
                svr.handler.OnCliErrorStr(clientID, cliIP, cliAddr, "read from conn EOF")

                // manually empty
                allbuf = make([]byte, 0)
                break
            } else {
                svr.handler.OnCliError(clientID, cliIP, cliAddr, "read from conn failed", err)

                // manually empty
                allbuf = make([]byte, 0)
                break
            }
        }

        if 0 == lenRcv {
            continue
        }

        allbuf = append(allbuf, buffer[:lenRcv]...)
        byAfterDepackBuff = cliDataRcved(clientID, cliIP, cliAddr, lenRcv, allbuf, svr, asyncReceive)
        allbuf = byAfterDepackBuff
    }

    // fmt.Printf("%v end loop of client read, clientID:%v, ip:%v, addr:%v\n", ftag, clientID, cliIP, cliAddr)
}

// loop send for client send
func cliLoopSend(conn net.Conn, clientID uint64, cliIP string, cliAddr string,
    chMsgsToBeSend <-chan interface{}, cliSessn *clientSessnSt,
    svr *Ctcpsvr, requireSendedCb bool, asyncSended bool) {
    const ftag = "cliLoopSend()"

    // close count
    svr.wg.Add(1)
    defer svr.wg.Done()

    var (
        bysTobeSend []byte
        dumyBys     = make([]byte, 0)
        msg         interface{}
        ok          bool
        length      int
        err         error
    )

    for {
        bysTobeSend = dumyBys

        select {
        case <-svr.chExit:
            // server close
            //fmt.Printf("%v closing of client chExit-1, clientID:%v, ip:%v, addr:%v\n", ftag, clientID, cliIP, cliAddr)
            sEvnetMsg := fmt.Sprintf("%s exit because exit signal, clientID:%v, ip:%v, addr:%v",
                ftag, clientID, cliIP, cliAddr)
            svr.handler.OnEvent(sEvnetMsg)
            return

        case msg, ok = <-chMsgsToBeSend:
            if !ok {
                sEvnetMsg := fmt.Sprintf("%s exit because send channel closed, clientID:%v, ip:%v, addr:%v",
                    ftag, clientID, cliIP, cliAddr)
                svr.handler.OnEvent(sEvnetMsg)
                return
            }
        }

        bysTobeSend, err = svr.handler.Pack(clientID, cliIP, cliAddr, msg)
        if nil != err {
            svr.handler.OnCliError(clientID, cliIP, cliAddr, "pack failed", err)
            continue
        } else if nil == bysTobeSend {
            svr.handler.OnCliErrorStr(clientID, cliIP, cliAddr, "empty []byte to send")
            continue
        }

        length, err = conn.Write(bysTobeSend)
        if nil != err {
            select {
            case <-svr.chExit:
                // server close
                //fmt.Printf("%v closing of client chExit-3, clientID:%v, ip:%v, addr:%v\n", ftag, clientID, cliIP, cliAddr)
                return

            default:
            }

            if cliSessn.closed {
                break
            }

            errMsgTmp := fmt.Sprintf("write failed, msg=%v", msg)
            svr.handler.OnCliError(clientID, cliIP, cliAddr, errMsgTmp, err)
            //fmt.Printf("%s %s, IP:%v, addr:%v, len:%v, err:%v\n", ftag, errMsgTmp, cliIP, cliAddr, len, err)
            break
        }

        cliDataSended(clientID, cliIP, cliAddr, msg, bysTobeSend, length, svr, requireSendedCb, asyncSended)
    }

    // fmt.Printf("%v end loop of client send, clientID:%v, ip:%v, addr:%v\n", ftag, clientID, cliIP, cliAddr)
}

//
func closeCli(clientID uint64, reason string, svr *Ctcpsvr) {
    const ftag = "closeCli()"

    cliSessn, ok := svr.cliSns.getClientSession(clientID)
    if !ok {
        errMsgTmp := fmt.Sprintf("%v no such id, close reason=%v", ftag, reason)
        svr.handler.OnCliErrorStr(clientID, "", "", errMsgTmp)

        // fmt.Printf("%v no such id, client-id:%v failed, close reason:%v\n",
        // 	ftag, clientID, reason)
        return
    }

    if nil == cliSessn {
        return
    }

    closeCliConn(cliSessn.conn, clientID, cliSessn.ip, cliSessn.addr, cliSessn, svr)

    // fmt.Printf("%v close client-idy:%v, ip:%v, addr:%v, close reason:%v\n",
    // 	ftag, clientID, cliSessn.ip, cliSessn.addr, reason)
}

// close clients
func closeClients(clientIDs []uint64, reason string, svr *Ctcpsvr) {
    // const ftag = "closeClients()"

    if nil == clientIDs || 0 == len(clientIDs) {
        return
    }

    for _, cid := range clientIDs {
        closeCli(cid, reason, svr)
    }
}

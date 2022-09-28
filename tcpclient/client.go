package tcpclient

import (
    "fmt"
    "io"
    "net"
)

// connect to remote server
func connectToServer(ip string, port uint16, cli *CtcpCli) error {
    // const ftag = "connectToServer()"

    svrAddr := fmt.Sprintf("%s:%d", ip, port)
    conn, err := net.Dial("tcp", svrAddr)
    if nil != err {
        return err
    }

    cliNewConnection(conn, ip, port, cli)

    go cliLoopRead(conn, ip, port, cli, cli.cnf.AsyncReceive)
    go cliLoopSend(conn, ip, port, cli.chMsgsToBeSend, cli,
        cli.cnf.RequireSendedCb, cli.cnf.AsyncSended)

    return nil
}

func closeCliConn(conn net.Conn, ip string, port uint16, cli *CtcpCli) {
    conn.Close()

    // notify upper handler
    cliDisconnected(ip, port, cli)
}

// loop read for one client
func cliLoopRead(conn net.Conn, ip string, port uint16,
    cli *CtcpCli, asyncReceive bool) {
    const ftag = "cliLoopRead()"

    defer closeCliConn(conn, ip, port, cli)

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
            case <-cli.chExit:
                // client close
                return

            default:
            }

            if io.EOF == err {
                errMsgTmp := fmt.Sprintf("%v read from connect EOF, server IP:%s, port:%d", ftag, ip, port)
                cli.handler.OnErrorStr(errMsgTmp)
                break
            } else {
                errMsgTmp := fmt.Sprintf("%v read from connect failed, server IP:%s, port:%d", ftag, ip, port)
                cli.handler.OnError(errMsgTmp, err)
                break
            }
        }

        if 0 == lenRcv {
            continue
        }

        allbuf = append(allbuf, buffer[:lenRcv]...)
        byAfterDepackBuff = cliDataRcved(ip, port, lenRcv, allbuf, cli, asyncReceive)
        allbuf = byAfterDepackBuff
    }

    // fmt.Printf("%v end loop of client read, server IP:%v, port:%v\n", ftag, ip, port)
}

// loop send for client send
func cliLoopSend(conn net.Conn, ip string, port uint16,
    chMsgsToBeSend <-chan interface{}, cli *CtcpCli,
    requireSendedCb bool, asyncSended bool) {
    const ftag = "cliLoopSend()"

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
        case <-cli.chExit:
            // client close
            //fmt.Printf("%v closing of client chExit-1, ip:%s, port:%d\n", ftag, ip, port)
            cli.handler.OnEvent("cliLoopSend() exit because exit signal")
            return

        case msg, ok = <-chMsgsToBeSend:
            if !ok {
                cli.handler.OnEvent("cliLoopSend() exit because send channel closed")
                return
            }
        }

        bysTobeSend, err = cli.handler.Pack(msg)
        if nil != err {
            cli.handler.OnError("pack failed", err)
            continue
        } else if nil == bysTobeSend {
            cli.handler.OnErrorStr("empty []byte to send")
            continue
        }

        length, err = conn.Write(bysTobeSend)
        if nil != err {
            select {
            case <-cli.chExit:
                // client close
                //fmt.Printf("%v closing of client chExit-3, ip:%s, port:%d\n", ftag, ip, port)
                return

            default:
            }

            errMsgTmp := fmt.Sprintf("write failed, msg=%v", msg)
            cli.handler.OnError(errMsgTmp, err)
            continue
        }

        cliDataSended(ip, port, msg, bysTobeSend, length, cli, requireSendedCb, asyncSended)
    }

    //fmt.Printf("%v end loop of client send, ip:%s, port:%d\n", ftag, ip, port)
}

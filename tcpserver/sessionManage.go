package tcpserver

import (
    "net"
)

/*
functions for tcp server clients manage
*/

// add new connection into server's client session management
func svrNewConnection(conn net.Conn, clientID uint64, cliIP string, cliAddr string, svr *Ctcpsvr) *clientSessnSt {
    cliSessn := svr.cliSns.addNewConnection(conn, clientID, cliIP, cliAddr, svr.cnf.SendBuffsize)
    svr.handler.OnNewConnection(clientID, cliIP, cliAddr)
    return cliSessn
}

// report to server disconnect from one client
func svrDisconnect(clientID uint64, cliIP string, cliAddr string, svr *Ctcpsvr) {
    svr.cliSns.delConnect(clientID, cliIP, cliAddr)
    go svr.handler.OnDisconnected(clientID, cliIP, cliAddr)
}

// received client data
func cliDataRcved(clientID uint64, cliIP string, cliAddr string, length int, rawData []byte,
    svr *Ctcpsvr, asyncReceive bool) []byte {
    // if need to count incoming traffic, could be here
    //

    // self define message depack protocol
    byAfterDepackBuff, pPacks := svr.handler.Depack(clientID, cliIP, cliAddr, rawData)

    if nil != pPacks && 0 != len(pPacks) {
        if asyncReceive {
            go svr.handler.OnReceiveData(clientID, cliIP, cliAddr, pPacks)
        } else {
            svr.handler.OnReceiveData(clientID, cliIP, cliAddr, pPacks)
        }
    }

    return byAfterDepackBuff
}

// client data already sended
func cliDataSended(clientID uint64, cliIP string, cliAddr string, msg interface{}, bysSended []byte, length int, svr *Ctcpsvr,
    requireSendedCb bool, asyncSended bool) {
    // if need to count outgoing traffic, could be here
    //

    // report
    if requireSendedCb {
        if asyncSended {
            go svr.handler.OnSendedData(clientID, cliIP, cliAddr, msg, bysSended, length)
        } else {
            svr.handler.OnSendedData(clientID, cliIP, cliAddr, msg, bysSended, length)
        }
    }
}

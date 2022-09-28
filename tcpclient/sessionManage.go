package tcpclient

import "net"

func cliNewConnection(conn net.Conn, ip string, port uint16, cli *CtcpCli) {
    cli.addNewConnection(conn, ip, port)
    go cli.handler.OnNewConnection(ip, port)
}

func cliDisconnected(ip string, port uint16, cli *CtcpCli) {
    cli.disconnected()
    go cli.handler.OnDisconnected(ip, port)
}

// receive data
func cliDataRcved(ip string, port uint16, length int, rawData []byte,
    cli *CtcpCli, asyncReceive bool) []byte {
    // if need to count incoming traffic, could be here
    //

    // self define message depack protocol
    byAfterDepackBuff, pPacks := cli.handler.Depack(rawData)
    if nil != pPacks && 0 != len(pPacks) {
        if asyncReceive {
            go cli.handler.OnReceiveData(ip, port, pPacks)
        } else {
            cli.handler.OnReceiveData(ip, port, pPacks)
        }
    }

    return byAfterDepackBuff
}

// data already sened
func cliDataSended(ip string, port uint16, msg interface{}, bysSended []byte, length int, cli *CtcpCli,
    requireSendedCb bool, asyncSended bool) {
    // if need to count outgoing traffic, could be here
    //

    // report
    if requireSendedCb {
        if asyncSended {
            go cli.handler.OnSendedData(ip, port, msg, bysSended, length)
        } else {
            cli.handler.OnSendedData(ip, port, msg, bysSended, length)
        }
    }

}

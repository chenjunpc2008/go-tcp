package main

import (
    "fmt"
    "log"
    "sync/atomic"

    "github.com/chenjunpc2008/go-tcp/example/echoprotocol"
)

type appHandler struct {
}

// pack message into the []byte to be written
func (hdl *appHandler) Pack(clientID uint64, clientIP string, clientAddr string, msg interface{}) ([]byte, error) {
    const ftag = "appHandler.Pack()"

    // TODO: Do your message pack here -- msg to []byte sends to client

    var (
        ok   bool
        buff []byte
        sErr string
    )

    buff, ok, sErr = echoprotocol.Pack(msg, 0)
    if ok {
        return buff, nil
    }

    return buff, fmt.Errorf("echoprotocol.Pack() error %s", sErr)
}

// depack the message packages from read []byte
func (hdl *appHandler) Depack(clientID uint64, clientIP string, clientAddr string, rawData []byte) ([]byte, []interface{}) {
    const ftag = "appHandler.Depack()"

    // TODO: Do your message depack here -- []byte to msg receive from client

    var (
        dataRemain []byte
        pakgs      []interface{}
        ok         bool
        errMsg     string
    )

    dataRemain, pakgs, ok, errMsg = echoprotocol.Depack(rawData)
    if !ok {
        // depack have err
        log.Println(ftag, clientID, clientIP, clientAddr, errMsg)
    }

    return dataRemain, pakgs
}

// new connections event
func (hdl *appHandler) OnNewConnection(clientID uint64, clientIP string, clientAddr string) {
    const ftag = "appHandler.OnNewConnection()"

    log.Println(ftag, clientID, clientIP, clientAddr)
}

// disconnected event
func (hdl *appHandler) OnDisconnected(clientID uint64, clientIP string, clientAddr string) {
    const ftag = "appHandler.OnDisconnected()"

    log.Println(ftag, clientID, clientIP, clientAddr)
}

// receive data event
func (hdl *appHandler) OnReceiveData(clientID uint64, clientIP string, clientAddr string, pPacks []interface{}) {
    const ftag = "appHandler.OnReceiveData()"

    // add
    size := uint64(len(pPacks))
    atomic.AddUint64(&gTotalReceivedPkg, size)

    // log.Println(ftag, clientID, clientIP, clientAddr)

    // TODO: do your receive process here
    genEchoBack(clientID, clientIP, clientAddr, pPacks)
}

// data already sended event
func (hdl *appHandler) OnSendedData(clientID uint64, clientIP string, clientAddr string, msg interface{}, bysSended []byte, length int) {
    const ftag = "appHandler.OnSendedData()"

    // add
    atomic.AddUint64(&gTotalSendedPkg, 1)

    // log.Println(ftag, clientID, clientIP, clientAddr)
}

// event
func (hdl *appHandler) OnEvent(msg string) {
    const ftag = "appHandler.OnEvent()"

    log.Println(ftag, msg)
}

// error
func (hdl *appHandler) OnError(msg string, err error) {
    const ftag = "appHandler.OnError()"

    log.Println(ftag, msg, err)
}

// error
func (hdl *appHandler) OnCliError(clientID uint64, clientIP string, clientAddr string, msg string, err error) {
    const ftag = "appHandler.OnCliError()"

    log.Println(ftag, clientID, clientIP, clientAddr, msg, err)
}

// error
func (hdl *appHandler) OnCliErrorStr(clientID uint64, clientIP string, clientAddr string, msg string) {
    const ftag = "appHandler.OnCliErrorStr()"

    log.Println(ftag, clientID, clientIP, clientAddr, msg)
}

func genEchoBack(clientID uint64, clientIP string, clientAddr string, pPacks []interface{}) {
    const ftag = "genEchoBack()"

    for range pPacks {
        msg := prepareData(gDataSize)
        busy, err := gServer.SendToClient(clientID, msg)
        if nil != err {
            log.Println(ftag, clientID, clientIP, clientAddr, "SendToClient failed", err)
        } else if busy {
            log.Println(ftag, clientID, clientIP, clientAddr, "SendToClient failed because of busy")
        }
    }
}

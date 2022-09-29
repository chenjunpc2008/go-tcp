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
func (hdl *appHandler) Pack(msg interface{}) ([]byte, error) {
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
func (hdl *appHandler) Depack(rawData []byte) ([]byte, []interface{}) {
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
        log.Println(ftag, errMsg)
    }

    return dataRemain, pakgs
}

// new connections event
func (hdl *appHandler) OnNewConnection(serverIP string, serverPort uint16) {
    const ftag = "appHandler.OnNewConnection()"

    log.Println(ftag, serverIP, serverPort)
}

// disconnected event
func (hdl *appHandler) OnDisconnected(serverIP string, serverPort uint16) {
    const ftag = "appHandler.OnDisconnected()"

    log.Println(ftag, serverIP, serverPort)
}

// receive data event
func (hdl *appHandler) OnReceiveData(serverIP string, serverPort uint16, pPacks []interface{}) {
    const ftag = "appHandler.OnReceiveData()"

    // add
    size := uint64(len(pPacks))
    atomic.AddUint64(&gTotalReceivedPkg, size)

    // log.Println(ftag, serverIP, serverPort)

    // TODO: do your receive process here
}

// data already sended event
func (hdl *appHandler) OnSendedData(serverIP string, serverPort uint16, msg interface{}, bysSended []byte, length int) {
    const ftag = "appHandler.OnSendedData()"

    // add
    atomic.AddUint64(&gTotalSendedPkg, 1)

    // log.Println(ftag, serverIP, serverPort)
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
func (hdl *appHandler) OnErrorStr(msg string) {
    const ftag = "appHandler.OnErrorStr()"

    log.Println(ftag, msg)
}

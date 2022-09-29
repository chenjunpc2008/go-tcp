# go-tcp
Go TCP library: <https://github.com/chenjunpc2008/go-tcp>

## tcpclient
Go TCP client

## tcpserver
Go TCP server

# benchmark
Use ```example/pressure-server``` and ```example/pressure-client```, ```1000 concurrent clients```, ```4 KB payload``` get a result of ```100000 q/s```.

# Usage

## tcpserver
---
for example: ```example/pressure-server```

1. Create your own tcp protocol package pack/depack(or Marshal/Unmarshal) method for tcp transport
2. Create a struct for server event call back, and put your own tcp protocol package pack/depack method in Pack()/Depack()
    ```go
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
        // log.Println(ftag, serverIP, serverPort)
    }

    // data already sended event
    func (hdl *appHandler) OnSendedData(serverIP string, serverPort uint16, msg interface{}, bysSended []byte, length int) {
        const ftag = "appHandler.OnSendedData()"
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
    ```
3. Use the server and go
    ```go
    appHdl := &appHandler{}
    cnf := tcpserver.DefaultConfig()

    gServer = tcpserver.NewTCPSvr(appHdl, cnf)

    err = gServer.StartServer(uint16(gPort))
    if nil != err {
        log.Panicln("StartServer failed", err)
    }
    ```


## tcpclient
---
for example: ```example/pressure-client```

1. Create your own tcp protocol package pack/depack(or Marshal/Unmarshal) method for tcp transport
2. Create a struct for server event call back, and put your own tcp protocol package pack/depack method in Pack()/Depack()
    ```go
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
        // log.Println(ftag, serverIP, serverPort)

        // TODO: do your receive process here
    }

    // data already sended event
    func (hdl *appHandler) OnSendedData(serverIP string, serverPort uint16, msg interface{}, bysSended []byte, length int) {
        const ftag = "appHandler.OnSendedData()"
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
    ```
3. Use the client and go
    ```go
    appHdl := &appHandler{}
    cnf := tcpclient.DefaultConfig()

    client := tcpclient.New(appHdl, cnf)

    err = client.ConnectToServer(serverIP, serverPort)
    if nil != err {
        log.Panicln("ConnectToServer failed", err)
    }

    // send
    busy, err = client.SendToServer(msg)
    if nil != err {
        log.Println("SendToServer failed", err)
        break
    }

    if busy {
        log.Println("SendToServer failed because busy")
        break
    }
    ```

package tcpclient

// ProtocolIF interface to self define pack and depack
type ProtocolIF interface {
    // pack message into the []byte to be written
    Pack(msg interface{}) ([]byte, error)

    // depack the message packages from read []byte
    Depack(rawData []byte) ([]byte, []interface{})
}

// EventHandler client callback control handler
type EventHandler interface {
    // new connections event
    OnNewConnection(serverIP string, serverPort uint16)
    // disconnected event
    OnDisconnected(serverIP string, serverPort uint16)
    // receive data event
    OnReceiveData(serverIP string, serverPort uint16, pPacks []interface{})
    // data already sended event
    OnSendedData(serverIP string, serverPort uint16, msg interface{}, bysSended []byte, length int)
    // event
    OnEvent(msg string)
    // error
    OnError(msg string, err error)
    // error
    OnErrorStr(msg string)

    // data protocol
    ProtocolIF
}

package tcpserver

// ProtocolIF interface to self define pack and depack
type ProtocolIF interface {
    // pack message into the []byte to be written
    Pack(clientID uint64, cliIP string, cliAddr string, msg interface{}) ([]byte, error)

    // depack the message packages from read []byte
    Depack(clientID uint64, cliIP string, cliAddr string, rawData []byte) ([]byte, []interface{})
}

/*
EventHandler server callback control handler
*/
type EventHandler interface {
    // new connections event
    OnNewConnection(clientID uint64, clientIP string, clientAddr string)
    // disconnected event
    OnDisconnected(clientID uint64, clientIP string, clientAddr string)
    // receive data event
    OnReceiveData(clientID uint64, clientIP string, clientAddr string, pPacks []interface{})
    // data already sended event
    OnSendedData(clientID uint64, clientIP string, clientAddr string, msg interface{}, bysSended []byte, length int)
    // event
    OnEvent(msg string)
    // error
    OnError(msg string, err error)
    // error
    OnCliError(clientID uint64, clientIP string, clientAddr string, msg string, err error)
    // error
    OnCliErrorStr(clientID uint64, clientIP string, clientAddr string, msg string)

    // data protocol
    ProtocolIF
}

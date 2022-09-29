package echoprotocol

import (
    "encoding/binary"
)

/*
Ntohl32 net to host long 32bit
*/
func Ntohl32(byData []byte) (uint32, bool) {
    bSize := len(byData)
    if 0 == bSize {
        return 0, true
    }

    if 4 < bSize {
        return 0, false
    }

    var (
        intRes uint32
    )

    intRes = binary.BigEndian.Uint32(byData)

    return intRes, true
}

/*
Ntohll64 net to host long long 64bit
*/
func Ntohll64(byData []byte) (uint64, bool) {
    bSize := len(byData)
    if 0 == bSize {
        return 0, true
    }

    if 8 < bSize {
        return 0, false
    }

    var (
        intRes uint64
    )

    intRes = binary.BigEndian.Uint64(byData)

    return intRes, true
}

/*
Htonl32 host to net long 32bit
*/
func Htonl32(iNum uint32) []byte {
    var (
        byNet []byte = []byte{byte('\x00'), byte('\x00'), byte('\x00'), byte('\x00')}
    )

    binary.BigEndian.PutUint32(byNet, iNum)
    return byNet
}

/*
Htonll64 host to net long long 64bit
*/
func Htonll64(iNum uint64) []byte {
    var (
        byNet []byte = []byte{byte('\x00'), byte('\x00'), byte('\x00'), byte('\x00'),
            byte('\x00'), byte('\x00'), byte('\x00'), byte('\x00')}
    )

    binary.BigEndian.PutUint64(byNet, iNum)
    return byNet
}

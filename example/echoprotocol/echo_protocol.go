package echoprotocol

import (
    "fmt"
    "time"
)

/*
// when in C++
// set Byte Alignment to 1
#pragma pack(push,1)

struct ECHO_MSG
{
    // BeginString 4byte 0-3
    char beginstring[4];
    // Timestamp 4byte 4-7
    uint32_t timestamp;
    // DataLen RawData length 8-11
    uint32_t datalen;
};

// set Byte Alignment back to default(normally 4)
#pragma pack(pop)
*/

/*
EchoMsg message sends between client and server

it will be like
  0             11    DataLen
  | --  HEAD -- | --  RawData -- |
*/
type EchoMsg struct {
    // BeginString 4byte 0-3
    BeginString [4]byte
    // Timestamp 4byte 4-7
    Timestamp uint32
    // DataLen RawData length 8-11
    DataLen uint32
    // RawData
    RawData []byte
}

const (
    // NetPackHeadSize head size, fixed 12 bytes
    NetPackHeadSize = 12
    // NetPackBeginstrSize head begin string size
    NetPackBeginstrSize = 4
    // NetPackTimestampSize head timestamp size
    NetPackTimestampSize = 4
    // NetPackDataLenSize head DataLen size
    NetPackDataLenSize = 4
)

/*
Depack depack

@param rawData []byte : origin data
@return []byte : data left after depack
@return *list.List[]NetPacket : depacked package lists
@return bool : is depack success
		 true -- success
		 false -- have error, check errMsg
@return string : err msgs when depack
*/
func Depack(rawData []byte) (rawDataLocal []byte, outPacks []interface{}, bOk bool, strErrMsg string) {
    const ftag = "echoprotocol.Depack()"
    rawDataLocal = rawData

    if NetPackHeadSize > len(rawData) {
        // not enough length for depack, keep buff
        return rawData, nil, true, ""
    }

    // is matched beginstring
    bMatchBeginstr := false

    // pos
    var (
        uiBufferSize uint64
        i            uint64
        uiCurrentPos uint64
        byTemp       []byte
        strDiscard   string
    )

    uiCurrentPos = 0
    i = 0

    itcpPackBeginstr := [...]byte{byte('N'), byte('E'), byte('T'), byte(':')}

    outPacks = make([]interface{}, 0)

    strErrMsg = fmt.Sprintf("%v ", ftag)
    bOk = true

    for {
        uiBufferSize = uint64(len(rawDataLocal))
        if NetPackHeadSize > uiBufferSize {
            return
        }

        bMatchBeginstr = true

        /*
        	BeginString
        */
        for i = 0; i < NetPackBeginstrSize && bMatchBeginstr; i++ {
            if itcpPackBeginstr[i] != rawDataLocal[i] {
                bMatchBeginstr = false
                break
            }
        }

        if !bMatchBeginstr {
            // not head from start, neet to clean
            rawDataLocal, strDiscard = cleanNotHead(rawDataLocal[1:])
            strErrMsg += "not head"
            strErrMsg += strDiscard
            bOk = false
            // depack from start
            continue
        }

        //
        uiCurrentPos = NetPackBeginstrSize

        // Timestamp
        byTemp = rawDataLocal[uiCurrentPos : uiCurrentPos+NetPackTimestampSize]
        iTimestamp, bRes := Ntohl32(byTemp)
        if !bRes {
            strErrMsg += fmt.Sprintf("error when get Timestamp, %v\n", byTemp)
            // have error, clean
            rawDataLocal, strDiscard = cleanNotHead(rawDataLocal[1:])
            strErrMsg += strDiscard
            bOk = false
            continue
        }

        onePack := new(EchoMsg)
        onePack.Timestamp = iTimestamp

        //
        uiCurrentPos += NetPackTimestampSize

        // data len
        byTemp = rawDataLocal[uiCurrentPos : uiCurrentPos+NetPackDataLenSize]
        iDataLen, bRes := Ntohl32(byTemp)
        if !bRes {
            strErrMsg += fmt.Sprintf("error when get data len, %v\n", byTemp)
            // have error, clean
            rawDataLocal, strDiscard = cleanNotHead(rawDataLocal[1:])
            strErrMsg += strDiscard
            bOk = false
            continue
        }

        if uint64(NetPackHeadSize+iDataLen) > uiBufferSize {
            // not enough length for depack, keep buff
            return
        }
        onePack.DataLen = iDataLen
        uiCurrentPos += NetPackDataLenSize

        // hard copy
        onePack.RawData = make([]byte, iDataLen, iDataLen*2+1)
        copy(onePack.RawData, rawDataLocal[uiCurrentPos:uiCurrentPos+uint64(iDataLen)])

        outPacks = append(outPacks, onePack)

        // 清除已接受的数据包
        if uiCurrentPos+uint64(iDataLen) == uiBufferSize {
            rawDataLocal = make([]byte, 0)
        } else {
            rawDataLocal = rawDataLocal[uiCurrentPos+uint64(iDataLen):]
        }
    }
}

/*
cleanNotHead clear datas not head

@param rawData []byte : raw datas
@return []byte : datas left after clean
*/
func cleanNotHead(rawData []byte) ([]byte, string) {
    for i := 0; i < len(rawData); i++ {
        // look for byte('N') begin
        if byte('N') == rawData[i] {
            // return data begins from byte('N')
            strDiscard := fmt.Sprintf("discard:%v\n", rawData[0:i])
            return rawData[i:], strDiscard
        }
    }

    // return empty
    strDiscard := fmt.Sprintf("discard:%v\n", rawData[:])
    return make([]byte, 0), strDiscard
}

/*
Pack pack

@param msg string : msg content
@param timestamp uint32 : timestamp，when not 0, use incoming timestamp
@return []byte : []byte after pack
@return bool : is pack success
        true -- success
        false -- have error, check errMsg
@return string : err msgs when pack
*/
func Pack(msg interface{}, timestamp uint32) ([]byte, bool, string) {

    if nil == msg {
        return make([]byte, 0), true, ""
    }

    var (
        sData string
        ok    bool
        itime uint32
    )

    sData, ok = msg.(string)
    if !ok {
        return make([]byte, 0), false, "not string type"
    }

    // msg
    msgData := []byte(sData)
    lMsgLen := uint32(len(msgData))

    // BeginString 4byte 0-3
    itcpPackBeginstr := []byte{byte('N'), byte('E'), byte('T'), byte(':')}

    totalSize := (NetPackHeadSize+lMsgLen)*2 + 1
    sendBuffer := make([]byte, 0, totalSize)

    sendBuffer = append(sendBuffer, itcpPackBeginstr...)

    // Timestamp 4byte 4-7
    if 0 == timestamp {
        itime = uint32(time.Now().Unix())
    } else {
        itime = timestamp
    }
    byteBuff := Htonl32(itime)
    sendBuffer = append(sendBuffer, byteBuff...)

    // DataLen RawData length 8-11
    byteBuff = Htonl32(lMsgLen)
    sendBuffer = append(sendBuffer, byteBuff...)

    // RawData
    sendBuffer = append(sendBuffer, msgData...)

    return sendBuffer, true, ""
}

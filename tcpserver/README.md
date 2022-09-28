# cchantcpserver
## tcp server frame
* notice server and client events via event callback handler.
* could define your own package pack/unpack protocol.
* each client with have it's own read and write goroutine, write messages will store in the send queue(use channel) before acctually write into system tcp send buffer.
* server will handler client connection session, upper layer will only have to process business session layer's management.

## tcp server架构
* 通过事件回调handler向外反馈server和client事件
* 自定义消息的拆解包协议
* 每个客户端有独立的读写go协程，待发送消息会先存入发送队列（使用channel），再写入系统tcp发送缓冲区
* 服务端将管理client连接，上层应用仅需按协议处理session层的内容

```
accepting connection:
+------------+    +------------+    +----------------+
|            |    |            |    |                |
| tcp server |--->| accept     |--->| add to client  |
|            |    | connection |    |   connections  |
|            |    |            |    |   management   |
+------------+    +------------+    +----------------+
                                            |
                                            |
+------------+    +-------------------+     |
|            |    |                   |     |
| your own   |<---| OnNewConnection() | <---+
|  process   |    | callback          |  
|            |    |                   | 
+------------+    +-------------------+

in client life time:
read tcp datas
+------------+    +-----------------------+    +-----------------+
| read       |    | unpack packet payload |    | OnReceiveData() |
| connection |--->| use your own protocol |--->| callback        |
+------------+    +-----------------------+    +-----------------+

send tcp datas
+------------+    +-----------+    +------------+
| your own   |    | server    |    | client's   |
| app send   |--->| send func |--->| send queue |
+------------+    +-----------+    +------------+
                                            |
                                            |
   +--------------+     +--------------+    |
   | client's send |    | pack packet  |    |
   | loop write    |<---| use your own |<---+
   | []byte to sys |    |  protocol    | 
   | tcp send buff |    +--------------+
   +---------------+                        
```

//@title tcpClient.go
//@Description 负责向SIP服务器端请求流信息（流地址和流类型）

package tcpClient

import (
	"encoding/json"
	"fmt"
	"net"
)

type TCPResponseStatus struct {
	ResponseStatus string `json:"responseStatus"`
}

var ch = make(chan int)

func cConnHandler(conn net.Conn, streaminfo []byte) (string, error) {
	// buf := make([]byte, 1024)
	buf := make([]byte, 65536)

	for {
		fmt.Println("给服务器发送数据：", string(streaminfo))
		//向服务器发送请求数据
		_, err := conn.Write(streaminfo)
		if err != nil {
			fmt.Println("给SIP服务器发送数据失败", err)
			return "", err
		}
		fmt.Println("发送成功")
		//服务器端返回的数据写入空buf
		cnt, err := conn.Read(buf)
		if err != nil {
			fmt.Printf("读取SIP回复的状态码失败%s\n", err)
			return "", err
		}
		var ResponseStatusJson TCPResponseStatus
		ResponseStatusJson.ResponseStatus = "true"
		responseJson, _ := json.Marshal(ResponseStatusJson)
		_, err = conn.Write(responseJson)
		if err != nil {
			fmt.Println("给SIP服务器发送responseStatus失败", err)
			return "", err
		}
		return string(buf[0:cnt]), nil
	}
}

//客户端套接字
func ClientSocket(streaminfo []byte) (string, error) {
	tcpAddr, _ := net.ResolveTCPAddr("tcp", "10.18.104.201:8801")
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return "", err
	}
	defer conn.Close()
	rst, err := cConnHandler(conn, streaminfo)
	if err != nil {
		return "", err
	}
	return rst, nil
}

package networking

import (
	config "PROJECT-GROUP-10/config"
	"net"
	"strings"
)

var connections [config.E]net.UDPConn

var localIP string

func getLocalIp() (string, error) {
	if localIP == "" {
		conn, err := net.DialTCP("tcp4", nil, &net.TCPAddr{IP: []byte{8, 8, 8, 8}, Port: 53})
		if err != nil {
			return "", err
		}
		defer conn.Close()
		localIP = strings.Split(conn.LocalAddr().String(), ":")[0]
	}
	return localIP, nil
}

func heartBeatTransmitter() error {

}

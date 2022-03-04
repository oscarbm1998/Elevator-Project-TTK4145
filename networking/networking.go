package networking

import (
	config "PROJECT-GROUP-10/config"
	"net"
	"strings"
	"time"
)

var connections [config.NUMBER_OF_ELEVATORS]net.UDPConn

type heartBeatMessage struct {
	my_time       time.Time
	my_direction  int
	current_floor int
}

var myHeartBeatMessage heartBeatMessage
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
	timer := time.NewTimer(config.HEARTBEAT_TIME)
	for {
		select {
		case <-timer.C:
			myHeartBeatMessage.my_time = time.Now()

		}
	}
}

package networking

import (
	config "PROJECT-GROUP-10/config"
	"fmt"
	"net"
	"strings"
	"time"
)

var HB_con_Out [config.NUMBER_OF_ELEVATORS - 1]*net.UDPConn
var HB_con_In [config.NUMBER_OF_ELEVATORS - 1]*net.UDPConn
var elevatorsIPs [config.NUMBER_OF_ELEVATORS - 1]string

var myFloor int
var myDirection int
var myStatus int

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

func heartBeatTransmitter() (err error) {
	timer := time.NewTimer(config.HEARTBEAT_TIME)
	var msg string
	for {
		select {
		case <-timer.C:
			timer.Reset(config.HEARTBEAT_TIME)
			msg = "S_" + time.Now().String() + "_"
			msg = msg + string(config.ELEVATOR_ID) + "_"
			msg = msg + string(myDirection) + "_"
			msg = msg + string(myFloor) + "_"
			msg = msg + string(myStatus) + "_E"
			for ID := 0; ID < config.NUMBER_OF_ELEVATORS-1; ID++ {
				_, err := HB_con_Out[ID].Write([]byte(msg))
				return err
			}
			msg = ""
		}
	}
}

func resolveHBConn() (err error) {
	for i := 0; i < config.NUMBER_OF_ELEVATORS; i++ {
		//Outgoing
		network, err := net.ResolveUDPAddr("udp", elevatorsIPs[i]+string(config.HEARTBEAT_TRANS_PORT))
		printError("resolveHBconn setup error: ", err)
		con, err := net.DialUDP("udp", nil, network)
		HB_con_Out[i] = con
		printError("resolveHBconn dial error: ", err)

		//Incomming
		networkIn, err := net.ResolveUDPAddr("udp", elevatorsIPs[i]+string(config.HEARTBEAT_REC_PORT))
		printError("resolveHBconn setup error: ", err)
		conIn, err := net.DialUDP("udp", nil, networkIn)
		HB_con_In[i] = conIn
		printError("resolveHBconn dial error: ", err)
	}
	return err
}

func printError(str string, err error) {
	if err != nil {
		fmt.Print(str)
		fmt.Println(err)
	}
}

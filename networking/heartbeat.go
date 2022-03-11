package networking

import (
	config "PROJECT-GROUP-10/config"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

var HB_con_Out [config.NUMBER_OF_ELEVATORS - 1]*net.UDPConn
var elevatorsIPs [config.NUMBER_OF_ELEVATORS - 1]string

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
	var msg, date, clock string
	var ID int = config.ELEVATOR_ID

	for {
		select {
		case <-timer.C:
			timer.Reset(config.HEARTBEAT_TIME)

			//Sampling date and time, and making it nice european style
			year, month, day := time.Now().Date()
			date = strconv.Itoa(day) + "/" + month.String() + "/" + strconv.Itoa(year)
			hour, minute, second := time.Now().Clock()
			clock = strconv.Itoa(hour) + ":" + strconv.Itoa(minute) + ":" + strconv.Itoa(second)
			msg = date + " " + clock + "_"

			//Adding elevator data
			msg = msg + strconv.Itoa(ID) + "_"
			msg = msg + strconv.Itoa(elevator_nodes[ID-1].direction) + "_"
			msg = msg + strconv.Itoa(elevator_nodes[ID-1].floor) + "_"
			msg = msg + strconv.Itoa(elevator_nodes[ID-1].status)

			//Sending to all nodes
			for i := 0; i < config.NUMBER_OF_ELEVATORS-1; i++ {
				_, err := HB_con_Out[i].Write([]byte(msg))
				return err
			}
			msg = ""
		}
	}
}

func resolveHBConn() (err error) {
	for i := 0; i < config.NUMBER_OF_ELEVATORS; i++ {
		//Outgoing
		network, err := net.ResolveUDPAddr("udp", string(config.HEARTBEAT_TRANS_PORT))
		printError("resolveHBconn setup error: ", err)
		con, err := net.DialUDP("udp", nil, network)
		HB_con_Out[i] = con
		printError("resolveHBconn dial error: ", err)

	}
	return err
}

func heartBeathandler() {
	//Initiate connections
	err := resolveHBConn()
	if err != nil {
		panic(err)
	}

	//Initiate the UDP heartbeat listener
	ch_heartbeatmsg := make(chan string)

	go heartbeat_UDPListener(ch_heartbeatmsg)

	//Initiate heartbeat timers as go routines for each elevator
	var ch_foundDead chan int
	var ch_timerStop, ch_timerReset [config.NUMBER_OF_ELEVATORS - 1]chan bool
	var ID int

	for i := 0; i < config.NUMBER_OF_ELEVATORS-1; i++ {
		go heartbeatTimer(i+1, ch_foundDead, ch_timerReset[i], ch_timerStop[i])
	}
	//Kill its own timer
	ch_timerStop[config.ELEVATOR_ID-1] <- true

	for {
		select {
		case <-ch_heartbeatmsg:
			//Parsing the received heartbeat message
			data := strings.Split(<-ch_heartbeatmsg, "_")
			ID, _ = strconv.Atoi(data[1])
			Elevator_nodes[ID-1].last_seen = data[0]
			Elevator_nodes[ID-1].ID = ID
			Elevator_nodes[ID-1].direction, _ = strconv.Atoi(data[2])
			Elevator_nodes[ID-1].floor, _ = strconv.Atoi(data[3])
			Elevator_nodes[ID-1].status, _ = strconv.Atoi(data[4])

			//Reset the appropriate timer
			ch_timerReset[ID-1] <- true

		case <-ch_foundDead:
			//Timer has run out,
			fmt.Printf("found %d dead", <-ch_foundDead)
			Elevator_nodes[<-ch_foundDead-1].status = 404
		}
	}
}

//Timer, waiting for something to timeout. Run as a go routine, accessed through channels
func heartbeatTimer(ID int, ch_foundDead chan int, ch_timerReset, ch_timer_stop chan bool) {
	timer := time.NewTimer(config.HEARTBEAT_TIME)
	for {
		select {
		case <-timer.C:
			ch_foundDead <- ID
		case <-ch_timerReset:
			timer.Reset(config.HEARTBEAT_TIME)
		case <-ch_timer_stop:
			timer.Stop()
		}
	}
}

func heartbeat_UDPListener(ch_heartbeatmsg chan<- string) error {
	fmt.Println("Heartbead UDP listener thread starting")
	buf := make([]byte, 1024)
	var msg string
	adr, _ := net.ResolveUDPAddr("udp", strconv.Itoa(config.HEARTBEAT_REC_PORT))
	con, _ := net.ListenUDP("udp", adr)
	for {
		n, _, err := con.ReadFromUDP(buf)
		msg = string(buf[0:n])
		ch_heartbeatmsg <- msg
		return err
	}
}

func printError(str string, err error) {
	if err != nil {
		fmt.Print(str)
		fmt.Println(err)
	}
}

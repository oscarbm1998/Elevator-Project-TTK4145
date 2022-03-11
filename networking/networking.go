package networking

import (
	config "PROJECT-GROUP-10/config"
	"fmt"
	"net"
	"strconv"
	"strings"
)

type elevator_node struct {
	last_seen string
	ID        int
	direction int
	floor     int
	status    int
}

var elevator_nodes [config.NUMBER_OF_ELEVATORS]elevator_node
var command_cons, readback_cons [config.NUMBER_OF_ELEVATORS - 1]*net.UDPConn
var readback_con *net.UDPConn

func networking_main() {
	//Initialize heartbeat
	go heartBeatTransmitter()
	go heartBeathandler()

	//initiate command transmit connections
	for i := 0; i < config.NUMBER_OF_ELEVATORS-1; i++ {
		if i != config.ELEVATOR_ID+1 {
			network, err := net.ResolveUDPAddr("udp", ":"+strconv.Itoa(config.COMMAND_SEND_PORT+i))
			printError("Command transmit network resolve error: ", err)
			conn, err := net.DialUDP("udp", nil, network)
			printError("Command transmit network dial error: ", err)
			command_cons[i] = conn
			//Readback cons
			network, err = net.ResolveUDPAddr("udp", ":"+strconv.Itoa(config.COMMAND_READBACK_PORT+i))
			printError("Command transmit network resolve error: ", err)
			conn, err = net.DialUDP("udp", nil, network)
			printError("Command transmit network dial error: ", err)
			readback_cons[i] = conn
		}
	}
	//Initiate command readback port connection
	adr, _ := net.ResolveUDPAddr("udp", ":"+strconv.Itoa(config.COMMAND_READBACK_PORT+config.ELEVATOR_ID-1))
	readback_con, _ = net.ListenUDP("udp", adr)

	//Initiate command listener
	ch_command := make(chan string)

	go command_listener(ch_command)
	//Listen for commands
	for {
		select {
		case <-ch_command:
			fmt.Println("Networking: command received")

		}
	}
}

func send_command(ID, floor, direction int) (success bool) {
	var attempts int = 1
	var cmd string
	if ID == config.ELEVATOR_ID {
		panic("Networking: I do not need networking to command myself")
	}

	cmd = strconv.Itoa(ID) + "_" + strconv.Itoa(floor) + "_" + strconv.Itoa(direction)
	//Send command
	_, err := command_cons[ID-1].Write([]byte(cmd))
	printError("Networking: Error sending command: ", err)
	if err == nil {
		fmt.Println("Networking: command sent to elevator " + strconv.Itoa(ID))
	}
	//Wait for readback
	buf := make([]byte, 1024)
	for {
		n, _, err := readback_con.ReadFromUDP(buf)
		msg := string(buf[0:n])

		//Sending again if the readback is wrong
		if msg != cmd {
			fmt.Println("Networking: bad readback, sending again")
			_, err = command_cons[ID-1].Write([]byte(cmd))
			printError("Networking: Error sending command: ", err)
			attempts++
		} else if msg == "CMD_REJECT" { //Command rejected
			fmt.Printf("Network: elevator rejected the command")
			success = false
			break
		} else {
			fmt.Println("Networking: readback OK")
			_, err = command_cons[ID-1].Write([]byte("CMD_OK"))
			success = true
			break
		}

		if attempts > 2 {
			fmt.Println("Networking: too many command readback attemps")
			success = false
			break
		}
	}

	if err != nil {
		success = false
	}

	return success
}

func command_listener(ch_netcommand chan string) {
	adr, _ := net.ResolveUDPAddr("udp", ":"+strconv.Itoa(config.COMMAND_REC_PORT+config.ELEVATOR_ID-1))
	con, _ := net.ListenUDP("udp", adr)
	buf := make([]byte, 1024)
	for {
		//Listen for incomming commands on command reception port
		n, _, err := con.ReadFromUDP(buf)
		printError("Networking: error from command listener: ", err)
		msg := string(buf[0:n])
		data := strings.Split(msg, "_")
		ID, _ := strconv.Atoi(data[0])
		floor, _ := strconv.Atoi(data[1])
		direction, _ := strconv.Atoi(data[2])

		if reject_command(floor, direction) { //Check if i can perfrom the task
			_, err = readback_cons[ID-1].Write([]byte("CMD_REJECT"))
		} else { //Accept the command by reading it back
			_, err = readback_cons[ID-1].Write([]byte(msg))
			//Wait for OK
			n, _, err = con.ReadFromUDP(buf)
			msg = string(buf[0:n])
			if msg == "CMD_OK" {
				ch_netcommand <- msg
			}
		}
	}
}

func reject_command(direction, floor int) (reject bool) {
	if elevator_nodes[config.ELEVATOR_ID-1].status == 0 || floor < 0 || floor > config.NUMBER_OF_FLOORS {
		return true
	} else {
		return false
	}
}

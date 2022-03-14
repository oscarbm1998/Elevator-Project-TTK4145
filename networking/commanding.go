package networking

import (
	config "PROJECT-GROUP-10/config"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

func send_command(ID, floor, direction int) (success bool) {
	//(NB!!) This function does not check if the node is alive before attempting transmission.
	var attempts int = 1
	var cmd, rbc, broadcast string
	if ID == config.ELEVATOR_ID {
		panic("Networking: I do not need networking to command myself")
	}

	//Generate command
	//Format: ToElevatorID_ToFloor_InDirection_FromElevatorID
	cmd = strconv.Itoa(ID) + "_" + strconv.Itoa(floor) + "_" + strconv.Itoa(direction) + "_" + strconv.Itoa(config.ELEVATOR_ID)
	rbc = strconv.Itoa(config.ELEVATOR_ID) + "_" + strconv.Itoa(floor) + "_" + strconv.Itoa(direction) + "_" + strconv.Itoa(ID)

	//Initiate command broadcast connection
	broadcast = "255.255.255.255:" + strconv.Itoa(config.COMMAND_PORT)
	network, _ := net.ResolveUDPAddr("udp", broadcast)
	cmd_con, _ := net.DialUDP("udp", nil, network)

	//Initiate readback connection and timer
	ch_rbc_msg := make(chan string)
	ch_rbc_close := make(chan bool)
	go command_readback_listener(ch_rbc_msg, ch_rbc_close)
	//Send command
	_, err := cmd_con.Write([]byte(cmd))
	printError("Networking: Error sending command: ", err)

	//Starting a timer for timeout
	timOut := time.Second * 3
	timer := time.NewTimer(timOut)
	for {
		select {
		case msg := <-ch_rbc_msg:
			data := strings.Split(msg, "_")
			rbc_id, _ := strconv.Atoi(data[0])
			//Sending again if the readback is wrong
			if rbc_id == config.ELEVATOR_ID {
				timer.Reset(timOut)
				if msg == rbc {
					fmt.Println("Networking: readback OK")
					cmd_con.Write([]byte(strconv.Itoa(ID) + "_CMD_OK"))
					success = true
					return
				} else if rbc == strconv.Itoa(config.ELEVATOR_ID)+"_CMD_REJECT" { //Command rejected
					fmt.Printf("Network: elevator rejected the command")
					success = false
					return
				} else {
					fmt.Println("Networking: bad readback, sending command again")
					_, err = cmd_con.Write([]byte(cmd))
					printError("Networking: Error sending command: ", err)
					attempts++
				}
				if attempts > 3 {
					fmt.Println("Networking: too many command readback attemps")
					success = false
					return
				}
			}
		case <-timer.C:
			fmt.Println("Networking: sending command timed out, no readback")
			timer.Stop()
			success = false
			return
		}
		break
	}

	//Stopping readback listener and returning the results
	ch_rbc_close <- true
	return success
}

func command_readback_listener(ch_msg chan string, ch_close chan bool) {
	network, _ := net.ResolveUDPAddr("udp", ":"+strconv.Itoa(config.COMMAND_RBC_PORT))
	con, _ := net.ListenUDP("udp", network)
	buf := make([]byte, 1024)
	for {
		select {
		case <-ch_close:
			con.Close()
			break
		default:
			con.SetReadDeadline(time.Now().Add(3 * time.Second))
			n, _, err := con.ReadFromUDP(buf)
			if err != nil {
				if e, ok := err.(net.Error); !ok || e.Timeout() {
					printError("Networking: command readback net error: ", err)
				} else {
					fmt.Println("Networking: Getting nothing on readback channel, so quitting")
					con.Close()
					break
				}
				break
			}
			msg := string(buf[0:n])
			ch_msg <- msg
		}
		defer con.Close()
	}
}

/*
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
			fmt.Println("Network: incomming command rejected")
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
*/

func reject_command(direction, floor int) (reject bool) {
	if Elevator_nodes[config.ELEVATOR_ID-1].Status == 0 || floor < 0 || floor > config.NUMBER_OF_FLOORS {
		return true
	} else {
		return false
	}
}

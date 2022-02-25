package singleElevator

const queue_length int = 7

type elevator_Info struct {
	floor     int //where the elevator should go
	direction int //1 up -1 down
}

type elevator_event struct {
	floor     int //the qeue of where the elevators should go
	direction int //1 up -1 down
}

var elevator elevator_Info
var queue [queue_length]elevator_event

func queue_AddEvent(new_event elevator_event) {
	var overlap bool = false
	for i := 0; i < queue_length; i++ {
		if queue[i] == new_event {
			overlap = true
		}
	}
	if !overlap {
		for i := 0; i < queue_length; i++ {
			switch queue.direction {
			case 1:
				if queue[i].direction == new_event.direction {
					if new_event.floor < queue[i].floor && new_event.floor > queue[i+1].floor {
						queue = add_event_between(queue, new_event, i)
					}
				}
			case -1:
			}

		}
	}
}

func shift_Qeue() {

}

func pass_Qeue() {

}
func array_leftshift(in []int, length int) (out []int) {
	for i := 0; i < length; i++ {
		out[i] = in[i+1]
	}
	return out
}

func add_event_between(in [queue_length]elevator_event, new elevator_event, position int) (out [queue_length]elevator_event) {
	for i := 0; i < position-1; i++ {
		out[i] = in[i]
	}
	out[position] = new
	for i := position + 1; i < queue_length-1; i++ {
		out[i] = in[i+1]
	}
}

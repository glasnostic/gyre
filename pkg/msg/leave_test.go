package msg

import (
	zmq "github.com/armen/go-zmq"

	"testing"
)

func TestLeave(t *testing.T) {
	context, err := zmq.NewContext()
	if err != nil {
		t.Fatal(err)
	}

	// Output
	output, err := context.Socket(zmq.Dealer)
	if err != nil {
		t.Fatal(err)
	}
	address := []byte("Shout")
	output.SetIdentitiy(address)
	err = output.Bind("inproc://selftest")
	if err != nil {
		t.Fatal(err)
	}

	// Input
	input, err := context.Socket(zmq.Router)
	if err != nil {
		t.Fatal(err)
	}
	err = input.Connect("inproc://selftest")
	if err != nil {
		t.Fatal(err)
	}
	// Create a Leave message and send it through the wire
	leave := NewLeave()
	leave.SetSequence(123)
	leave.Group = "Life is short but Now lasts for ever"
	leave.Status = 123

	err = leave.Send(output)
	if err != nil {
		t.Fatal(err)
	}

	transit, err := Recv(input)
	if err != nil {
		t.Fatal(err)
	}

	msg := transit.(*Leave)
	if msg.Sequence() != 123 {
		t.Fatalf("expected %d, got %d", 123, msg.Sequence())
	}
	if msg.Group != "Life is short but Now lasts for ever" {
		t.Fatalf("expected %s, got %s", "Life is short but Now lasts for ever", msg.Group)
	}
	if msg.Status != 123 {
		t.Fatalf("expected %d, got %d", 123, msg.Status)
	}

	err = msg.Send(input)
	if err != nil {
		t.Fatal(err)
	}
	transit, err = Recv(output)
	if err != nil {
		t.Fatal(err)
	}
	if string(address) != string(msg.Address()) {
		t.Fatalf("expected %v, got %v", address, msg.Address())
	}
}
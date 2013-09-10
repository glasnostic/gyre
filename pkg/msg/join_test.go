package msg

import (
	zmq "github.com/armen/go-zmq"

	"testing"
)

func TestJoin(t *testing.T) {
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
	// Create a Join message and send it through the wire
	join := NewJoin()
	join.SetSequence(123)
	join.Group = "Life is short but Now lasts for ever"
	join.Status = 123

	err = join.Send(output)
	if err != nil {
		t.Fatal(err)
	}

	transit, err := Recv(input)
	if err != nil {
		t.Fatal(err)
	}

	msg := transit.(*Join)
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
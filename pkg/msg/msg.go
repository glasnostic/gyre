package msg

import (
	zmq "github.com/armen/go-zmq"

	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
)

const (
	Signature uint16 = 0xAAA0 | 1
	StringMax        = 255
	Version          = 1
)

type Transit interface {
	Marshal() ([]byte, error)
	Unmarshal([][]byte) error
	String() string
	Send(*zmq.Socket) error
	SetAddress([]byte)
	Address() []byte
	SetSequence(uint16)
	Sequence() uint16
}

// Recv receives marshaled data from 0mq socket
func Recv(socket *zmq.Socket) (t Transit, err error) {
	// Read valid message frame from socket; we loop over any
	// garbage data we might receive from badly-connected peers
	for {
		// Read all frames
		frames, err := socket.Recv()
		if err != nil {
			return nil, err
		}
		t, err := RecvRaw(frames, socket.GetType())
		if err != nil {
			continue
		}
		return t, err
	}
}

// RecvRaw receives marshaled data from raw frames
func RecvRaw(frames [][]byte, sType zmq.SocketType) (t Transit, err error) {
	var (
		buffer  *bytes.Buffer
		address []byte
	)

	// If we're reading from a ROUTER socket, get address
	if sType == zmq.Router {
		if len(frames) <= 1 {
			return nil, errors.New("malformed message")
		}
		address = frames[0]
		frames = frames[1:]
	}
	// Check the signature
	var signature uint16
	buffer = bytes.NewBuffer(frames[0])
	binary.Read(buffer, binary.BigEndian, &signature)
	if signature != Signature {
		// Invalid signature
		return nil, errors.New("malformed message")
	}

	// Get message id and parse per message type
	var id uint8
	binary.Read(buffer, binary.BigEndian, &id)

	switch id {
	case HelloId:
		t = NewHello()
	case WhisperId:
		t = NewWhisper()
	case ShoutId:
		t = NewShout()
	case JoinId:
		t = NewJoin()
	case LeaveId:
		t = NewLeave()
	case PingId:
		t = NewPing()
	case PingOkId:
		t = NewPingOk()
	}
	t.SetAddress(address)
	err = t.Unmarshal(frames)

	return t, err
}

// Clone clones a message
func Clone(t Transit) Transit {
	switch msg := t.(type) {
	case *Hello:
		cloned := NewHello()
		cloned.sequence = msg.sequence
		cloned.Ipaddress = msg.Ipaddress
		cloned.Mailbox = msg.Mailbox
		for idx, str := range msg.Groups {
			cloned.Groups[idx] = str
		}
		cloned.Status = msg.Status
		for key, val := range msg.Headers {
			cloned.Headers[key] = val
		}
		return cloned

	case *Whisper:
		cloned := NewWhisper()
		cloned.sequence = msg.sequence
		cloned.Content = append(cloned.Content, msg.Content...)
		return cloned

	case *Shout:
		cloned := NewShout()
		cloned.sequence = msg.sequence
		cloned.Group = msg.Group
		cloned.Content = append(cloned.Content, msg.Content...)
		return cloned

	case *Join:
		cloned := NewJoin()
		cloned.sequence = msg.sequence
		cloned.Group = msg.Group
		cloned.Status = msg.Status
		return cloned

	case *Leave:
		cloned := NewLeave()
		cloned.sequence = msg.sequence
		cloned.Group = msg.Group
		cloned.Status = msg.Status
		return cloned

	case *Ping:
		cloned := NewPing()
		cloned.sequence = msg.sequence
		return cloned

	case *PingOk:
		cloned := NewPingOk()
		cloned.sequence = msg.sequence
		return cloned

	}

	return nil
}

// putString marshals a string into the buffer
func putString(buffer *bytes.Buffer, str string) {
	size := len(str)
	if size > StringMax {
		size = StringMax
	}
	sz := fmt.Sprintf("%d", size)
	str = fmt.Sprintf("%"+sz+"s", str)
	binary.Write(buffer, binary.BigEndian, byte(size))
	binary.Write(buffer, binary.BigEndian, []byte(str))
}

// putKeyValString marshals a key=val pair into the buffer
func putKeyValString(buffer *bytes.Buffer, key, val string) {
	str := fmt.Sprintf("%s=%s", key, val)
	putString(buffer, str)
}

// getString unmarshals a string from the buffer
func getString(buffer *bytes.Buffer) string {
	var size byte
	binary.Read(buffer, binary.BigEndian, &size)
	str := make([]byte, size)
	binary.Read(buffer, binary.BigEndian, &str)
	return string(str)
}

// getKeyValString unmarshals a key=val pair from the buffer
func getKeyValString(buffer *bytes.Buffer) (key, val string) {
	str := getString(buffer)
	strs := strings.SplitN(str, "=", 2)

	if len(strs) == 2 {
		return strs[0], strs[1]
	}

	return "", ""
}
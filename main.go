package main

import (
	"binprot/protocol"
	"fmt"
)

func main() {

	p := protocol.Protocol{
		Version:     1,
		PayloadType: protocol.CREATE_QUEUE,
		Payload:     []byte{},
		KeyValuePairs: map[string]string{
			"foo": "bar",
		},
	}

	buf, ok := protocol.StringifyProtocol(p)
	if ok != nil {
		fmt.Println("Error stringifying protocol:", ok)
		return
	}

	p, err := protocol.ParseProtocol(buf)
	if err != nil {
		fmt.Println("Error parsing protocol:", err)
		return
	}
	fmt.Printf(`	Version: %d
	Payload Type: %d
	Key-Value Pairs: %v`, p.Version, p.PayloadType, p.KeyValuePairs)
}

package aimcp

import (
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func marshal(pb proto.Message) (string, error) {
	// Marshal the message to JSON
	data, err := protojson.MarshalOptions{Indent: " ", EmitUnpopulated: true}.Marshal(pb)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

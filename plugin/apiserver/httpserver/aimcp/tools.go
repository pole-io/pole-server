package aimcp

import (
	"github.com/gogo/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
)

func marshal(pb proto.Message) (string, error) {
	m := jsonpb.Marshaler{Indent: " ", EmitDefaults: true}
	// Marshal the message to JSON
	jsonStr, err := m.MarshalToString(pb)
	if err != nil {
		return "", err
	}
	return jsonStr, nil
}

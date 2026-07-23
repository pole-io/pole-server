package rules

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// OwnerNamespaceFromProto reads the governance resource ownership namespace.
// Reflection keeps the control-plane source compatible while specification
// clients roll forward to the new common namespace field.
func OwnerNamespaceFromProto(message proto.Message) string {
	if message == nil {
		return ""
	}
	ref := message.ProtoReflect()
	field := ref.Descriptor().Fields().ByName(protoreflect.Name("namespace"))
	if field == nil || field.Kind() != protoreflect.StringKind {
		return ""
	}
	return ref.Get(field).String()
}

// SetOwnerNamespaceOnProto writes the governance resource ownership namespace
// when the linked specification already exposes the field.
func SetOwnerNamespaceOnProto(message proto.Message, namespace string) {
	if message == nil {
		return
	}
	ref := message.ProtoReflect()
	field := ref.Descriptor().Fields().ByName(protoreflect.Name("namespace"))
	if field == nil || field.Kind() != protoreflect.StringKind {
		return
	}
	ref.Set(field, protoreflect.ValueOfString(namespace))
}

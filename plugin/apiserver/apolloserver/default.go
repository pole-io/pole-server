package apolloserver

import (
	"github.com/pole-io/pole-server/apis/apiserver"
	"github.com/pole-io/pole-server/pluginapi"
)

const (
	Protocol    string = "service-apollo"
	DefaultPort uint32 = 8080
)

func Register(registry *pluginapi.Registry) error {
	return apiserver.RegisterFactory(registry, pluginapi.Descriptor{
		Kind:   pluginapi.KindAPIServer,
		Name:   Protocol,
		Origin: pluginapi.OriginBuiltin,
	}, func() (apiserver.Apiserver, error) {
		return &ApolloServer{}, nil
	})
}

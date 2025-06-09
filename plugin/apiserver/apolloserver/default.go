package apolloserver

import "github.com/pole-io/pole-server/apis/apiserver"

const (
	Protocol    string = "service-apollo"
	DefaultPort uint32 = 8080
)

func init() {
	// 注册Apollo API服务器
	if err := apiserver.Register(Protocol, &ApolloServer{}); err != nil {
		panic(err)
	}
}

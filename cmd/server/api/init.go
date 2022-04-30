package api

import (
	"github.com/saitofun/chat/cmd/config"
	"github.com/saitofun/chat/pkg/depends/protoc"
	"github.com/saitofun/qlib/net/qsock"
)

var (
	server *qsock.Server
)

func init() {
	var err error
	server, err = qsock.NewServer(
		qsock.ServerOptionConnCap(10),
		qsock.ServerOptionListenAddr(config.ServerAddr),
		qsock.ServerOptionProtocol(qsock.ProtocolTCP),
		qsock.ServerOptionParser(protoc.Parser),
		qsock.ServerOptionHandler(func(ev *qsock.Event) {
			switch pl := ev.Payload().(type) {
			case *protoc.Echo:
				OnEcho(pl, ev.Endpoint())
			case *protoc.Instruct:
				OnGmCmd(pl, ev.Endpoint())

			}
		}),
	)
	if err != nil {
		panic(err)
	}
}

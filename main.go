package main

import (
	"context"
	"net"
	"os"
	"sync"

	"github.com/johntech-o/grace/gracenet"
)

var (
	GlobalContext    context.Context
	GlobalCancelFunc context.CancelFunc
	GlobalNet        = gracenet.Net{}
)

var (
	MuxStat     = NewMuxStats()
	GatewayStat = NewGatewayStats()
	MachineStat = NewMachineStats()
)

var MuxManager *MuxServer

var CgiPoolManager = NewCgiConnPoolManager()

var CGIErrorCollection *ErrorCollection

var CGITaskDebug *TaskDebug

var HttpConfigManager *HttpConfig

func init() {
	GlobalContext, GlobalCancelFunc = func() (context.Context, context.CancelFunc) {
		return context.WithCancel(context.Background())
	}()
	MuxManager = NewMuxServer(GlobalContext)
}

func main() {
	var (
		err    error
		l1, l2 net.Listener
		wg     sync.WaitGroup
	)
	println("start bus: ", os.Getpid())

	if err = initSetting(); err != nil {
		goto exit
	}

	l1, err = GlobalNet.Listen("tcp", ServerConf().GatewayIP+":"+ServerConf().GatewayPort)
	if err != nil {
		goto exit
	}
	l2, err = GlobalNet.Listen("tcp", ServerConf().AdminIP+":"+ServerConf().AdminPort)
	if err != nil {
		goto exit
	}
	go signalHandler(l1, l2)

	go MuxManager.Serve()

	go MuxStat.Run()
	go GatewayStat.Run()
	go MachineStat.Run()

	wg.Add(1)
	go func() {
		defer wg.Done()
		gw := NewGateway(ServerConf().GatewayIP, ServerConf().GatewayPort, MuxManager)
		err := gw.Serve(l1)
		flog.Error("gateway serve finish err", err)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := AdminServer(l2)
		flog.Error("gateway serve finish err", err)
	}()
	wg.Wait()

exit:
	flog.Error("service stop", err)
	flog.Error("bus process end", os.Getpid())
	println("stop bus: ", os.Getpid(), err.Error())
	os.Exit(0)
}

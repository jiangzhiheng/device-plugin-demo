package deviceplugin

import (
	"context"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"k8s.io/klog/v2"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"net"
	"os"
	"path"
	"syscall"
	"time"
)

type MyDevicePlugin struct {
	server *grpc.Server
	stop   chan struct{}
	dm     *DeviceMonitor
}

func (c *MyDevicePlugin) Run() error {
	// list all device
	err := c.dm.List()
	if err != nil {
		klog.Fatalf("list device error:%v", err)
	}

	go func() {
		if err = c.dm.Watch(); err != nil {
			klog.Errorf("watch devices error:%v", err)
		}
	}()

	// register device plugin to grpc server
	pluginapi.RegisterDevicePluginServer(c.server, c)

	// delete old unix socket before start
	socket := path.Join(pluginapi.DevicePluginPath, DeviceSocket)
	err = syscall.Unlink(socket)
	if err != nil && !os.IsNotExist(err) {
		return errors.WithMessagef(err, "delete socket %s failed", socket)
	}

	sock, err := net.Listen("unix", socket)
	go c.server.Serve(sock)

	// 发起阻塞连接等待服务器启动
	conn, err := connect(DeviceSocket, 5*time.Second)
	if err != nil {
		return err
	}
	conn.Close()
	return nil
}

func connect(socketPath string, timeout time.Duration) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	c, err := grpc.DialContext(ctx, socketPath,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		// 使用 grpc.WithBlock() 设置连接为阻塞模式，这意味着 DialContext 将一直阻塞直到连接成功或超时。
		grpc.WithBlock(),
		grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			if deadline, ok := ctx.Deadline(); ok {
				return net.DialTimeout("unix", addr, time.Until(deadline))
			}
			return net.DialTimeout("unix", addr, ConnectTimeout)
		}),
	)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func NewMyDevicePlugin() *MyDevicePlugin {
	return &MyDevicePlugin{
		server: grpc.NewServer(grpc.EmptyServerOption{}),
		stop:   make(chan struct{}),
		dm:     NewDeviceMonitor(DevicePath),
	}
}

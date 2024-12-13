package deviceplugin

import (
	"context"
	"github.com/pkg/errors"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"path"
)

func (c *MyDevicePlugin) Register() error {
	conn, err := connect(pluginapi.KubeletSocket, ConnectTimeout)
	if err != nil {
		return errors.WithMessagef(err, "connect to %s failed", pluginapi.KubeletSocket)
	}
	defer conn.Close()

	client := pluginapi.NewRegistrationClient(conn)
	req := &pluginapi.RegisterRequest{
		Version:      pluginapi.Version,
		Endpoint:     path.Base(DeviceSocket),
		ResourceName: ResourceName,
	}
	_, err = client.Register(context.Background(), req)
	if err != nil {
		return errors.WithMessage(err, "register to kubelet failed")
	}
	return nil
}

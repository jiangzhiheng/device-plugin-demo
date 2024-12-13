package deviceplugin

import (
	"context"
	"github.com/pkg/errors"
	"k8s.io/klog/v2"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"strings"
)

/*
这里实现DevicePluginServer的5个grpc方法，用于kubelet调用设备插件
*/

// GetDevicePluginOptions 这个接口用于获取设备插件的信息，可以在其返回的响应中指定一些设备插件的配置选项，可以看做是插件的元数据
func (c *MyDevicePlugin) GetDevicePluginOptions(ctx context.Context, _ *pluginapi.Empty) (*pluginapi.DevicePluginOptions, error) {
	return &pluginapi.DevicePluginOptions{
		PreStartRequired: true,
	}, nil
}

// ListAndWatch 该接口用于列出可用的设备并持续监视这些设备的状态变化
// 返回设备列表的流，每当设备状态更改或设备消失时，ListAndWatch 会返回新的列表
func (c *MyDevicePlugin) ListAndWatch(_ *pluginapi.Empty, srv pluginapi.DevicePlugin_ListAndWatchServer) error {
	devs := c.dm.Devices()
	klog.Infof("find devices [%s]", String(devs))
	err := srv.Send(&pluginapi.ListAndWatchResponse{
		Devices: devs,
	})

	if err != nil {
		return errors.WithMessage(err, "send device failed")
	}

	klog.Infof("waiting for device update")
	for range c.dm.notify {
		devs = c.dm.Devices()
		klog.Infof("device update,new device list [%s]", String(devs))
		_ = srv.Send(&pluginapi.ListAndWatchResponse{
			Devices: devs,
		})
	}
	return nil
}

// GetPreferredAllocation 将分配偏好信息提供给 device plugin,以便 device plugin 在分配时可以做出更好的选择
func (c *MyDevicePlugin) GetPreferredAllocation(_ context.Context, _ *pluginapi.PreferredAllocationRequest) (*pluginapi.PreferredAllocationResponse, error) {
	return &pluginapi.PreferredAllocationResponse{}, nil
}

// Allocate 该接口用于向设备插件请求分配指定数量的设备资源。 Allocate 在容器创建期间被调用，以便设备插件能够执行设备特定的操作，并指导 Kubelet 完成使设备在容器内可用的必要步骤。
func (c *MyDevicePlugin) Allocate(ctx context.Context, reqs *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse, error) {
	ret := &pluginapi.AllocateResponse{}
	for _, req := range reqs.ContainerRequests {
		klog.Infof("[Allocate] reveived request:%v", strings.Join(req.DevicesIDs, ","))
		resp := pluginapi.ContainerAllocateResponse{
			Envs: map[string]string{
				"device": strings.Join(req.DevicesIDs, ","),
			},
		}
		ret.ContainerResponses = append(ret.ContainerResponses, &resp)
	}
	return ret, nil
}

// PreStartContainer 该接口在容器启动之前调用，用于配置容器使用的设备资源。
func (c *MyDevicePlugin) PreStartContainer(ctx context.Context, reqs *pluginapi.PreStartContainerRequest) (*pluginapi.PreStartContainerResponse, error) {
	return &pluginapi.PreStartContainerResponse{}, nil
}

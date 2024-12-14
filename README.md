### device-plugin-demo
研究k8s device plugin的实现，假设我的设备是在`/etc/mydevice`下，例如`/etc/mydevice/device1，/etc/mydevice/device2`

### 实现步骤
#### step1 注册设备到kubelet
#### step2 实现 ` DevicePluginServer`接口
具体实现在`pkg/deviceplugin/api.go`
```go
type DevicePluginServer interface {
    GetDevicePluginOptions(context.Context, *Empty) (*DevicePluginOptions, error)
    ListAndWatch(*Empty, DevicePlugin_ListAndWatchServer) error
    GetPreferredAllocation(context.Context, *PreferredAllocationRequest) (*PreferredAllocationResponse, error)
    Allocate(context.Context, *AllocateRequest) (*AllocateResponse, error)
    PreStartContainer(context.Context, *PreStartContainerRequest) (*PreStartContainerResponse, error)
}
```

### 部署
#### 部署设备插件
```shell
# 创建设备模拟文件
mkdir /etc/mydevice && cd  /etc/mydevice
touch device1

make all
k apply -f _output/template/deploy.yaml
```
#### 查看启动日志
```shell
k logs device-plugin-demo-lj54m
I1214 11:26:37.043858       1 main.go:10] my device plugin staring
I1214 11:26:37.044049       1 devicemonitor.go:30] /etc/mydevice is dir,skip
I1214 11:26:37.044143       1 devicemonitor.go:43] watching device
I1214 11:26:38.046945       1 api.go:26] find devices [device1]
I1214 11:26:38.047004       1 api.go:35] waiting for device update
```

#### 查看是否注册到k8s
```shell
#k describe node xxxx

# ···
Capacity:
  cpu:                  16
  ephemeral-storage:    103019184Ki
  hugepages-1Gi:        0
  hugepages-2Mi:        0
  memory:               32609268Ki
  mydevice.com/device:  1
  pods:                 110
···
```
#### 创建测试pod
```shell
k apply -f deploy test-pod.yaml

k exec -it xxxx -- sh
# 进入pod之后查看设备
/ # env|grep device
device=device1
```
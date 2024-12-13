### device-plugin-demo
研究k8s device plugin的实现，假设我的设备是在`/etc/mydevice`下，例如`/etc/mydevice/device1，/etc/mydevice/device2`

### 实现步骤
#### step1 注册设备到kubelet

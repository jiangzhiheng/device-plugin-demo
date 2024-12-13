package deviceplugin

import "time"

const (
	ResourceName   string = "mydevice.com/device"
	DevicePath     string = "/etc/mydevice"
	DeviceSocket   string = "mydevice.sock"
	ConnectTimeout        = time.Second * 5
)

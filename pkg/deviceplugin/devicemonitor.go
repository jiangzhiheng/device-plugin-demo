package deviceplugin

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	"io/fs"
	"k8s.io/klog/v2"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"path"
	"path/filepath"
	"strings"
)

type DeviceMonitor struct {
	path    string
	devices map[string]*pluginapi.Device
	notify  chan struct{} // notify when device update
}

// List all device
/*
filepath.Walk 是 Golang 标准库 path/filepath 包中的一个函数，用于遍历文件系统中的目录和文件。它接收两个参数：一个是要遍历的顶级目录的路径，
另一个是一个实现了 filepath.WalkFunc 类型的回调函数。每当遍历到一个文件或目录时，filepath.Walk 就会调用这个回调函数，并传递两个参数：当前遍
历到的文件或目录的路径，以及一个 os.FileInfo 类型的值，其中包含了文件的元信息（如大小、权限等）。
*/
func (d *DeviceMonitor) List() error {
	err := filepath.Walk(d.path, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			klog.Infof("%s is dir,skip", path)
			return nil
		}
		d.devices[info.Name()] = &pluginapi.Device{
			ID:     info.Name(),
			Health: pluginapi.Healthy,
		}
		return nil
	})
	return errors.WithMessagef(err, "walk [%s] failed", d.path)
}

func (d *DeviceMonitor) Watch() error {
	klog.Infoln("watching device")
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return errors.WithMessage(err, "new watcher failed")
	}
	defer w.Close()

	errChan := make(chan error)
	go func() {
		// recover goroutine panic
		defer func() {
			if r := recover(); r != nil {
				errChan <- fmt.Errorf("device watcher panic: %v", r)
			}
		}()
		for {
			select {
			case event, ok := <-w.Events:
				if !ok {
					continue
				}
				klog.Infof("fsnotify device event:%s %s", event.Name, event.Op.String())
				if event.Op == fsnotify.Create {
					dev := path.Base(event.Name)
					d.devices[dev] = &pluginapi.Device{
						ID:     dev,
						Health: pluginapi.Healthy,
					}
					d.notify <- struct{}{}
					klog.Infof("find new device [%s]", dev)
				} else if event.Op&fsnotify.Remove == fsnotify.Remove {
					/*
						event.Op&fsnotify.Remove == fsnotify.Remove 是一个位运算表达式，用于检查 event.Op 是否表示一个 "remove"（删除）事件。
						因为 event.Op 可能表示多种事件类型的组合，而不仅仅是一个单一的事件类型。event.Op 是一个位掩码（bitmask），它将多个事件类型组合在一起。每个事件类型都是一个位，如果该位为 1，则表示该事件类型发生。
						例如，如果 event.Op 的值为 fsnotify.Write | fsnotify.Remove（即 6），则表示同时发生了写和删除事件。在这种情况下，event.Op == fsnotify.Remove 的结果将为 false，因为 event.Op 的值不仅包含删除事件，还包含写事件。
					*/
					dev := path.Base(event.Name)
					delete(d.devices, dev)
					klog.Infof("device [%s] removed", dev)
				}
			case err, ok := <-w.Errors:
				if !ok {
					continue
				}
				klog.Errorf("fsnotify watch device failed: %v", err)
			}
		}
	}()

	err = w.Add(d.path)
	if err != nil {
		return fmt.Errorf("watch device error:%v", err)
	}
	return <-errChan
}

func NewDeviceMonitor(path string) *DeviceMonitor {
	return &DeviceMonitor{
		path:    path,
		devices: make(map[string]*pluginapi.Device),
		notify:  make(chan struct{}),
	}
}

func (d *DeviceMonitor) Devices() []*pluginapi.Device {
	devices := make([]*pluginapi.Device, 0, len(d.devices))
	for _, device := range d.devices {
		devices = append(devices, device)
	}
	return devices
}

func String(devs []*pluginapi.Device) string {
	ids := make([]string, 0, len(devs))
	for _, device := range devs {
		ids = append(ids, device.ID)
	}
	return strings.Join(ids, ",")
}

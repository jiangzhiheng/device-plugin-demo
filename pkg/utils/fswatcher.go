package utils

import (
	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	"k8s.io/klog/v2"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

func WatchKubelet(stop chan<- struct{}) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return errors.WithMessage(err, "Unable to create fsnotify watcher")
	}
	defer watcher.Close()

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					continue
				}
				klog.Infof("fsnotify events: %s %s", event.Name, event.Op.String())
				if event.Name == pluginapi.KubeletSocket && event.Op == fsnotify.Create {
					klog.Warning("inotify: kubelet.sock created, restarting.")
					stop <- struct{}{}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					continue
				}
				klog.Infof("fsnotify failed restarting,detail:%v", err)
			}
		}
	}()
	err = watcher.Add(pluginapi.KubeletSocket)
	if err != nil {
		return errors.WithMessagef(err, "Unable to add path %s to watcher ", pluginapi.KubeletSocket)
	}
	return nil
}
package main

import (
	"github.com/jiangzhiheng/device-plugin-demo/pkg/deviceplugin"
	"github.com/jiangzhiheng/device-plugin-demo/pkg/utils"
	"k8s.io/klog/v2"
)

func main() {
	klog.Infof("my device plugin staring")
	dp := deviceplugin.NewMyDevicePlugin()
	go func() {
		err := dp.Run()
		if err != nil {
			panic(err)
		}
	}()

	// register when device plugin start
	if err := dp.Register(); err != nil {
		klog.Fatalf("register to kubelet failed:%v", err)
	}
	// watch kubelet.sock, when kubelet restart ,exit device plugin, then will restart by daemonset
	stop := make(chan struct{})
	err := utils.WatchKubelet(stop)
	if err != nil {
		klog.Fatalf("Start to kubelet failed:%v", err)
	}
	<-stop
	klog.Infof("kubelet restart. exiting...")
}

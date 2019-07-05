package rke

import (
	"github.com/rancher/types/apis/management.cattle.io/v3"
)

func loadK8sVersionWindowsServiceOptions() map[string]v3.KubernetesServicesOptions {
	// since 1.14, windows has been supported
	return map[string]v3.KubernetesServicesOptions{
		"v1.15": {
			Kubelet:   getWindowsKubeletOptions(),
			Kubeproxy: getWindowsKubeproxyOptions(),
		},
		"v1.14": {
			Kubelet:   getWindowsKubeletOptions(),
			Kubeproxy: getWindowsKubeproxyOptions(),
		},
	}
}

func getWindowsKubeletOptions() map[string]string {
	return map[string]string{
		"tls-cipher-suites":                 tlsCipherSuites,
		"address":                           "0.0.0.0",
		"allow-privileged":                  "true",
		"anonymous-auth":                    "false",
		"authentication-token-webhook":      "true",
		"cgroups-per-qos":                   "false",
		"event-qps":                         "0",
		"network-plugin":                    "cni",
		"read-only-port":                    "0",
		"streaming-connection-idle-timeout": "30m",
		"v":                                 "2",
		"enforce-node-allocatable":          "''",
		"resolv-conf":                       "''",
		"cni-bin-dir":                       "[PREFIX_PATH]/opt/cni/bin",
		"cni-conf-dir":                      "[PREFIX_PATH]/etc/cni/net.d",
		"cert-dir":                          "[PREFIX_PATH]/var/lib/kubelet/pki",
		"volume-plugin-dir":                 "[PREFIX_PATH]/var/lib/kubelet/volumeplugins",
		"kube-reserved":                     "cpu=500m,memory=500Mi,ephemeral-storage=1Gi",
		"system-reserved":                   "cpu=1000m,memory=2Gi,ephemeral-storage=2Gi",
		"image-pull-progress-deadline":      "30m", // windows images always too larger
	}
}

func getWindowsKubeproxyOptions() map[string]string {
	return map[string]string{
		"v":                    "2",
		"proxy-mode":           "kernelspace",
		"healthz-bind-address": "127.0.0.1",
	}
}

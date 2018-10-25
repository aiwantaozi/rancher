package utils

import (
	"fmt"
	"net"

	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"

	loggingconfig "github.com/rancher/rancher/pkg/controllers/user/logging/config"
	"github.com/rancher/rancher/pkg/settings"
	"github.com/rancher/types/apis/management.cattle.io/v3"
	"github.com/rancher/types/config/dialer"
)

const (
	networkTCP  = "tcp"
	networkUDP  = "udp"
	schemaHttps = "https"
)

var (
	timeout = 5 * time.Second
)

type TLS struct {
	EnableTLS          bool
	Cert               string
	Key                string
	Ca                 string
	InsecureSkipVerify bool
}

type WrapLogging struct {
	CurrentTarget string
	endpoint      []string
	network       string
	WrapSyslog
	WrapSplunk
	WrapElasticsearch
	WrapKafka
	WrapFluentForwarder
	TLS
}

type WrapClusterLogging struct {
	v3.ClusterLoggingSpec
	ExcludeNamespace string
	WrapLogging
}

type WrapProjectLogging struct {
	v3.ProjectLoggingSpec
	ClusterName     string
	GrepNamespace   string
	IsSystemProject bool
	WrapLogging
	WrapProjectName string
}

type WrapElasticsearch struct {
	DateFormat string
	Host       string
	Scheme     string
}

type WrapSplunk struct {
	Host   string
	Port   string
	Scheme string
}

type WrapKafka struct {
	Brokers   string
	Zookeeper string
}

type WrapSyslog struct {
	Host string
	Port string
}

type WrapFluentForwarder struct {
	EnableShareKey bool
	FluentServers  []FluentServer
}

type FluentServer struct {
	Host string
	Port string
	v3.FluentServer
}

func CheckEndpoint(dialer dialer.Factory, clusterName, protocal string, endpoints []string) error {
	for _, v := range endpoints {
		netDialerErr := dial(protocal, v, netDialer)
		if netDialerErr == nil {
			fmt.Println("--here 1")
			continue
		}
		clusterDialer, err := dialer.ClusterDialer(clusterName)
		if err != nil {
			return err
		}

		clusterDialErr := dial(protocal, v, clusterDialer)
		if clusterDialErr == nil {
			fmt.Println("--here 2")
			continue
		}
		return fmt.Errorf("endpoint %v is not reachable on the internet %v, also not reachable inside cluster %v", v, netDialerErr, clusterDialErr)
	}
	return nil
}

func dial(protocal string, endpoint string, dailer func(network, address string) (net.Conn, error)) error {
	rawConn, err := dailer(protocal, endpoint)
	if err != nil {
		return fmt.Errorf("create raw conn failed, %v", err)
	}
	rawConn.Close()
	return nil
}

// func (w *WrapClusterLogging) Validate(dialer dialer.Factory) error {
// 	fmt.Println("---------0")
// 	wrapLogging, err := GetWrapConfig(w.ElasticsearchConfig, w.SplunkConfig, w.SyslogConfig, w.KafkaConfig, w.FluentForwarderConfig)
// 	if err != nil {
// 		return err
// 	}
// 	fmt.Println("---------00")

// 	w.WrapLogging = wrapLogging
// 	err = checkEndpoint(dialer, w.ClusterName)
// 	return err
// }

// func (w *WrapProjectLogging) Validate(dialer dialer.Factory) error {
// 	wrapLogging, err := GetWrapConfig(w.ElasticsearchConfig, w.SplunkConfig, w.SyslogConfig, w.KafkaConfig, w.FluentForwarderConfig)
// 	if err != nil {
// 		return err
// 	}

// 	w.WrapLogging = wrapLogging
// 	return w.checkEndpoint(dialer, w.ClusterName)
// }

func ToWrapClusterLogging(clusterLogging v3.ClusterLoggingSpec) (*WrapClusterLogging, error) {
	excludeNamespace := strings.Replace(settings.SystemNamespaces.Get(), ",", "|", -1)
	wp := WrapClusterLogging{
		ExcludeNamespace:   excludeNamespace,
		ClusterLoggingSpec: clusterLogging,
	}

	wrapLogging, err := GetWrapConfig(clusterLogging.ElasticsearchConfig, clusterLogging.SplunkConfig, clusterLogging.SyslogConfig, clusterLogging.KafkaConfig, clusterLogging.FluentForwarderConfig)
	if err != nil {
		return nil, err
	}
	wp.WrapLogging = wrapLogging
	return &wp, nil
}

func ToWrapProjectLogging(grepNamespace string, isSystemProject bool, projectLogging v3.ProjectLoggingSpec) (*WrapProjectLogging, error) {
	wp := WrapProjectLogging{
		ProjectLoggingSpec: projectLogging,
		GrepNamespace:      grepNamespace,
		IsSystemProject:    isSystemProject,
		WrapProjectName:    strings.Replace(projectLogging.ProjectName, ":", "_", -1),
	}

	wrapLogging, err := GetWrapConfig(projectLogging.ElasticsearchConfig, projectLogging.SplunkConfig, projectLogging.SyslogConfig, projectLogging.KafkaConfig, projectLogging.FluentForwarderConfig)
	if err != nil {
		return nil, err
	}
	wp.WrapLogging = wrapLogging
	return &wp, nil
}

func GetWrapConfig(es *v3.ElasticsearchConfig, sp *v3.SplunkConfig, sl *v3.SyslogConfig, kf *v3.KafkaConfig, ff *v3.FluentForwarderConfig) (wrapLogging WrapLogging, err error) {
	if es != nil {
		var h, s string
		h, s, err = parseEndpoint(es.Endpoint)
		if err != nil {
			return
		}

		wrapLogging.WrapElasticsearch = WrapElasticsearch{
			Host:       h,
			Scheme:     s,
			DateFormat: getDateFormat(es.DateFormat),
		}

		wrapLogging.CurrentTarget = loggingconfig.Elasticsearch
		wrapLogging.network = networkTCP
		wrapLogging.endpoint = []string{h}
		if s == schemaHttps {
			wrapLogging.TLS = TLS{
				EnableTLS: true,
				Ca:        es.Certificate,
				Key:       es.ClientKey,
				Cert:      es.ClientCert,
			}
		}
	}

	if sp != nil {
		var h, s, host, port string
		h, s, err = parseEndpoint(sp.Endpoint)
		if err != nil {
			return
		}

		host, port, err = net.SplitHostPort(h)
		if err != nil {
			return
		}
		wrapLogging.WrapSplunk = WrapSplunk{
			Host:   host,
			Port:   port,
			Scheme: s,
		}

		wrapLogging.CurrentTarget = loggingconfig.Splunk
		wrapLogging.network = networkTCP
		wrapLogging.endpoint = []string{h}
		if s == schemaHttps {
			wrapLogging.TLS = TLS{
				EnableTLS: s == schemaHttps,
				Ca:        sp.Certificate,
				Key:       sp.ClientKey,
				Cert:      sp.ClientCert,
			}
		}
	}

	if sl != nil {
		var host, port string
		host, port, err = net.SplitHostPort(sl.Endpoint)
		if err != nil {
			return
		}
		wrapLogging.WrapSyslog = WrapSyslog{
			Host: host,
			Port: port,
		}
		wrapLogging.CurrentTarget = loggingconfig.Syslog
		wrapLogging.network = sl.Protocol
		wrapLogging.endpoint = []string{sl.Endpoint}
		if sl.Protocol == networkTCP && (sl.Certificate != "" || sl.ClientCert != "") {
			wrapLogging.TLS = TLS{
				EnableTLS: true,
				Ca:        sl.Certificate,
				Key:       sl.ClientKey,
				Cert:      sl.ClientCert,
			}
		}
	}

	if kf != nil {
		if len(kf.BrokerEndpoints) == 0 && kf.ZookeeperEndpoint == "" {
			err = errors.New("one of the kafka endpoint must be set")
			return
		}
		if len(kf.BrokerEndpoints) != 0 {
			var bs []string
			var h string
			for _, v := range kf.BrokerEndpoints {
				h, _, err = parseEndpoint(v)
				if err != nil {
					return
				}
				bs = append(bs, h)
			}

			wrapLogging.WrapKafka = WrapKafka{
				Brokers: strings.Join(bs, ","),
			}
			if sl.Protocol == networkTCP && (kf.Certificate != "" || kf.ClientCert != "") {
				wrapLogging.TLS = TLS{
					EnableTLS: true,
					Ca:        kf.Certificate,
					Key:       kf.ClientKey,
					Cert:      kf.ClientCert,
				}
			}
		} else {
			if kf.ZookeeperEndpoint != "" {
				var h string
				if h, _, err = parseEndpoint(kf.ZookeeperEndpoint); err != nil {
					return
				}
				wrapLogging.WrapKafka = WrapKafka{
					Zookeeper: h,
				}
			}
		}
		wrapLogging.CurrentTarget = loggingconfig.Kafka
		wrapLogging.network = networkTCP
		wrapLogging.endpoint = []string{kf.ZookeeperEndpoint}
	}

	if ff != nil {
		var enableShareKey bool
		var fss []FluentServer
		for _, v := range ff.FluentServers {
			var host, port string
			host, port, err = net.SplitHostPort(v.Endpoint)
			if err != nil {
				return
			}
			if v.SharedKey != "" {
				enableShareKey = true
			}
			fs := FluentServer{
				Host:         host,
				Port:         port,
				FluentServer: v,
			}
			fss = append(fss, fs)
			wrapLogging.endpoint = append(wrapLogging.endpoint, v.Endpoint)
		}
		wrapLogging.WrapFluentForwarder = WrapFluentForwarder{
			EnableShareKey: enableShareKey,
			FluentServers:  fss,
		}
		if ff.EnableTLS {
			wrapLogging.TLS = TLS{
				EnableTLS: true,
				Ca:        ff.Certificate,
			}
		}
		wrapLogging.CurrentTarget = loggingconfig.FluentForwarder
		wrapLogging.network = networkTCP
	}

	return
}

func parseEndpoint(endpoint string) (host string, scheme string, err error) {
	u, err := url.ParseRequestURI(endpoint)
	if err != nil {
		return "", "", errors.Wrapf(err, "invalid endpoint %s", endpoint)
	}

	if u.Host == "" || u.Scheme == "" {
		return "", "", fmt.Errorf("invalid endpoint %s, empty host or schema", endpoint)
	}

	return u.Host, u.Scheme, nil
}

func getDateFormat(dateformat string) string {
	ToRealMap := map[string]string{
		"YYYY-MM-DD": "%Y-%m-%d",
		"YYYY-MM":    "%Y-%m",
		"YYYY":       "%Y",
	}
	if _, ok := ToRealMap[dateformat]; ok {
		return ToRealMap[dateformat]
	}
	return "%Y-%m-%d"
}

func netDialer(network string, url string) (net.Conn, error) {
	return net.DialTimeout(network, url, timeout)
}

func testReachable(network string, url string) error {
	timeout := time.Duration(10 * time.Second)
	conn, err := net.DialTimeout(network, url, timeout)
	if err != nil {
		return fmt.Errorf("url %s unreachable, error: %v", url, err)
	}
	conn.Close()
	return nil
}

func GetClusterTarget(spec v3.ClusterLoggingSpec) string {
	if spec.ElasticsearchConfig != nil {
		return loggingconfig.Elasticsearch
	} else if spec.SplunkConfig != nil {
		return loggingconfig.Splunk
	} else if spec.KafkaConfig != nil {
		return loggingconfig.Kafka
	} else if spec.SyslogConfig != nil {
		return loggingconfig.Syslog
	} else if spec.FluentForwarderConfig != nil {
		return loggingconfig.FluentForwarder
	}
	return "none"
}

func GetProjectTarget(spec v3.ProjectLoggingSpec) string {
	if spec.ElasticsearchConfig != nil {
		return loggingconfig.Elasticsearch
	} else if spec.SplunkConfig != nil {
		return loggingconfig.Splunk
	} else if spec.KafkaConfig != nil {
		return loggingconfig.Kafka
	} else if spec.SyslogConfig != nil {
		return loggingconfig.Syslog
	} else if spec.FluentForwarderConfig != nil {
		return loggingconfig.FluentForwarder
	}
	return "none"
}

package utils

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/rancher/types/apis/management.cattle.io/v3"
	"github.com/rancher/types/config/dialer"

	loggingconfig "github.com/rancher/rancher/pkg/controllers/user/logging/config"
)

const (
	sslv23  = "SSLv23"
	tlsv1   = "TLSv1"
	tlsv1_1 = "TLSv1_1"
	tlsv1_2 = "TLSv1_2"
)

type WrapLogging struct {
	CurrentTarget string
	*WrapSyslog
	*WrapSplunk
	*WrapElasticsearch
	*WrapKafka
	*WrapFluentForwarder
}

type WrapClusterLogging struct {
	v3.LoggingCommonField
	*WrapLogging
	ExcludeNamespace       string
	ExcludeSystemComponent bool
}

type WrapProjectLogging struct {
	v3.LoggingCommonField
	*WrapLogging
	GrepNamespace   string
	IsSystemProject bool
	WrapProjectName string
}

func NewWrapClusterLogging(logging v3.ClusterLoggingSpec, excludeNamespace string) (*WrapClusterLogging, error) {
	wrapLogging := NewWrapLogging(logging.ElasticsearchConfig, logging.SplunkConfig, logging.SyslogConfig, logging.KafkaConfig, logging.FluentForwarderConfig)

	if err := wrapLogging.Wrapper(); err != nil {
		return nil, errors.Wrapf(err, "wrapper logging target failed")
	}

	return &WrapClusterLogging{
		LoggingCommonField:     logging.LoggingCommonField,
		WrapLogging:            wrapLogging,
		ExcludeNamespace:       excludeNamespace,
		ExcludeSystemComponent: logging.ExcludeSystemComponent,
	}, nil
}

func NewWrapProjectLogging(logging v3.ProjectLoggingSpec, grepNamespace string, isSystemProject bool) (*WrapProjectLogging, error) {
	wrapLogging := NewWrapLogging(logging.ElasticsearchConfig, logging.SplunkConfig, logging.SyslogConfig, logging.KafkaConfig, logging.FluentForwarderConfig)
	if err := wrapLogging.Wrapper(); err != nil {
		return nil, err
	}

	wrapProjectName := strings.Replace(logging.ProjectName, ":", "_", -1)
	return &WrapProjectLogging{
		LoggingCommonField: logging.LoggingCommonField,
		WrapLogging:        wrapLogging,
		GrepNamespace:      grepNamespace,
		IsSystemProject:    isSystemProject,
		WrapProjectName:    wrapProjectName,
	}, nil
}

type LoggingTarget interface {
	TestReachable(dial dialer.Dialer) error
}

type WrapElasticsearch struct {
	*v3.ElasticsearchConfig
	DateFormat string
	Host       string
	Scheme     string
}

type WrapSplunk struct {
	*v3.SplunkConfig
	Host   string
	Port   string
	Scheme string
}

type WrapKafka struct {
	*v3.KafkaConfig
	Brokers   string
	Zookeeper string
}

type WrapSyslog struct {
	*v3.SyslogConfig
	Host string
	Port string
}

type WrapFluentForwarder struct {
	*v3.FluentForwarderConfig
	EnableShareKey bool
	FluentServers  []FluentServer
}

type FluentServer struct {
	Host string
	Port string
	v3.FluentServer
}

func (w *WrapElasticsearch) TestReachable(dial dialer.Dialer) error {
	testReachable2(dial, "test")
	return nil
}

func (w *WrapSplunk) TestReachable(dial dialer.Dialer) error {
	fmt.Println("---here")
	fmt.Printf("%+v", w)
	return testReachable2(dial, w.Endpoint)
}

func (w *WrapKafka) TestReachable(dial dialer.Dialer) error {
	testReachable2(dial, "test")
	return nil
}

func (w *WrapSyslog) TestReachable(dial dialer.Dialer) error {
	testReachable2(dial, "test")
	return nil
}

func (w *WrapFluentForwarder) TestReachable(dial dialer.Dialer) error {
	testReachable2(dial, "test")
	return nil
}

func (w *WrapLogging) GetLoggingTarget() LoggingTarget {
	if w.WrapElasticsearch != nil {
		return w.WrapElasticsearch
	} else if w.WrapSplunk != nil {
		return w.WrapSplunk
	} else if w.WrapSyslog != nil {
		return w.WrapSyslog
	} else if w.WrapKafka != nil {
		return w.WrapKafka
	} else if w.WrapFluentForwarder != nil {
		return w.WrapFluentForwarder
	}
	return nil
}

func (w *WrapLogging) Wrapper() error {
	if w.WrapElasticsearch != nil {
		h, s, err := parseEndpoint(w.ElasticsearchConfig.Endpoint)
		if err != nil {
			return err
		}
		w.WrapElasticsearch.Host = h
		w.WrapElasticsearch.Scheme = s
		w.WrapElasticsearch.DateFormat = getDateFormat(w.ElasticsearchConfig.DateFormat)
		w.CurrentTarget = loggingconfig.Elasticsearch
	}

	if w.WrapSplunk != nil {
		h, s, err := parseEndpoint(w.SplunkConfig.Endpoint)
		if err != nil {
			return err
		}

		host, port, err := net.SplitHostPort(h)
		if err != nil {
			return err
		}

		w.WrapSplunk.Host = host
		w.WrapSplunk.Scheme = s
		w.WrapSplunk.Port = port
		w.CurrentTarget = loggingconfig.Splunk
	}

	if w.WrapSyslog != nil {
		host, port, err := net.SplitHostPort(w.SyslogConfig.Endpoint)
		if err != nil {
			return err
		}

		w.WrapSyslog.Host = host
		w.WrapSyslog.Port = port
		w.CurrentTarget = loggingconfig.Syslog
	}

	if w.WrapKafka != nil {
		if len(w.KafkaConfig.BrokerEndpoints) == 0 && w.KafkaConfig.ZookeeperEndpoint == "" {
			err := errors.New("one of the kafka endpoint must be set")
			return err
		}
		if len(w.KafkaConfig.BrokerEndpoints) != 0 {
			var bs []string
			for _, v := range w.KafkaConfig.BrokerEndpoints {
				h, _, err := parseEndpoint(v)
				if err != nil {
					return err
				}
				bs = append(bs, h)
			}
			w.WrapKafka.Brokers = strings.Join(bs, ",")
		} else {
			if w.KafkaConfig.ZookeeperEndpoint != "" {
				h, _, err := parseEndpoint(w.KafkaConfig.ZookeeperEndpoint)
				if err != nil {
					return err
				}
				w.WrapKafka.Zookeeper = h
			}
		}
		w.CurrentTarget = loggingconfig.Kafka
	}

	if w.WrapFluentForwarder != nil {
		var enableShareKey bool
		var fss []FluentServer
		for _, v := range w.FluentForwarderConfig.FluentServers {
			host, port, err := net.SplitHostPort(v.Endpoint)
			if err != nil {
				return err
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
		}
		w.WrapFluentForwarder = &WrapFluentForwarder{
			EnableShareKey: enableShareKey,
			FluentServers:  fss,
		}
		w.CurrentTarget = loggingconfig.FluentForwarder
	}

	return nil
}

func NewWrapLogging(es *v3.ElasticsearchConfig, sp *v3.SplunkConfig, sl *v3.SyslogConfig, kf *v3.KafkaConfig, ff *v3.FluentForwarderConfig) (wrapLogging *WrapLogging) {
	wp := &WrapLogging{}
	if es != nil {
		wp.WrapElasticsearch = &WrapElasticsearch{ElasticsearchConfig: es}
	} else if sp != nil {

		wp.WrapSplunk = &WrapSplunk{SplunkConfig: sp}
	} else if sp != nil {

		wp.WrapSyslog = &WrapSyslog{SyslogConfig: sl}
	} else if kf != nil {

		wp.WrapKafka = &WrapKafka{KafkaConfig: kf}
	} else if ff != nil {

		wp.WrapFluentForwarder = &WrapFluentForwarder{FluentForwarderConfig: ff}
	}
	return wp
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

func testReachable(network string, url string) error {
	timeout := time.Duration(10 * time.Second)
	conn, err := net.DialTimeout(network, url, timeout)
	if err != nil {
		return fmt.Errorf("url %s unreachable, error: %v", url, err)
	}
	conn.Close()
	return nil
}

func testReachable2(dial dialer.Dialer, url, rootCA, clientCert, clientKey, clientKeyPass, sslVersion string, sslVerify bool) error {
	tlsConfig, err := buildTLSConfig(rootCA, clientCert, clientKey, clientKeyPass, sslVersion, sslVerify)
	if err != nil {
		return errors.Wrap(err, "build tls config failed")
	}

	transport := &http.Transport{
		Dial:            dial,
		TLSClientConfig: tlsConfig,
	}

	client := http.Client{
		Transport: transport,
		Timeout:   5 * time.Second,
	}

	if _, err = client.Head(url); err != nil {
		return fmt.Errorf("url %s unreachable, error: %v", url, err)
	}
	return nil
}

func buildTLSConfig(rootCA, clientCert, clientKey, clientKeyPass, sslVersion string, sslVerify bool) (*tls.Config, error) {
	config := &tls.Config{
		InsecureSkipVerify: !sslVerify,
	}

	if clientCert != "" {
		cert, err := tls.LoadX509KeyPair(clientCert, clientKey)
		if err != nil {
			return nil, errors.Wrap(err, "load client cert and key failed")
		}

		config.Certificates = []tls.Certificate{cert}
	}

	if rootCA != "" {
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM([]byte(rootCA))

		config.RootCAs = caCertPool
	}

	if sslVersion != "" {
		switch sslVersion {
		case sslv23:
			config.MaxVersion = tls.VersionSSL30
		case tlsv1:
			config.MaxVersion = tls.VersionTLS10
			config.MinVersion = tls.VersionTLS10
		case tlsv1_1:
			config.MaxVersion = tls.VersionTLS11
			config.MinVersion = tls.VersionTLS11
		case tlsv1_2:
			config.MaxVersion = tls.VersionTLS12
			config.MinVersion = tls.VersionTLS12
		}
	}

	return config, nil
}

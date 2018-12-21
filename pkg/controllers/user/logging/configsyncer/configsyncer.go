package configsyncer

import (
	"fmt"
	"sort"
	"strings"

	"github.com/rancher/norman/controller"
	loggingconfig "github.com/rancher/rancher/pkg/controllers/user/logging/config"
	"github.com/rancher/rancher/pkg/controllers/user/logging/generator"
	"github.com/rancher/rancher/pkg/project"
	"github.com/rancher/types/apis/core/v1"
	corev1 "github.com/rancher/types/apis/core/v1"
	mgmtv3 "github.com/rancher/types/apis/management.cattle.io/v3"
	"github.com/rancher/types/config"
	"github.com/rancher/types/config/dialer"

	"github.com/pkg/errors"
	k8scorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
)

// This controller is responsible for generate fluentd config
// and updating the config secret
// so the config reload could detect the file change and reload

func NewConfigSyncer(cluster *config.UserContext, SecretManager *SecretManager, isDeploy *bool) *ConfigSyncer {

	clusterName := cluster.ClusterName
	return &ConfigSyncer{
		restConfig:           &cluster.RESTConfig,
		isDeploy:             isDeploy,
		clusterName:          clusterName,
		clusterLoggingLister: cluster.Management.Management.ClusterLoggings(clusterName).Controller().Lister(),
		projectLoggingLister: cluster.Management.Management.ProjectLoggings(metav1.NamespaceAll).Controller().Lister(),
		projectLister:        cluster.Management.Management.Projects(clusterName).Controller().Lister(),
		namespaceLister:      cluster.Core.Namespaces(metav1.NamespaceAll).Controller().Lister(),
		secretManager:        SecretManager,
		pods:                 cluster.Core.Pods(loggingconfig.LoggingNamespace),
		dialer:               cluster.Management.Dialer,
		clusters:             cluster.Management.Management.Clusters(metav1.NamespaceAll),
	}
}

type ConfigSyncer struct {
	isDeploy             *bool
	clusterName          string
	clusterLoggingLister mgmtv3.ClusterLoggingLister
	projectLoggingLister mgmtv3.ProjectLoggingLister
	projectLister        mgmtv3.ProjectLister
	namespaceLister      v1.NamespaceLister
	secretManager        *SecretManager
	pods                 corev1.PodInterface
	dialer               dialer.Factory
	restConfig           *rest.Config
	clusters             mgmtv3.ClusterInterface
}

func (s *ConfigSyncer) NamespaceSync(key string, obj *k8scorev1.Namespace) (runtime.Object, error) {
	return obj, s.sync()
}

func (s *ConfigSyncer) ClusterLoggingSync(key string, obj *mgmtv3.ClusterLogging) (runtime.Object, error) {
	return obj, s.sync()
}

func (s *ConfigSyncer) ProjectLoggingSync(key string, obj *mgmtv3.ProjectLogging) (runtime.Object, error) {
	return obj, s.sync()
}

func (s *ConfigSyncer) sync() error {

	if *s.isDeploy == false {
		return nil
	}

	clusterLoggings, err := s.clusterLoggingLister.List("", labels.NewSelector())
	if err != nil {
		return errors.Wrapf(err, "List cluster loggings failed")
	}

	allProjectLoggings, err := s.projectLoggingLister.List("", labels.NewSelector())
	if err != nil {
		return errors.Wrapf(err, "List project logging failed")
	}

	var projectLoggings []*mgmtv3.ProjectLogging
	for _, logging := range allProjectLoggings {
		if controller.ObjectInCluster(s.clusterName, logging) {
			projectLoggings = append(projectLoggings, logging)
		}
	}

	sort.Slice(projectLoggings, func(i, j int) bool {
		return projectLoggings[i].Name < projectLoggings[j].Name
	})

	systemProjectID, err := s.getSystemProjectID()
	if err != nil {
		return err
	}

	if err = s.syncSSLCert(clusterLoggings, projectLoggings); err != nil {
		return err
	}

	if err = s.syncClusterConfig(clusterLoggings, systemProjectID); err != nil {
		return err
	}

	return s.syncProjectConfig(projectLoggings, systemProjectID)
}

func (s *ConfigSyncer) getSystemProjectID() (string, error) {
	projects, err := s.projectLister.List(s.clusterName, labels.Set(project.SystemProjectLabel).AsSelector())
	if err != nil {
		return "", errors.Wrapf(err, "list project failed")
	}

	var systemProject *mgmtv3.Project
	for _, v := range projects {
		if v.Spec.DisplayName == project.System {
			systemProject = v
		}
	}

	if systemProject == nil {
		return "", nil
	}

	systemProjectID := fmt.Sprintf("%s:%s", systemProject.Namespace, systemProject.Name)
	return systemProjectID, nil
}

func (s *ConfigSyncer) addExcludeNamespaces(systemProjectID string) (string, error) {
	namespaces, err := s.namespaceLister.List(metav1.NamespaceAll, labels.NewSelector())
	if err != nil {
		return "", errors.Wrapf(err, "list namespace failed")
	}

	var systemNamespaces []string
	for _, v := range namespaces {
		if v.Annotations[project.ProjectIDAnn] == systemProjectID {
			systemNamespaces = append(systemNamespaces, v.Name)
		}
	}

	return strings.Join(systemNamespaces, "|"), nil
}

func (s *ConfigSyncer) syncClusterConfig(clusterLoggings []*mgmtv3.ClusterLogging, systemProjectID string) error {
	secretName := loggingconfig.RancherLoggingConfigSecretName()
	namespace := loggingconfig.LoggingNamespace

	if len(clusterLoggings) == 0 {
		data := map[string][]byte{
			loggingconfig.LoggingSecretClusterConfigKey: []byte{},
		}
		return s.secretManager.updateSecret(secretName, namespace, data)
	}

	var excludeNamespaces string
	clusterLogging := clusterLoggings[0]
	if clusterLogging.Spec.ExcludeSystemComponent {
		var err error
		if excludeNamespaces, err = s.addExcludeNamespaces(systemProjectID); err != nil {
			return err
		}
	}

	buf, err := generator.GenerateClusterConfig(clusterLogging.Spec, excludeNamespaces)
	if err != nil {
		return err
	}

	data := map[string][]byte{
		loggingconfig.LoggingSecretClusterConfigKey: buf,
	}

	return s.secretManager.updateSecret(secretName, namespace, data)
}

func (s *ConfigSyncer) syncProjectConfig(projectLoggings []*mgmtv3.ProjectLogging, systemProjectID string) error {
	secretName := loggingconfig.RancherLoggingConfigSecretName()
	namespace := loggingconfig.LoggingNamespace

	if len(projectLoggings) == 0 {
		data := map[string][]byte{
			loggingconfig.LoggingSecretProjectConfigKey: []byte{},
		}
		return s.secretManager.updateSecret(secretName, namespace, data)
	}

	namespaces, err := s.namespaceLister.List(metav1.NamespaceAll, labels.NewSelector())
	if err != nil {
		return errors.Wrap(err, "list namespace failed")
	}

	buf, err := generator.GenerateProjectConfig(projectLoggings, namespaces, systemProjectID)
	if err != nil {
		return err
	}

	data := map[string][]byte{
		loggingconfig.LoggingSecretProjectConfigKey: buf,
	}

	return s.secretManager.updateSecret(secretName, namespace, data)
}

func (s *ConfigSyncer) syncSSLCert(clusterLoggings []*mgmtv3.ClusterLogging, projectLoggings []*mgmtv3.ProjectLogging) error {
	secretname := loggingconfig.RancherLoggingConfigSecretName()
	namespace := loggingconfig.LoggingNamespace

	sslConfig := make(map[string][]byte)
	for _, v := range clusterLoggings {
		ca, cert, key := getSSLConfig(v.Spec.ElasticsearchConfig, v.Spec.SplunkConfig, v.Spec.KafkaConfig, v.Spec.SyslogConfig, v.Spec.FluentForwarderConfig)
		sslConfig[loggingconfig.SecretDataKeyCa(loggingconfig.ClusterLevel, v.Namespace)] = []byte(ca)
		sslConfig[loggingconfig.SecretDataKeyCert(loggingconfig.ClusterLevel, v.Namespace)] = []byte(cert)
		sslConfig[loggingconfig.SecretDataKeyCertKey(loggingconfig.ClusterLevel, v.Namespace)] = []byte(key)
	}

	for _, v := range projectLoggings {
		ca, cert, key := getSSLConfig(v.Spec.ElasticsearchConfig, v.Spec.SplunkConfig, v.Spec.KafkaConfig, v.Spec.SyslogConfig, v.Spec.FluentForwarderConfig)
		projectKey := strings.Replace(v.Spec.ProjectName, ":", "_", -1)
		sslConfig[loggingconfig.SecretDataKeyCa(loggingconfig.ProjectLevel, projectKey)] = []byte(ca)
		sslConfig[loggingconfig.SecretDataKeyCert(loggingconfig.ProjectLevel, projectKey)] = []byte(cert)
		sslConfig[loggingconfig.SecretDataKeyCertKey(loggingconfig.ProjectLevel, projectKey)] = []byte(key)
	}

	return s.secretManager.updateSecret(secretname, namespace, sslConfig)
}

func getSSLConfig(esConfig *mgmtv3.ElasticsearchConfig, spConfig *mgmtv3.SplunkConfig, kfConfig *mgmtv3.KafkaConfig, syslogConfig *mgmtv3.SyslogConfig, fluentForwarder *mgmtv3.FluentForwarderConfig) (string, string, string) {
	var certificate, clientCert, clientKey string
	if esConfig != nil {
		certificate = esConfig.Certificate
		clientCert = esConfig.ClientCert
		clientKey = esConfig.ClientKey
	} else if spConfig != nil {
		certificate = spConfig.Certificate
		clientCert = spConfig.ClientCert
		clientKey = spConfig.ClientKey
	} else if kfConfig != nil {
		certificate = kfConfig.Certificate
		clientCert = kfConfig.ClientCert
		clientKey = kfConfig.ClientKey
	} else if syslogConfig != nil {
		certificate = syslogConfig.Certificate
		clientCert = syslogConfig.ClientCert
		clientKey = syslogConfig.ClientKey
	} else if fluentForwarder != nil {
		certificate = fluentForwarder.Certificate
	}

	return certificate, clientCert, clientKey
}

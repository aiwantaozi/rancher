package configsyncer

import (
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"

	"github.com/rancher/norman/controller"
	loggingconfig "github.com/rancher/rancher/pkg/controllers/user/logging/config"
	"github.com/rancher/rancher/pkg/controllers/user/logging/generator"
	"github.com/rancher/rancher/pkg/controllers/user/logging/utils"
	"github.com/rancher/rancher/pkg/project"
	"github.com/rancher/types/apis/core/v1"
	mgmtv3 "github.com/rancher/types/apis/management.cattle.io/v3"
	"github.com/rancher/types/config"

	"github.com/pkg/errors"
	k8scorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

// This controller is responsible for generate fluentd config
// and updating the config secret
// so the config reload could detect the file change and reload

const (
	clusterLevel = "cluster"
	projectLevel = "project"
)

func NewConfigSyncer(cluster *config.UserContext, SecretManager *SecretManager, isDeploy *bool) *ConfigSyncer {
	clusterName := cluster.ClusterName
	return &ConfigSyncer{
		isDeploy:             isDeploy,
		clusterName:          clusterName,
		clusterLoggingLister: cluster.Management.Management.ClusterLoggings(clusterName).Controller().Lister(),
		projectLoggingLister: cluster.Management.Management.ProjectLoggings(metav1.NamespaceAll).Controller().Lister(),
		projectLister:        cluster.Management.Management.Projects(clusterName).Controller().Lister(),
		namespaceLister:      cluster.Core.Namespaces(metav1.NamespaceAll).Controller().Lister(),
		secretManager:        SecretManager,
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
}

func (d *ConfigSyncer) NamespaceSync(key string, obj *k8scorev1.Namespace) (runtime.Object, error) {
	return obj, d.sync()
}

func (d *ConfigSyncer) ClusterLoggingSync(key string, obj *mgmtv3.ClusterLogging) (runtime.Object, error) {
	return obj, d.sync()
}

func (d *ConfigSyncer) ProjectLoggingSync(key string, obj *mgmtv3.ProjectLogging) (runtime.Object, error) {
	return obj, d.sync()
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
	clusterSecretName := loggingconfig.ClusterLoggingName
	namespace := loggingconfig.LoggingNamespace

	if len(clusterLoggings) == 0 {
		data := map[string][]byte{
			clusterLevel + ".conf": []byte{},
		}
		return s.secretManager.updateSecret(clusterSecretName, namespace, data)
	}

	var excludeNamespaces string
	clusterLogging := clusterLoggings[0]
	if clusterLogging.Spec.ExcludeSystemComponent {
		var err error
		if excludeNamespaces, err = s.addExcludeNamespaces(systemProjectID); err != nil {
			return err
		}
	}

	wl, err := utils.NewWrapClusterLogging(clusterLogging.Spec, excludeNamespaces)
	if err != nil {
		return errors.Wrap(err, "to wraper cluster logging failed")
	}

	conf := map[string]interface{}{
		"clusterTarget": wl,
		"clusterName":   s.clusterName,
	}

	configPath := loggingconfig.ClusterConfigPath
	err = generator.GenerateConfigFile(configPath, generator.ClusterTemplate, clusterLevel, conf)
	if err != nil {
		return errors.Wrap(err, "generate cluster config file failed")
	}

	file, err := os.Open(configPath)
	if err != nil {
		return errors.Wrapf(err, "find %s logging configuration file %s failed", clusterLevel, configPath)
	}

	defer file.Close()

	buf, err := ioutil.ReadAll(file)
	if err != nil {
		return errors.Wrapf(err, "read %s logging configuration file %s failed", clusterLevel, configPath)
	}

	data := map[string][]byte{
		clusterLevel + ".conf": buf,
	}

	return s.secretManager.updateSecret(clusterSecretName, namespace, data)
}

func (s *ConfigSyncer) syncProjectConfig(projectLoggings []*mgmtv3.ProjectLogging, systemProjectID string) error {
	projectSecretName := loggingconfig.ProjectLoggingName
	namespace := loggingconfig.LoggingNamespace

	if len(projectLoggings) == 0 {
		data := map[string][]byte{
			projectLevel + ".conf": []byte{},
		}
		return s.secretManager.updateSecret(projectSecretName, namespace, data)
	}

	namespaces, err := s.namespaceLister.List(metav1.NamespaceAll, labels.NewSelector())
	if err != nil {
		return errors.Wrap(err, "list namespace failed")
	}

	var wl []utils.ProjectLoggingTemplateWrap
	for _, v := range projectLoggings {

		var grepNamespace []string
		for _, v2 := range namespaces {
			if nsProjectName, ok := v2.Annotations[project.ProjectIDAnn]; ok && nsProjectName == v.Spec.ProjectName {
				grepNamespace = append(grepNamespace, v2.Name)
			}
		}

		if len(grepNamespace) == 0 {
			continue
		}

		formatgrepNamespace := fmt.Sprintf("(%s)", strings.Join(grepNamespace, "|"))
		isSystemProject := v.Spec.ProjectName == systemProjectID
		wpl, err := utils.NewWrapProjectLogging(v.Spec, formatgrepNamespace, isSystemProject)
		if err != nil {
			return err
		}

		if wpl == nil {
			continue
		}

		wl = append(wl, *wpl)
	}

	conf := map[string]interface{}{
		"projectTargets": wl,
	}
	configPath := loggingconfig.ProjectConfigPath
	err = generator.GenerateConfigFile(configPath, generator.ProjectTemplate, projectLevel, conf)
	if err != nil {
		return errors.Wrap(err, "generate project config file failed")
	}

	file, err := os.Open(configPath)
	if err != nil {
		return errors.Wrapf(err, "find %s logging configuration file %s failed", projectLevel, configPath)
	}
	defer file.Close()
	buf, err := ioutil.ReadAll(file)
	if err != nil {
		return errors.Wrapf(err, "read %s logging configuration file %s failed", projectLevel, configPath)
	}

	data := map[string][]byte{
		projectLevel + ".conf": buf,
	}

	return s.secretManager.updateSecret(projectSecretName, namespace, data)
}

func (s *ConfigSyncer) syncSSLCert(clusterLoggings []*mgmtv3.ClusterLogging, projectLoggings []*mgmtv3.ProjectLogging) error {
	name := loggingconfig.SSLSecretName
	namespace := loggingconfig.LoggingNamespace

	sslConfig := make(map[string][]byte)
	for _, v := range clusterLoggings {
		ca, cert, key := getSSLConfig(v.Spec.ElasticsearchConfig, v.Spec.SplunkConfig, v.Spec.KafkaConfig, v.Spec.SyslogConfig, v.Spec.FluentForwarderConfig)
		sslConfig[loggingconfig.SecretDataKeyCa(clusterLevel, v.Namespace)] = []byte(ca)
		sslConfig[loggingconfig.SecretDataKeyCert(clusterLevel, v.Namespace)] = []byte(cert)
		sslConfig[loggingconfig.SecretDataKeyCertKey(clusterLevel, v.Namespace)] = []byte(key)
	}

	for _, v := range projectLoggings {
		ca, cert, key := getSSLConfig(v.Spec.ElasticsearchConfig, v.Spec.SplunkConfig, v.Spec.KafkaConfig, v.Spec.SyslogConfig, v.Spec.FluentForwarderConfig)
		projectKey := strings.Replace(v.Spec.ProjectName, ":", "_", -1)
		sslConfig[loggingconfig.SecretDataKeyCa(projectLevel, projectKey)] = []byte(ca)
		sslConfig[loggingconfig.SecretDataKeyCert(projectLevel, projectKey)] = []byte(cert)
		sslConfig[loggingconfig.SecretDataKeyCertKey(projectLevel, projectKey)] = []byte(key)
	}

	return s.secretManager.updateSecret(name, namespace, sslConfig)
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

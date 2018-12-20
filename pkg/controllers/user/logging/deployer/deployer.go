package deployer

import (
	"strconv"

	"github.com/rancher/norman/controller"
	loggingconfig "github.com/rancher/rancher/pkg/controllers/user/logging/config"
	"github.com/rancher/rancher/pkg/controllers/user/logging/configsyncer"
	"github.com/rancher/rancher/pkg/controllers/user/logging/utils"
	"github.com/rancher/rancher/pkg/image"
	"github.com/rancher/types/apis/apps/v1beta2"
	"github.com/rancher/types/apis/core/v1"
	mgmtv3 "github.com/rancher/types/apis/management.cattle.io/v3"
	rbacv1 "github.com/rancher/types/apis/rbac.authorization.k8s.io/v1"
	"github.com/rancher/types/config"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	k8sappsv1beta2 "k8s.io/api/apps/v1beta2"
	k8scorev1 "k8s.io/api/core/v1"
	k8srbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// This controller is responsible for deploy fluend
// and log aggregator

type Deployer struct {
	isDeploy             *bool
	clusterName          string
	clusterLister        mgmtv3.ClusterLister
	clusterLoggingLister mgmtv3.ClusterLoggingLister
	projectLoggingLister mgmtv3.ProjectLoggingLister
	loggingDeployer      *loggingDeployer
}

type loggingDeployer struct {
	daemonsets          v1beta2.DaemonSetInterface
	clusterRolebindings rbacv1.ClusterRoleBindingInterface
	services            v1.ServiceInterface
	serviceAccounts     v1.ServiceAccountInterface
	namespaces          v1.NamespaceInterface
	secretSyncer        *configsyncer.SecretManager
}

func NewDeployer(cluster *config.UserContext, secretSyncer *configsyncer.SecretManager, isDeploy *bool) *Deployer {
	namespace := loggingconfig.LoggingNamespace
	clusterName := cluster.ClusterName

	loggingDeployer := &loggingDeployer{
		daemonsets:          cluster.Apps.DaemonSets(namespace),
		clusterRolebindings: cluster.RBAC.ClusterRoleBindings(namespace),
		services:            cluster.Core.Services(namespace),
		serviceAccounts:     cluster.Core.ServiceAccounts(namespace),
		namespaces:          cluster.Core.Namespaces(metav1.NamespaceAll),
		secretSyncer:        secretSyncer,
	}

	return &Deployer{
		isDeploy:             isDeploy,
		clusterName:          clusterName,
		clusterLister:        cluster.Management.Management.Clusters(metav1.NamespaceAll).Controller().Lister(),
		clusterLoggingLister: cluster.Management.Management.ClusterLoggings(clusterName).Controller().Lister(),
		projectLoggingLister: cluster.Management.Management.ProjectLoggings(metav1.NamespaceAll).Controller().Lister(),
		loggingDeployer:      loggingDeployer,
	}
}

func (d *Deployer) ClusterLoggingSync(key string, obj *mgmtv3.ClusterLogging) (runtime.Object, error) {
	return obj, d.sync()
}

func (d *Deployer) ProjectLoggingSync(key string, obj *mgmtv3.ProjectLogging) (runtime.Object, error) {
	return obj, d.sync()
}

func (d *Deployer) sync() error {
	allDisabled, err := d.isAllLoggingDisable()
	if err != nil {
		return err
	}

	if allDisabled {
		if err := d.cleanResource(); err != nil {
			return err
		}
		*d.isDeploy = false
		return nil
	}

	if err := d.deploy(); err != nil {
		return err
	}

	*d.isDeploy = true
	return nil
}

func (d *Deployer) deploy() error {
	if err := d.loggingDeployer.initeNamespace(); err != nil {
		return err
	}

	if err := d.loggingDeployer.secretSyncer.InitLoggingSecret(); err != nil {
		return err
	}

	cluster, err := d.clusterLister.Get("", d.clusterName)
	if err != nil {
		return errors.Wrapf(err, "get dockerRootDir from cluster %s failed", d.clusterName)
	}

	if err := d.loggingDeployer.createLogAggregator(cluster.Status.Driver); err != nil {
		return err
	}

	return d.loggingDeployer.createFluentd(cluster.Spec.DockerRootDir)
}

func (d *Deployer) isAllLoggingDisable() (bool, error) {
	clusterLoggings, err := d.clusterLoggingLister.List("", labels.NewSelector())
	if err != nil {
		return false, err
	}

	allClusterProjectLoggings, err := d.projectLoggingLister.List("", labels.NewSelector())
	if err != nil {
		return false, err
	}

	var projectLoggings []*mgmtv3.ProjectLogging
	for _, v := range allClusterProjectLoggings {
		if controller.ObjectInCluster(d.clusterName, v) {
			projectLoggings = append(projectLoggings, v)
		}
	}

	if len(clusterLoggings) == 0 && len(projectLoggings) == 0 {
		return true, nil
	}

	for _, v := range clusterLoggings {
		wl := utils.NewLoggingTargetTestWrap(v.Spec.ElasticsearchConfig, v.Spec.SplunkConfig, v.Spec.SyslogConfig, v.Spec.KafkaConfig, v.Spec.FluentForwarderConfig)
		if wl != nil {
			return false, nil
		}
	}

	for _, v := range projectLoggings {
		wpl := utils.NewLoggingTargetTestWrap(v.Spec.ElasticsearchConfig, v.Spec.SplunkConfig, v.Spec.SyslogConfig, v.Spec.KafkaConfig, v.Spec.FluentForwarderConfig)
		if wpl != nil {
			return false, nil
		}
	}
	return true, nil
}

func (d *Deployer) cleanResource() error {
	var zero int64
	name := loggingconfig.LoggingNamespace
	foreground := metav1.DeletePropagationForeground

	if err := d.loggingDeployer.namespaces.Delete(name, &metav1.DeleteOptions{GracePeriodSeconds: &zero, PropagationPolicy: &foreground}); err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	return nil
}

func (d *loggingDeployer) initeNamespace() error {
	name := loggingconfig.LoggingNamespace
	if _, err := d.namespaces.Controller().Lister().Get(metav1.NamespaceAll, name); err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}

		initNamespace := k8scorev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
		}

		if _, err := d.namespaces.Create(&initNamespace); err != nil && !apierrors.IsAlreadyExists(err) {
			return err
		}
	}
	return nil
}

func (d *loggingDeployer) createLogAggregator(driverName string) (err error) {
	name := loggingconfig.LogAggregatorName
	namespace := loggingconfig.LoggingNamespace

	defer func() {
		if err != nil && !apierrors.IsAlreadyExists(err) {
			if err = d.removeDeamonset(name); err != nil {
				logrus.Error("recycle log-aggregator daemonset failed", err)
			}
		}
	}()

	serviceAccount := newServiceAccount(name, namespace)
	_, err = d.serviceAccounts.Create(serviceAccount)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}

	rolebinding := newRoleBinding(name, namespace)
	_, err = d.clusterRolebindings.Create(rolebinding)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}

	driverDir := getDriverDir(driverName)
	daemonset := NewLogAggregatorDaemonset(driverDir)
	_, err = d.daemonsets.Create(daemonset)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}
	return nil
}

func (d *loggingDeployer) createFluentd(dockerRootDir string) (err error) {
	name := loggingconfig.FluentdName
	namespace := loggingconfig.LoggingNamespace

	defer func() {
		if err != nil && !apierrors.IsAlreadyExists(err) {
			if err = d.removeDeamonset(name); err != nil {
				logrus.Errorf("recycle %s failed, %v", name, err)
			}
		}
	}()

	serviceAccount := newServiceAccount(name, namespace)
	_, err = d.serviceAccounts.Create(serviceAccount)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}

	roleBind := newRoleBinding(name, namespace)
	_, err = d.clusterRolebindings.Create(roleBind)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}

	daemonset := NewFluentdDaemonset(dockerRootDir)
	_, err = d.daemonsets.Create(daemonset)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}

	service := newService(name, namespace)
	_, err = d.services.Create(service)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}
	return nil
}

func (d *loggingDeployer) removeDeamonset(name string) error {
	deleteOp := metav1.DeletePropagationBackground
	var errgrp errgroup.Group
	errgrp.Go(func() error {
		return d.daemonsets.Delete(name, &metav1.DeleteOptions{PropagationPolicy: &deleteOp})
	})

	errgrp.Go(func() error {
		return d.serviceAccounts.Delete(name, &metav1.DeleteOptions{})
	})

	errgrp.Go(func() error {
		return d.clusterRolebindings.Delete(name, &metav1.DeleteOptions{})
	})

	if err := errgrp.Wait(); err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	return nil
}

func newService(name, namespace string) *k8scorev1.Service {
	return &k8scorev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
			Labels: map[string]string{
				loggingconfig.LabelK8sApp: loggingconfig.FluentdName,
			},
		},
		Spec: k8scorev1.ServiceSpec{
			Selector: map[string]string{
				loggingconfig.LabelK8sApp: loggingconfig.FluentdName,
			},
			Ports: []k8scorev1.ServicePort{
				{
					Name:       "metrics",
					Port:       24231,
					TargetPort: intstr.Parse(strconv.FormatInt(24231, 10)),
					Protocol:   k8scorev1.Protocol(k8scorev1.ProtocolTCP),
				},
			},
		},
	}
}

func newServiceAccount(name, namespace string) *k8scorev1.ServiceAccount {
	return &k8scorev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
}

func newRoleBinding(name, namespace string) *k8srbacv1.ClusterRoleBinding {
	return &k8srbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Subjects: []k8srbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      name,
				Namespace: namespace,
			},
		},
		RoleRef: k8srbacv1.RoleRef{
			Kind:     "ClusterRole",
			Name:     "cluster-admin",
			APIGroup: "rbac.authorization.k8s.io",
		},
	}
}

func NewFluentdDaemonset(dockerRootDir string) *k8sappsv1beta2.DaemonSet {
	name := loggingconfig.FluentdName
	helperName := loggingconfig.FluentdHelperName
	namespace := loggingconfig.LoggingNamespace
	clusterVolumeName := loggingconfig.ClusterLoggingName
	projectVolumeName := loggingconfig.ProjectLoggingName
	sslSecretVolumeName := loggingconfig.SSLSecretName

	privileged := true
	terminationGracePeriodSeconds := int64(30)

	if dockerRootDir == "" {
		dockerRootDir = "/var/lib/docker"
	}
	dockerRootContainers := dockerRootDir + "/containers"

	logVolMounts, logVols := buildHostPathVolumes([]string{"varlibdockercontainers", "varlogcontainers", "varlogpods", "rkelog", "customlog", "fluentdlog"},
		map[string][]string{
			"varlibdockercontainers": []string{dockerRootContainers, dockerRootContainers},
			"varlogcontainers":       []string{"/var/log/containers", "/var/log/containers"},
			"varlogpods":             []string{"/var/log/pods", "/var/log/pods"},
			"rkelog":                 []string{"/var/lib/rancher/rke/log", "/var/lib/rancher/rke/log"},
			"customlog":              []string{"/var/lib/rancher/log-volumes", "/var/lib/rancher/log-volumes"},
			"fluentdlog":             []string{"/fluentd/log", "/var/lib/rancher/fluentd/log"},
		})

	configVolMounts, configVols := buildSecretVolumes([]string{clusterVolumeName, projectVolumeName},
		map[string][]string{
			clusterVolumeName: []string{"/fluentd/etc/config/cluster", clusterVolumeName},
			projectVolumeName: []string{"/fluentd/etc/config/project", projectVolumeName},
		})

	customConfigVolMounts, customConfigVols := buildHostPathVolumes([]string{"clustercustomlogconfig", "projectcustomlogconfig"},
		map[string][]string{
			"clustercustomlogconfig": []string{"/fluentd/etc/config/custom/cluster", "/var/lib/rancher/fluentd/etc/config/custom/cluster"},
			"projectcustomlogconfig": []string{"/fluentd/etc/config/custom/project", "/var/lib/rancher/fluentd/etc/config/custom/project"},
		})

	sslVolMounts, sslVols := buildSecretVolumes([]string{sslSecretVolumeName},
		map[string][]string{
			sslSecretVolumeName: []string{"/fluentd/etc/ssl", sslSecretVolumeName},
		})

	allConfigVolMounts, allConfigVols := append(configVolMounts, customConfigVolMounts...), append(configVols, customConfigVols...)
	allVolMounts, allVols := append(append(allConfigVolMounts, logVolMounts...), sslVolMounts...), append(append(allConfigVols, logVols...), sslVols...)

	return &k8sappsv1beta2.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				loggingconfig.LabelK8sApp: name,
			},
		},
		Spec: k8sappsv1beta2.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					loggingconfig.LabelK8sApp: name,
				},
			},
			Template: k8scorev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
					Labels: map[string]string{
						loggingconfig.LabelK8sApp: name,
					},
				},
				Spec: k8scorev1.PodSpec{
					Tolerations: []k8scorev1.Toleration{
						{
							Key:    "node-role.kubernetes.io/master",
							Effect: k8scorev1.TaintEffectNoSchedule,
						},
						{
							Key:    "node-role.kubernetes.io/etcd",
							Value:  "true",
							Effect: k8scorev1.TaintEffectNoExecute,
						},
						{
							Key:    "node-role.kubernetes.io/controlplane",
							Value:  "true",
							Effect: k8scorev1.TaintEffectNoSchedule,
						},
					},
					Containers: []k8scorev1.Container{
						{
							Name:            name,
							Image:           image.Resolve(mgmtv3.ToolsSystemImages.LoggingSystemImages.Fluentd),
							ImagePullPolicy: k8scorev1.PullIfNotPresent,
							Command:         []string{"fluentd"},
							Args:            []string{"-c", "/fluentd/etc/fluent.conf"},
							VolumeMounts:    allVolMounts,
							SecurityContext: &k8scorev1.SecurityContext{
								Privileged: &privileged,
							},
						},
						{
							Name:    helperName,
							Image:   image.Resolve(mgmtv3.ToolsSystemImages.LoggingSystemImages.FluentdHelper),
							Command: []string{"fluentd-helper"},
							Args: []string{
								"--watched-file-list", "/fluentd/etc/config/cluster", "--watched-file-list", "/fluentd/etc/config/project",
								"--watched-file-list", "/fluentd/etc/config/custom/cluster", "--watched-file-list", "/fluentd/etc/config/custom/project",
								"--watched-file-list", "/fluentd/etc/ssl",
							},
							ImagePullPolicy: k8scorev1.PullAlways,
							SecurityContext: &k8scorev1.SecurityContext{
								Privileged: &privileged,
							},
							VolumeMounts: append(allConfigVolMounts, sslVolMounts...),
						},
					},
					ServiceAccountName:            name,
					TerminationGracePeriodSeconds: &terminationGracePeriodSeconds,
					Volumes: allVols,
				},
			},
		},
	}
}

func NewLogAggregatorDaemonset(driverDir string) *k8sappsv1beta2.DaemonSet {
	name := loggingconfig.LogAggregatorName
	namespace := loggingconfig.LoggingNamespace
	privileged := true

	return &k8sappsv1beta2.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				loggingconfig.LabelK8sApp: loggingconfig.LogAggregatorName,
			},
		},
		Spec: k8sappsv1beta2.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					loggingconfig.LabelK8sApp: loggingconfig.LogAggregatorName,
				},
			},
			Template: k8scorev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
					Labels: map[string]string{
						loggingconfig.LabelK8sApp: loggingconfig.LogAggregatorName,
					},
				},
				Spec: k8scorev1.PodSpec{
					Tolerations: []k8scorev1.Toleration{
						{
							Key:    "node-role.kubernetes.io/master",
							Effect: k8scorev1.TaintEffectNoSchedule,
						},
						{
							Key:    "node-role.kubernetes.io/etcd",
							Value:  "true",
							Effect: k8scorev1.TaintEffectNoExecute,
						},
						{
							Key:    "node-role.kubernetes.io/controlplane",
							Value:  "true",
							Effect: k8scorev1.TaintEffectNoSchedule,
						},
					},
					Containers: []k8scorev1.Container{
						{
							Name:            loggingconfig.LogAggregatorName,
							Image:           image.Resolve(mgmtv3.ToolsSystemImages.LoggingSystemImages.LogAggregatorFlexVolumeDriver),
							ImagePullPolicy: k8scorev1.PullAlways,
							SecurityContext: &k8scorev1.SecurityContext{
								Privileged: &privileged,
							},
							VolumeMounts: []k8scorev1.VolumeMount{
								{
									Name:      "flexvolume-driver",
									MountPath: "/flexmnt",
								},
							},
						},
					},
					ServiceAccountName: loggingconfig.LogAggregatorName,
					Volumes: []k8scorev1.Volume{
						{
							Name: "flexvolume-driver",
							VolumeSource: k8scorev1.VolumeSource{
								HostPath: &k8scorev1.HostPathVolumeSource{
									Path: driverDir,
								},
							},
						},
					},
				},
			},
		},
	}
}

func getDriverDir(driverName string) string {
	switch driverName {
	case mgmtv3.ClusterDriverRKE:
		return "/var/lib/kubelet/volumeplugins"
	case loggingconfig.GoogleKubernetesEngine:
		return "/home/kubernetes/flexvolume"
	default:
		return "/usr/libexec/kubernetes/kubelet-plugins/volume/exec"
	}
}

func buildHostPathVolumes(keys []string, mounts map[string][]string) (vms []k8scorev1.VolumeMount, vs []k8scorev1.Volume) {
	for _, name := range keys {
		value, ok := mounts[name]
		if !ok {
			continue
		}
		vms = append(vms, k8scorev1.VolumeMount{
			Name:      name,
			MountPath: value[0],
		})
		vs = append(vs, k8scorev1.Volume{
			Name: name,
			VolumeSource: k8scorev1.VolumeSource{
				HostPath: &k8scorev1.HostPathVolumeSource{
					Path: value[1],
				},
			},
		})
	}
	return
}

func buildSecretVolumes(keys []string, mounts map[string][]string) (vms []k8scorev1.VolumeMount, vs []k8scorev1.Volume) {
	for _, name := range keys {
		value, ok := mounts[name]
		if !ok {
			continue
		}
		vms = append(vms, k8scorev1.VolumeMount{
			Name:      name,
			MountPath: value[0],
		})
		vs = append(vs, k8scorev1.Volume{
			Name: name,
			VolumeSource: k8scorev1.VolumeSource{
				Secret: &k8scorev1.SecretVolumeSource{
					SecretName: value[1],
				},
			},
		})
	}
	return
}

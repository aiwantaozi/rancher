package alert

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/rancher/rancher/pkg/controllers/user/alert/configsyncer"
	"github.com/rancher/rancher/pkg/controllers/user/alert/manager"
	"github.com/rancher/rancher/pkg/controllers/user/alert/watcher"
	"github.com/rancher/types/apis/management.cattle.io/v3"
	"github.com/rancher/types/config"
	"github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	SeverityInfo       = "info"
	SeverityCritical   = "critical"
	SeverityWarning    = "warning"
	defaultTimingField = v3.TimingField{
		GroupWaitSeconds:      180,
		GroupIntervalSeconds:  180,
		RepeatIntervalSeconds: 3600,
	}
)

var (
	ComparisonEqual          = "equal"
	ComparisonNotEqual       = "not-equal"
	ComparisonGreaterThan    = "greater-than"
	ComparisonLessThan       = "less-than"
	ComparisonGreaterOrEqual = "greater-or-equal"
	ComparisonLessOrEqual    = "less-or-equal"
)

func Register(ctx context.Context, cluster *config.UserContext) {
	alertmanager := manager.NewAlertManager(cluster)
	prometheusCRDManager := manager.NewPrometheusCRDManager(ctx, cluster)
	clusterAlertRules := cluster.Management.Management.ClusterAlertRules(cluster.ClusterName)
	projectAlertRules := cluster.Management.Management.ProjectAlertRules("")

	clusterAlertGroups := cluster.Management.Management.ClusterAlertGroups(cluster.ClusterName)
	projectAlertGroups := cluster.Management.Management.ProjectAlertGroups("")
	// deployer := deployer.NewDeployer(cluster, alertmanager)
	// cluste·manager, prometheusCRDManager)

	configSyncer := configsyncer.NewConfigSyncer(ctx, cluster, alertmanager, prometheusCRDManager)
	clusterAlertRules.AddClusterScopedHandler(ctx, "cluster-alert-group-controller", cluster.ClusterName, configSyncer.ClusterSync)
	projectAlertRules.AddClusterScopedHandler(ctx, "project-alert-group-controller", cluster.ClusterName, configSyncer.ProjectSync)

	// clusterAlertRules := cluster.Management.Management.ClusterAlertRules(cluster.ClusterName)
	// projectAlertRules := cluster.Management.Management.ProjectAlertRules("")

	// clusterAlertRules.AddClusterScopedHandler("cluster-alert-rule-controller", cluster.ClusterName, configSyncer.ClusterSync)
	// projectAlertRulesclusterAlertRules.AddClusterScopedHandler("project-alert-rule-controller", cluster.ClusterName, configSyncer.ProjectSync)

	projects := cluster.Management.Management.Projects("")
	projectLifecycle := &ProjectLifecycle{
		projectAlertRules:  projectAlertRules,
		projectAlertGroups: projectAlertGroups,
		clusterName:        cluster.ClusterName,
	}
	projects.AddClusterScopedLifecycle(ctx, "project-precan-alertpoicy-controller", cluster.ClusterName, projectLifecycle)
	initClusterPreCanAlerts(clusterAlertGroups, clusterAlertRules, cluster.ClusterName)

	watcher.StartEventWatcher(ctx, cluster, alertmanager)
	watcher.StartSysComponentWatcher(ctx, cluster, alertmanager)
	watcher.StartPodWatcher(ctx, cluster, alertmanager)
	watcher.StartWorkloadWatcher(ctx, cluster, alertmanager)
	watcher.StartNodeWatcher(ctx, cluster, alertmanager)
}

func initClusterPreCanAlerts(clusterAlertGroups v3.ClusterAlertGroupInterface, clusterAlertRules v3.ClusterAlertRuleInterface, clusterName string) {
	name := "etcd-alert"
	group := &v3.ClusterAlertGroup{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v3.ClusterGroupSpec{
			ClusterName: clusterName,
			CommonGroupField: v3.CommonGroupField{
				Description: "Alert for etcd component, leader existence, db size",
				DisplayName: "Alert for etcd",
				TimingField: defaultTimingField,
			},
		},
		Status: v3.Status{
			State: "active",
		},
	}

	if _, err := clusterAlertGroups.Create(group); err != nil && !apierrors.IsAlreadyExists(err) {
		logrus.Warnf("Failed to create precan rules for etcd: %v", err)
	}

	name = "no-leader"
	rule := &v3.ClusterAlertRule{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v3.ClusterAlertRuleSpec{
			ClusterName: clusterName,
			GroupName:   configsyncer.GetGroupID(clusterName, group.Name),
			CommonRuleField: v3.CommonRuleField{
				Severity:    SeverityCritical,
				DisplayName: "no-leader",
				TimingField: defaultTimingField,
			},
			MetricRule: &v3.MetricRule{
				Description:    "Alert for etcd component, leader existence, db size",
				Expression:     `etcd_server_has_leader{job="kube-etcd"}`,
				Comparison:     ComparisonEqual,
				Duration:       "3m",
				ThresholdValue: 0,
			},
		},
		Status: v3.Status{
			State: "active",
		},
	}

	if _, err := clusterAlertRules.Create(rule); err != nil && !apierrors.IsAlreadyExists(err) {
		logrus.Warnf("Failed to create precan rules for %s: %v", name, err)
	}

	name = "high-number-of-leader-changes"
	rule = &v3.ClusterAlertRule{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v3.ClusterAlertRuleSpec{
			ClusterName: clusterName,
			GroupName:   configsyncer.GetGroupID(clusterName, group.Name),
			CommonRuleField: v3.CommonRuleField{
				Severity:    SeverityCritical,
				DisplayName: "Alert for etcd",
				TimingField: defaultTimingField,
			},
			MetricRule: &v3.MetricRule{
				Description:    "todo",
				Expression:     `increase(etcd_server_leader_changes_seen_total{job="kube-etcd"}[1h])`,
				Comparison:     ComparisonGreaterThan,
				Duration:       "3m",
				ThresholdValue: 3,
			},
		},
		Status: v3.Status{
			State: "active",
		},
	}

	if _, err := clusterAlertRules.Create(rule); err != nil && !apierrors.IsAlreadyExists(err) {
		logrus.Warnf("Failed to create precan rules for %s: %v", name, err)
	}

	name = "db-over-size"
	rule = &v3.ClusterAlertRule{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v3.ClusterAlertRuleSpec{
			ClusterName: clusterName,
			GroupName:   configsyncer.GetGroupID(clusterName, group.Name),
			CommonRuleField: v3.CommonRuleField{
				Severity:    SeverityCritical,
				DisplayName: "db-over-size",
				TimingField: defaultTimingField,
			},
			MetricRule: &v3.MetricRule{
				Description:    "todo",
				Expression:     `sum(etcd_debugging_mvcc_db_total_size_in_bytes)`,
				Comparison:     ComparisonGreaterThan,
				Duration:       "3m",
				ThresholdValue: 10000,
			},
		},
		Status: v3.Status{
			State: "active",
		},
	}

	if _, err := clusterAlertRules.Create(rule); err != nil && !apierrors.IsAlreadyExists(err) {
		logrus.Warnf("Failed to create precan rules for %s: %v", name, err)
	}

	name = "etcd-system-service"
	rule = &v3.ClusterAlertRule{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v3.ClusterAlertRuleSpec{
			ClusterName: clusterName,
			GroupName:   configsyncer.GetGroupID(clusterName, group.Name),
			CommonRuleField: v3.CommonRuleField{
				Severity:    SeverityCritical,
				DisplayName: "etcd-system-service",
				TimingField: defaultTimingField,
			},
			SystemServiceRule: &v3.SystemServiceRule{
				Condition: "etcd",
			},
		},
		Status: v3.Status{
			State: "active",
		},
	}

	if _, err := clusterAlertRules.Create(rule); err != nil && !apierrors.IsAlreadyExists(err) {
		logrus.Warnf("Failed to create precan rules for %s: %v", name, err)
	}

	name = "kube-components-alert"
	group = &v3.ClusterAlertGroup{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v3.ClusterGroupSpec{
			ClusterName: clusterName,
			CommonGroupField: v3.CommonGroupField{
				Description: "Alert for controller-manager, scheduler",
				DisplayName: "Built-in Alert for controller-manager, scheduler component",
				TimingField: defaultTimingField,
			},
		},
		Status: v3.Status{
			State: "active",
		},
	}

	if _, err := clusterAlertGroups.Create(group); err != nil && !apierrors.IsAlreadyExists(err) {
		logrus.Warnf("Failed to create precan rules for controller manager, scheduler: %v", err)
	}

	name = "scheduler-system-service"
	rule = &v3.ClusterAlertRule{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v3.ClusterAlertRuleSpec{
			ClusterName: clusterName,
			GroupName:   configsyncer.GetGroupID(clusterName, group.Name),
			CommonRuleField: v3.CommonRuleField{
				Severity:    SeverityCritical,
				DisplayName: "scheduler-system-service",
				TimingField: defaultTimingField,
			},
			SystemServiceRule: &v3.SystemServiceRule{
				Condition: "scheduler",
			},
		},
		Status: v3.Status{
			State: "active",
		},
	}

	if _, err := clusterAlertRules.Create(rule); err != nil && !apierrors.IsAlreadyExists(err) {
		logrus.Warnf("Failed to create precan rules for %s: %v", name, err)
	}

	name = "controllermanager-system-service"
	rule = &v3.ClusterAlertRule{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v3.ClusterAlertRuleSpec{
			ClusterName: clusterName,
			GroupName:   configsyncer.GetGroupID(clusterName, group.Name),
			CommonRuleField: v3.CommonRuleField{
				Severity:    SeverityCritical,
				DisplayName: "controllermanager-system-service",
				TimingField: defaultTimingField,
			},
			SystemServiceRule: &v3.SystemServiceRule{
				Condition: "controller-manager",
			},
		},
		Status: v3.Status{
			State: "active",
		},
	}

	if _, err := clusterAlertRules.Create(rule); err != nil && !apierrors.IsAlreadyExists(err) {
		logrus.Warnf("Failed to create precan rules for %s: %v", name, err)
	}

	name = "node-alert"
	group = &v3.ClusterAlertGroup{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v3.ClusterGroupSpec{
			ClusterName: clusterName,
			CommonGroupField: v3.CommonGroupField{
				Description: "Alert for Node Memory, CPU, Disk Usage",
				DisplayName: "Built-in Alert for node mem, cpu, disk usage",
				TimingField: defaultTimingField,
			},
		},
		Status: v3.Status{
			State: "active",
		},
	}

	if _, err := clusterAlertGroups.Create(group); err != nil && !apierrors.IsAlreadyExists(err) {
		logrus.Warnf("Failed to create precan rules for node: %v", err)
	}

	name = "node-disk-running-full"
	rule = &v3.ClusterAlertRule{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v3.ClusterAlertRuleSpec{
			ClusterName: clusterName,
			GroupName:   configsyncer.GetGroupID(clusterName, group.Name),
			CommonRuleField: v3.CommonRuleField{
				Severity:    SeverityCritical,
				DisplayName: name,
				TimingField: defaultTimingField,
			},
			MetricRule: &v3.MetricRule{
				Description:    "todo",
				Expression:     `predict_linear(node_filesystem_free{job="node-exporter",mountpoint!~"^/etc/(?:resolv.conf|hosts|hostname)$"}[6h], 3600 * 24) < 0 and on(instance) up{job="node-exporter"}`,
				Comparison:     ComparisonEqual, //todo
				Duration:       "3m",
				ThresholdValue: 0,
			},
		},
		Status: v3.Status{
			State: "active",
		},
	}

	if _, err := clusterAlertRules.Create(rule); err != nil && !apierrors.IsAlreadyExists(err) {
		logrus.Warnf("Failed to create precan rules for %s: %v", name, err)
	}

	name = "high-memmory"
	rule = &v3.ClusterAlertRule{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v3.ClusterAlertRuleSpec{
			ClusterName: clusterName,
			GroupName:   configsyncer.GetGroupID(clusterName, group.Name),
			CommonRuleField: v3.CommonRuleField{
				Severity:    SeverityCritical,
				DisplayName: name,
				TimingField: defaultTimingField,
			},
			MetricRule: &v3.MetricRule{
				Description:    "todo",
				Expression:     `1 - sum(node_memory_MemAvailable_bytes) by (instance) / sum(node_memory_MemTotal_bytes) by (instance)`,
				Comparison:     ComparisonLessOrEqual,
				Duration:       "3m",
				ThresholdValue: 20,
			},
		},
		Status: v3.Status{
			State: "active",
		},
	}

	if _, err := clusterAlertRules.Create(rule); err != nil && !apierrors.IsAlreadyExists(err) {
		logrus.Warnf("Failed to create precan rules for %s: %v", name, err)
	}

	name = "high-cpu-load"
	rule = &v3.ClusterAlertRule{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v3.ClusterAlertRuleSpec{
			ClusterName: clusterName,
			GroupName:   configsyncer.GetGroupID(clusterName, group.Name),
			CommonRuleField: v3.CommonRuleField{
				Severity:    SeverityCritical,
				DisplayName: name,
				TimingField: defaultTimingField,
			},
			MetricRule: &v3.MetricRule{
				Description:    "todo",
				Expression:     `sum(node_load1) by (instance) / count(node_cpu_seconds_total{mode="system"}) by (instance)`,
				Comparison:     ComparisonGreaterThan,
				Duration:       "3m",
				ThresholdValue: 1,
			},
		},
		Status: v3.Status{
			State: "active",
		},
	}

	if _, err := clusterAlertRules.Create(rule); err != nil && !apierrors.IsAlreadyExists(err) {
		logrus.Warnf("Failed to create precan rules for %s: %v", name, err)
	}

	name = "deploment-event-alert"
	group = &v3.ClusterAlertGroup{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v3.ClusterGroupSpec{
			ClusterName: clusterName,
			CommonGroupField: v3.CommonGroupField{
				Description: "Alert for Event",
				DisplayName: "Alert when event happened",
				TimingField: defaultTimingField,
			},
		},
		Status: v3.Status{
			State: "active",
		},
	}

	if _, err := clusterAlertGroups.Create(group); err != nil && !apierrors.IsAlreadyExists(err) {
		logrus.Warnf("Failed to create precan rules for event: %v", err)
	}

	name = "deploment-event-alert"
	rule = &v3.ClusterAlertRule{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v3.ClusterAlertRuleSpec{
			ClusterName: clusterName,
			GroupName:   configsyncer.GetGroupID(clusterName, group.Name),
			CommonRuleField: v3.CommonRuleField{
				Severity:    SeverityCritical,
				DisplayName: name,
				TimingField: defaultTimingField,
			},
			EventRule: &v3.EventRule{
				EventType:    "Warning",
				ResourceKind: "Deployment",
			},
		},
		Status: v3.Status{
			State: "active",
		},
	}

	if _, err := clusterAlertRules.Create(rule); err != nil && !apierrors.IsAlreadyExists(err) {
		logrus.Warnf("Failed to create precan rules for %s: %v", name, err)
	}
}

func getCommonRuleField(groupID, displayName, severity string) v3.CommonRuleField {
	return v3.CommonRuleField{
		DisplayName: displayName,
		Severity:    severity,
		TimingField: defaultTimingField,
	}
}

type ProjectLifecycle struct {
	projectAlertGroups v3.ProjectAlertGroupInterface
	projectAlertRules  v3.ProjectAlertRuleInterface
	clusterName        string
}

//Create built-in project alerts
func (l *ProjectLifecycle) Create(obj *v3.Project) (runtime.Object, error) {
	name := "projectalert-workload-alert"
	group := &v3.ProjectAlertGroup{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: obj.Namespace,
		},
		Spec: v3.ProjectGroupSpec{
			ProjectName: l.clusterName + ":" + obj.Name,
			CommonGroupField: v3.CommonGroupField{
				DisplayName: "Alert for Workload",
				Description: "Built-in Alert for Workload",
				TimingField: defaultTimingField,
			},
		},
		Status: v3.Status{
			State: "active",
		},
	}

	if _, err := l.projectAlertGroups.Create(group); err != nil && !apierrors.IsAlreadyExists(err) {
		logrus.Warnf("Failed to create built-in rules for deployment: %v", err)
	}

	name = "less-than-half-workload-available"
	rule := &v3.ProjectAlertRule{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v3.ProjectAlertRuleSpec{
			ProjectName: obj.Name, //todo
			GroupName:   configsyncer.GetGroupID(obj.Name, group.Name),
			CommonRuleField: v3.CommonRuleField{
				Severity:    SeverityCritical,
				DisplayName: name,
				TimingField: defaultTimingField,
			},
			WorkloadRule: &v3.WorkloadRule{
				Selector: map[string]string{
					"app": "workload",
				},
				AvailablePercentage: 50,
			},
		},
		Status: v3.Status{
			State: "active",
		},
	}

	if _, err := l.projectAlertRules.Create(rule); err != nil && !apierrors.IsAlreadyExists(err) {
		logrus.Warnf("Failed to create precan rules for %s: %v", name, err)
	}

	name = "memory-close-to-resource-limited"
	rule = &v3.ProjectAlertRule{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v3.ProjectAlertRuleSpec{
			ProjectName: obj.Name, //todo
			GroupName:   configsyncer.GetGroupID(obj.Name, group.Name),
			CommonRuleField: v3.CommonRuleField{
				Severity:    SeverityCritical,
				DisplayName: name,
				TimingField: defaultTimingField,
			},
			MetricRule: &v3.MetricRule{
				Expression:     `sum(container_memory_working_set_bytes{namespace="$namespace", pod_name=~"$instance"}) by ("$instance") / sum(label_join(kube_pod_container_resource_limits_memory_bytes{namespace="$namespace", pod=~"$instanceName"},"pod_name", "", "pod")) by (pod_name)`,
				Comparison:     ComparisonGreaterThan, //todo
				Duration:       "3m",
				ThresholdValue: 0.8,
			},
		},
		Status: v3.Status{
			State: "active",
		},
	}

	if _, err := l.projectAlertRules.Create(rule); err != nil && !apierrors.IsAlreadyExists(err) {
		logrus.Warnf("Failed to create precan rules for %s: %v", name, err)
	}
	return obj, nil
}

func (l *ProjectLifecycle) Updated(obj *v3.Project) (runtime.Object, error) {
	return obj, nil
}

func (l *ProjectLifecycle) Remove(obj *v3.Project) (runtime.Object, error) {
	return obj, nil
}

package clustercrd

import (
	"context"
	"fmt"
	"reflect"

	"github.com/juju/errors"
	"github.com/rancher/norman/store/crd"
	"github.com/rancher/rancher/pkg/clustermanager"
	mgmtv3 "github.com/rancher/types/apis/management.cattle.io/v3"
	monitoringv1schema "github.com/rancher/types/apis/monitoring.cattle.io/v1/schema"
	monitoringv1client "github.com/rancher/types/client/monitoring/v1"
	"github.com/rancher/types/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func Register(ctx context.Context, mgmtContext *config.ManagementContext, clusterManager *clustermanager.Manager) {
	clustersClient := mgmtContext.Management.Clusters(metav1.NamespaceAll)

	c := &clusterCRD{
		ctx:            ctx,
		clustersClient: clustersClient,
		clusterManager: clusterManager,
	}

	clustersClient.AddHandler("cluster-crd", c.sync)
}

type clusterCRD struct {
	ctx            context.Context
	clustersClient mgmtv3.ClusterInterface
	clusterManager *clustermanager.Manager
}

func (ch *clusterCRD) sync(key string, cluster *mgmtv3.Cluster) error {
	if cluster == nil || cluster.DeletionTimestamp != nil ||
		!mgmtv3.ClusterConditionReady.IsTrue(cluster) {
		return nil
	}

	if mgmtv3.ClusterConditionAdditionalCRDCreated.IsTrue(cluster) {
		return nil
	}

	clusterTag := getClusterTag(cluster)
	src := cluster
	cpy := src.DeepCopy()

	err := ch.doSync(clusterTag, cpy)
	if err != nil {
		return err
	}

	if !reflect.DeepEqual(cpy, src) {
		_, err := ch.clustersClient.Update(cpy)
		if err != nil {
			return errors.Annotatef(err, "fail to update cluster %s", clusterTag)
		}
	}

	return err
}

func (ch *clusterCRD) doSync(clusterTag string, cluster *mgmtv3.Cluster) error {
	kubeConfig, err := clustermanager.ToRESTConfig(cluster, ch.clusterManager.ScaledContext)
	if err != nil {
		return errors.Annotatef(err, "failed to construct kubeConfig for cluster %s", clusterTag)
	}

	_, err = mgmtv3.ClusterConditionAdditionalCRDCreated.Do(cluster, func() (runtime.Object, error) {
		factory, err := crd.NewFactoryFromClient(*kubeConfig)
		if err != nil {
			return nil, errors.Annotatef(err, "failed to create CRD factory for cluster %s", clusterTag)
		}

		// create additional CRD on here
		factory.BatchCreateCRDs(ch.ctx, config.UserStorageContext, ch.clusterManager.ScaledContext.Schemas, &monitoringv1schema.Version,
			monitoringv1client.PrometheusType,
			monitoringv1client.PrometheusRuleType,
			monitoringv1client.AlertmanagerType,
			monitoringv1client.ServiceMonitorType,
		)

		factory.BatchWait()

		return cluster, nil
	})
	if err != nil {
		return err
	}

	return nil
}

func getClusterTag(cluster *mgmtv3.Cluster) string {
	return fmt.Sprintf("%s(%s)", cluster.Spec.DisplayName, cluster.Name)
}

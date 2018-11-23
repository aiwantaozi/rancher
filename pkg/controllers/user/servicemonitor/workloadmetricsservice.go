package servicemonitor

import (
	"strings"

	util "github.com/rancher/rancher/pkg/controllers/user/workload"
	"github.com/rancher/types/apis/core/v1"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/validation"
)

type MetricsServiceController struct {
	serviceLister v1.ServiceLister
	serviceClient v1.ServiceInterface
}

const (
	creatorIDAnnotation = "field.cattle.io/creatorId"
	metricsAnnotation   = "field.cattle.io/workloadMetrics"
)

func (c *MetricsServiceController) sync(k string, w *util.Workload) error {
	// do not create service for job, cronJob and for workload owned by controller (ReplicaSet)
	if strings.EqualFold(w.Kind, "job") || strings.EqualFold(w.Kind, "cronJob") {
		return nil
	}
	for _, o := range w.OwnerReferences {
		if o.Controller != nil && *o.Controller && (strings.Index(o.APIVersion, ".") < 0 || strings.Contains(o.APIVersion, "k8s.io")) {
			return nil
		}
	}

	if _, ok := w.Annotations[creatorIDAnnotation]; !ok {
		return nil
	}

	if _, ok := w.Annotations[metricsAnnotation]; !ok {
		return nil
	}

	if errs := validation.IsDNS1123Subdomain(w.Name); len(errs) != 0 {
		logrus.Debugf("Not creating service for workload [%s]: dns name is invalid", w.Name)
		return nil
	}

	return c.createService(w)
}

func (c *MetricsServiceController) createService(w *util.Workload) error {

	return nil
}

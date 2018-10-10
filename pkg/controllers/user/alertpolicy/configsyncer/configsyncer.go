package configsyner

import (
	"context"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/common/model"
	"github.com/rancher/norman/controller"
	alertconfig "github.com/rancher/rancher/pkg/controllers/user/alert/config"
	"github.com/rancher/rancher/pkg/controllers/user/alertpolicy/deployer"
	"github.com/rancher/rancher/pkg/controllers/user/alertpolicy/manager"
	monitorutil "github.com/rancher/rancher/pkg/monitoring"

	"github.com/rancher/types/apis/core/v1"
	"github.com/rancher/types/apis/management.cattle.io/v3"
	monitoringv1 "github.com/rancher/types/apis/monitoring.cattle.io/v1"
	"github.com/rancher/types/config"
	"github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
)

var (
	defaultGroupInterval  = 10
	eventGroupInterval    = 1
	defaultInitWait       = 10
	defaultRepeatInterval = 10
)

func NewConfigSyncer(ctx context.Context, cluster *config.UserContext, alertManager *manager.AlertManager, operatorCRDManager *manager.PromOperatorCRDManager) *ConfigSyncer {
	return &ConfigSyncer{
		secrets:                  cluster.Core.Secrets(monitorutil.SystemMonitoringNamespaceName),
		clusterAlertPolicyLister: cluster.Management.Management.ClusterAlertPolicies(cluster.ClusterName).Controller().Lister(),
		projectAlertPolicyLister: cluster.Management.Management.ProjectAlertPolicies("").Controller().Lister(),
		notifierLister:           cluster.Management.Management.Notifiers(cluster.ClusterName).Controller().Lister(),
		clusterName:              cluster.ClusterName,
		alertManager:             alertManager,
		operatorCRDManager:       operatorCRDManager,
	}
}

type ConfigSyncer struct {
	secrets                  v1.SecretInterface
	projectAlertPolicyLister v3.ProjectAlertPolicyLister
	clusterAlertPolicyLister v3.ClusterAlertPolicyLister
	notifierLister           v3.NotifierLister
	clusterName              string
	alertManager             *manager.AlertManager
	operatorCRDManager       *manager.PromOperatorCRDManager
}

func (d *ConfigSyncer) ProjectSync(key string, alert *v3.ProjectAlertPolicy) error {
	return d.sync()
}

func (d *ConfigSyncer) ClusterSync(key string, alert *v3.ClusterAlertPolicy) error {
	return d.sync()
}

func (d *ConfigSyncer) NotifierSync(key string, alert *v3.Notifier) error {
	return d.sync()
}

//sync: update the secret which store the configuration of alertmanager given the latest configured notifiers and alerts rules.
//For each alert, it will generate a route and a receiver in the alertmanager's configuration file, for metric rules it will update operator crd also.
func (d *ConfigSyncer) sync() error {
	if d.alertManager.IsDeploy == false {
		return nil
	}

	if _, err := d.alertManager.GetAlertManagerEndpoint(); err != nil {
		return err
	}

	notifiers, err := d.notifierLister.List("", labels.NewSelector())
	if err != nil {
		return errors.Wrapf(err, "List notifiers")
	}

	clusterAlertPolicies, err := d.clusterAlertPolicyLister.List("", labels.NewSelector())
	if err != nil {
		return errors.Wrapf(err, "List cluster alerts")
	}

	projectAlertPolicies, err := d.projectAlertPolicyLister.List("", labels.NewSelector())
	if err != nil {
		return errors.Wrapf(err, "List project alerts")
	}

	pAlerts := []*v3.ProjectAlertPolicy{}
	for _, alert := range projectAlertPolicies {
		if controller.ObjectInCluster(d.clusterName, alert) {
			pAlerts = append(pAlerts, alert)
		}
	}

	if includeMetrics(clusterAlertPolicies, pAlerts) {
		promRule := manager.GetDefaultPrometheusRule(d.clusterName)
		if err := d.addClusterAlert2Operator(promRule, clusterAlertPolicies); err != nil {
			return err
		}

		if err := d.addProjectAlert2Operator(promRule, pAlerts); err != nil {
			return err
		}

		old, err := d.operatorCRDManager.PrometheusRules.Get(d.clusterName, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			if _, err = d.operatorCRDManager.PrometheusRules.Create(promRule); err != nil && !apierrors.IsAlreadyExists(err) {
				return err
			}
		}
		updated := old.DeepCopy()
		updated.Spec = promRule.Spec
		if _, err = d.operatorCRDManager.PrometheusRules.Update(updated); err != nil {
			return err
		}
	}

	config := d.alertManager.GetDefaultConfig()
	config.Global.PagerdutyURL = "https://events.pagerduty.com/generic/2010-04-15/create_event.json"

	d.addClusterAlert2Config(config, clusterAlertPolicies, notifiers)
	d.addProjectAlert2Config(config, pAlerts, notifiers)

	data, err := yaml.Marshal(config)
	if err != nil {
		return errors.Wrapf(err, "Marshal secrets")
	}

	configSecret, err := d.secrets.Get("alertmanager", metav1.GetOptions{}) //todo: teamwork
	if err != nil {
		return errors.Wrapf(err, "Get secrets")
	}

	if string(configSecret.Data["config.yml"]) != string(data) { //todo: teamwork
		configSecret.Data["config.yml"] = data //todo: teamwork
		configSecret.Data["notification.tmpl"] = []byte(deployer.NotificationTmpl)

		_, err = d.secrets.Update(configSecret)
		if err != nil {
			return errors.Wrapf(err, "Update secrets")
		}

	} else {
		logrus.Debug("The config stay the same, will not update the secret")
	}

	return nil
}

func (d *ConfigSyncer) getNotifier(id string, notifiers []*v3.Notifier) *v3.Notifier {

	for _, n := range notifiers {
		if d.clusterName+":"+n.Name == id {
			return n
		}
	}

	return nil
}

func (d *ConfigSyncer) addProjectAlert2Operator(promRule *monitoringv1.PrometheusRule, alerts []*v3.ProjectAlertPolicy) error {
	for _, alert := range alerts {
		if alert.Status.AlertState == "inactive" {
			continue
		}

		groupID := alert.Namespace + "-" + alert.Name

		ruleGroup, err := d.operatorCRDManager.AddRule(groupID, alert.Name, alert.Spec.Metrics)
		if err != nil {
			return err
		}
		manager.AddRuleGroup(promRule, *ruleGroup)
	}
	return nil
}

func (d *ConfigSyncer) addClusterAlert2Operator(promRule *monitoringv1.PrometheusRule, alerts []*v3.ClusterAlertPolicy) error {
	for _, alert := range alerts {
		if alert.Status.AlertState == "inactive" {
			continue
		}

		groupID := alert.Namespace + "-" + alert.Name

		ruleGroup, err := d.operatorCRDManager.AddRule(groupID, alert.Name, alert.Spec.Metrics)
		if err != nil {
			return err
		}
		manager.AddRuleGroup(promRule, *ruleGroup)
	}

	return nil
}

func (d *ConfigSyncer) addProjectAlert2Config(config *alertconfig.Config, alerts []*v3.ProjectAlertPolicy, notifiers []*v3.Notifier) {
	for _, alert := range alerts {
		if alert.Status.AlertState == "inactive" {
			continue
		}

		groupID := alert.Namespace + "-" + alert.Name
		receiver := &alertconfig.Receiver{Name: groupID}
		exist := d.addRecipients(notifiers, receiver, alert.Spec.Recipients)
		if exist {
			config.Receivers = append(config.Receivers, receiver)
			r1 := d.newRoute(map[string]string{"group_id": groupID}, defaultInitWait, defaultRepeatInterval, defaultGroupInterval)
			d.appendRoute(config.Route, r1)
		}
	}
}

func (d *ConfigSyncer) addClusterAlert2Config(config *alertconfig.Config, alerts []*v3.ClusterAlertPolicy, notifiers []*v3.Notifier) {
	for _, alert := range alerts {
		if alert.Status.AlertState == "inactive" {
			continue
		}

		groupID := alert.Namespace + "-" + alert.Name

		receiver := &alertconfig.Receiver{Name: groupID}
		exist := d.addRecipients(notifiers, receiver, alert.Spec.Recipients)
		if exist {
			config.Receivers = append(config.Receivers, receiver)

			r1 := d.newRoute(map[string]string{"group_id": groupID}, defaultInitWait, defaultRepeatInterval, defaultGroupInterval)
			if alert.Spec.TargetEvents != nil {
				r2 := d.newRoute(map[string]string{"alert_type": "event"}, defaultInitWait, defaultRepeatInterval, eventGroupInterval)
				d.appendRoute(r1, r2)
			}
			d.appendRoute(config.Route, r1)
		}
	}
}

func (d *ConfigSyncer) newRoute(match map[string]string, initalWait, repeatInterval, groupInterval int) *alertconfig.Route {
	route := &alertconfig.Route{
		Receiver: match["group_id"],
		Match:    match,
	}

	gw := model.Duration(time.Duration(initalWait) * time.Second)
	route.GroupWait = &gw
	ri := model.Duration(time.Duration(repeatInterval) * time.Second)
	route.RepeatInterval = &ri

	if groupInterval != defaultGroupInterval {
		gi := model.Duration(time.Duration(groupInterval) * time.Second)
		route.GroupInterval = &gi
	}
	return route
}

func (d *ConfigSyncer) appendRoute(route *alertconfig.Route, subRoute *alertconfig.Route) {
	if route.Routes == nil {
		route.Routes = []*alertconfig.Route{}
	}
	route.Routes = append(route.Routes, subRoute)
}

// func (d *ConfigSyncer) addRoute(route *alertconfig.Route, match map[string]string, initalWait, repeatInterval, groupInterval int) {
// 	r := d.newRoute(match, initalWait, repeatInterval, groupInterval)
// 	d.appendRoute(route, r)
// }

// func (d *ConfigSyncer) addRoute(config *alertconfig.Config, id string, initalWait, repeatInterval, groupInterval int) {
// 	routes := config.Route.Routes
// 	if routes == nil {
// 		routes = []*alertconfig.Route{}
// 	}

// 	match := map[string]string{}
// 	match["group_id"] = id
// 	route := &alertconfig.Route{
// 		Receiver: id,
// 		Match:    match,
// 	}

// 	gw := model.Duration(time.Duration(initalWait) * time.Second)
// 	route.GroupWait = &gw
// 	ri := model.Duration(time.Duration(repeatInterval) * time.Second)
// 	route.RepeatInterval = &ri

// 	if groupInterval != defaultGroupInterval {
// 		gi := model.Duration(time.Duration(groupInterval) * time.Second)
// 		route.GroupInterval = &gi
// 	}

// 	routes = append(routes, route)
// 	config.Route.Routes = routes
// }

func (d *ConfigSyncer) addRecipients(notifiers []*v3.Notifier, receiver *alertconfig.Receiver, recipients []v3.Recipient) bool {
	receiverExist := false
	for _, r := range recipients {
		if r.NotifierName != "" {
			notifier := d.getNotifier(r.NotifierName, notifiers)
			if notifier == nil {
				logrus.Debugf("Can not find the notifier %s", r.NotifierName)
				continue
			}

			if notifier.Spec.PagerdutyConfig != nil {
				pagerduty := &alertconfig.PagerdutyConfig{
					ServiceKey:  alertconfig.Secret(notifier.Spec.PagerdutyConfig.ServiceKey),
					Description: `{{ template "rancher.title" . }}`,
				}
				if r.Recipient != "" {
					pagerduty.ServiceKey = alertconfig.Secret(r.Recipient)
				}
				receiver.PagerdutyConfigs = append(receiver.PagerdutyConfigs, pagerduty)
				receiverExist = true

			} else if notifier.Spec.WebhookConfig != nil {
				webhook := &alertconfig.WebhookConfig{
					URL: notifier.Spec.WebhookConfig.URL,
				}
				if r.Recipient != "" {
					webhook.URL = r.Recipient
				}
				receiver.WebhookConfigs = append(receiver.WebhookConfigs, webhook)
				receiverExist = true
			} else if notifier.Spec.SlackConfig != nil {
				slack := &alertconfig.SlackConfig{
					APIURL:    alertconfig.Secret(notifier.Spec.SlackConfig.URL),
					Channel:   notifier.Spec.SlackConfig.DefaultRecipient,
					Text:      `{{ template "slack.text" . }}`,
					Title:     `{{ template "rancher.title" . }}`,
					TitleLink: "",
					Color:     `{{ if eq (index .Alerts 0).Labels.severity "critical" }}danger{{ else if eq (index .Alerts 0).Labels.severity "warning" }}warning{{ else }}good{{ end }}`,
				}
				if r.Recipient != "" {
					slack.Channel = r.Recipient
				}
				receiver.SlackConfigs = append(receiver.SlackConfigs, slack)
				receiverExist = true

			} else if notifier.Spec.SMTPConfig != nil {
				header := map[string]string{}
				header["Subject"] = `{{ template "rancher.title" . }}`
				email := &alertconfig.EmailConfig{
					Smarthost:    notifier.Spec.SMTPConfig.Host + ":" + strconv.Itoa(notifier.Spec.SMTPConfig.Port),
					AuthPassword: alertconfig.Secret(notifier.Spec.SMTPConfig.Password),
					AuthUsername: notifier.Spec.SMTPConfig.Username,
					RequireTLS:   &notifier.Spec.SMTPConfig.TLS,
					To:           notifier.Spec.SMTPConfig.DefaultRecipient,
					Headers:      header,
					From:         notifier.Spec.SMTPConfig.Sender,
					HTML:         `{{ template "email.text" . }}`,
				}
				if r.Recipient != "" {
					email.To = r.Recipient
				}
				receiver.EmailConfigs = append(receiver.EmailConfigs, email)
				receiverExist = true
			}

		}
	}

	return receiverExist

}

func includeMetrics(clusterAlerts []*v3.ClusterAlertPolicy, projectAlerts []*v3.ProjectAlertPolicy) bool {
	for _, v := range clusterAlerts {
		if v.Spec.Metrics != nil {
			return true
		}
	}

	for _, v := range projectAlerts {
		if v.Spec.Metrics != nil {
			return true
		}
	}

	return false
}

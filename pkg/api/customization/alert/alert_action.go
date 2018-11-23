package alert

import (
	"net/http"
	"strings"

	"github.com/rancher/norman/api/access"
	"github.com/rancher/norman/httperror"
	"github.com/rancher/norman/types"
	"github.com/rancher/types/apis/management.cattle.io/v3"
	"github.com/sirupsen/logrus"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Handler struct {
	ClusterAlertGroup v3.ClusterAlertGroupInterface
	ProjectAlertGroup v3.ProjectAlertGroupInterface
	ClusterAlertRule  v3.ClusterAlertRuleInterface
	ProjectAlertRule  v3.ProjectAlertRuleInterface
	Notifiers         v3.NotifierInterface
}

func RuleFormatter(apiContext *types.APIContext, resource *types.RawResource) {
	resource.AddAction(apiContext, "unmute")
	resource.AddAction(apiContext, "activate")
	resource.AddAction(apiContext, "mute")
	resource.AddAction(apiContext, "deactivate")
}

func GroupFormatter(apiContext *types.APIContext, resource *types.RawResource) {
	resource.AddAction(apiContext, "unmute")
	resource.AddAction(apiContext, "activate")
	resource.AddAction(apiContext, "mute")
	resource.AddAction(apiContext, "deactivate")
}

func (h *Handler) ClusterAlertGroupActionHandler(actionName string, action *types.Action, request *types.APIContext) error {
	parts := strings.Split(request.ID, ":")
	ns := parts[0]
	id := parts[1]

	alert, err := h.ClusterAlertGroup.GetNamespaced(ns, id, metav1.GetOptions{})
	if err != nil {
		logrus.Errorf("Error while getting alert for %s :%v", request.ID, err)
		return err
	}

	switch actionName {
	case "activate":
		if alert.Status.State == "inactive" {
			alert.Status.State = "active"
		} else {
			return httperror.NewAPIError(httperror.ActionNotAvailable, "state is not inactive")
		}

	case "deactivate":
		if alert.Status.State == "active" {
			alert.Status.State = "inactive"
		} else {
			return httperror.NewAPIError(httperror.ActionNotAvailable, "state is not active")
		}

	case "mute":
		if alert.Status.State == "alerting" {
			alert.Status.State = "muted"
		} else {
			return httperror.NewAPIError(httperror.ActionNotAvailable, "state is not alerting")
		}

	case "unmute":
		if alert.Status.State == "muted" {
			alert.Status.State = "alerting"
		} else {
			return httperror.NewAPIError(httperror.ActionNotAvailable, "state is not muted")
		}
	}

	alert, err = h.ClusterAlertGroup.Update(alert)
	if err != nil {
		logrus.Errorf("Error while updating alert:%v", err)
		return err
	}

	data := map[string]interface{}{}
	if err := access.ByID(request, request.Version, request.Type, request.ID, &data); err != nil {
		return err
	}
	request.WriteResponse(http.StatusOK, data)
	return nil
}

func (h *Handler) ProjectAlertGroupActionHandler(actionName string, action *types.Action, request *types.APIContext) error {
	parts := strings.Split(request.ID, ":")
	ns := parts[0]
	id := parts[1]

	alert, err := h.ProjectAlertGroup.GetNamespaced(ns, id, metav1.GetOptions{})
	if err != nil {
		logrus.Errorf("Error while getting alert for %s :%v", request.ID, err)
		return err
	}

	switch actionName {
	case "activate":
		if alert.Status.State == "inactive" {
			alert.Status.State = "active"
		} else {
			return httperror.NewAPIError(httperror.ActionNotAvailable, "state is not inactive")
		}

	case "deactivate":
		if alert.Status.State == "active" {
			alert.Status.State = "inactive"
		} else {
			return httperror.NewAPIError(httperror.ActionNotAvailable, "state is not active")
		}

	case "mute":
		if alert.Status.State == "alerting" {
			alert.Status.State = "muted"
		} else {
			return httperror.NewAPIError(httperror.ActionNotAvailable, "state is not alerting")
		}

	case "unmute":
		if alert.Status.State == "muted" {
			alert.Status.State = "alerting"
		} else {
			return httperror.NewAPIError(httperror.ActionNotAvailable, "state is not muted")
		}
	}

	alert, err = h.ProjectAlertGroup.Update(alert)
	if err != nil {
		logrus.Errorf("Error while updating alert:%v", err)
		return err
	}

	data := map[string]interface{}{}
	if err := access.ByID(request, request.Version, request.Type, request.ID, &data); err != nil {
		return err
	}
	request.WriteResponse(http.StatusOK, data)
	return nil
}

type RuleHandler struct {
	ClusterAlertRule v3.ClusterAlertRuleInterface
	ProjectAlertRule v3.ProjectAlertRuleInterface
	Notifiers        v3.NotifierInterface
}

func (h *Handler) ClusterAlertRuleActionHandler(actionName string, action *types.Action, request *types.APIContext) error {
	parts := strings.Split(request.ID, ":")
	ns := parts[0]
	id := parts[1]

	alert, err := h.ClusterAlertRule.GetNamespaced(ns, id, metav1.GetOptions{})
	if err != nil {
		logrus.Errorf("Error while getting alert for %s :%v", request.ID, err)
		return err
	}

	switch actionName {
	case "activate":
		if alert.Status.State == "inactive" {
			alert.Status.State = "active"
		} else {
			return httperror.NewAPIError(httperror.ActionNotAvailable, "state is not inactive")
		}

	case "deactivate":
		if alert.Status.State == "active" {
			alert.Status.State = "inactive"
		} else {
			return httperror.NewAPIError(httperror.ActionNotAvailable, "state is not active")
		}

	case "mute":
		if alert.Status.State == "alerting" {
			alert.Status.State = "muted"
		} else {
			return httperror.NewAPIError(httperror.ActionNotAvailable, "state is not alerting")
		}

	case "unmute":
		if alert.Status.State == "muted" {
			alert.Status.State = "alerting"
		} else {
			return httperror.NewAPIError(httperror.ActionNotAvailable, "state is not muted")
		}
	}

	alert, err = h.ClusterAlertRule.Update(alert)
	if err != nil {
		logrus.Errorf("Error while updating alert:%v", err)
		return err
	}

	data := map[string]interface{}{}
	if err := access.ByID(request, request.Version, request.Type, request.ID, &data); err != nil {
		return err
	}
	request.WriteResponse(http.StatusOK, data)
	return nil
}

func (h *Handler) ProjectAlertRuleActionHandler(actionName string, action *types.Action, request *types.APIContext) error {
	parts := strings.Split(request.ID, ":")
	ns := parts[0]
	id := parts[1]

	alert, err := h.ProjectAlertRule.GetNamespaced(ns, id, metav1.GetOptions{})
	if err != nil {
		logrus.Errorf("Error while getting alert for %s :%v", request.ID, err)
		return err
	}

	switch actionName {
	case "activate":
		if alert.Status.State == "inactive" {
			alert.Status.State = "active"
		} else {
			return httperror.NewAPIError(httperror.ActionNotAvailable, "state is not inactive")
		}

	case "deactivate":
		if alert.Status.State == "active" {
			alert.Status.State = "inactive"
		} else {
			return httperror.NewAPIError(httperror.ActionNotAvailable, "state is not active")
		}

	case "mute":
		if alert.Status.State == "alerting" {
			alert.Status.State = "muted"
		} else {
			return httperror.NewAPIError(httperror.ActionNotAvailable, "state is not alerting")
		}

	case "unmute":
		if alert.Status.State == "muted" {
			alert.Status.State = "alerting"
		} else {
			return httperror.NewAPIError(httperror.ActionNotAvailable, "state is not muted")
		}
	}

	alert, err = h.ProjectAlertRule.Update(alert)
	if err != nil {
		logrus.Errorf("Error while updating alert:%v", err)
		return err
	}

	data := map[string]interface{}{}
	if err := access.ByID(request, request.Version, request.Type, request.ID, &data); err != nil {
		return err
	}
	request.WriteResponse(http.StatusOK, data)
	return nil
}

// func (h *Handler) ClusterAlertRuleHandler(actionName string, action *types.Action, request *types.APIContext) error {
// 	parts := strings.Split(request.ID, "-")
// 	groupID := parts[0]
// 	ruleID := parts[1]
// 	parts = strings.Split(groupID, ":")
// 	ns := parts[0]
// 	groupName := parts[1]

// 	alert, err := h.ClusterAlertGroup.GetNamespaced(ns, groupName, metav1.GetOptions{})
// 	if err != nil {
// 		logrus.Errorf("Error while getting alert for %s :%v", request.ID, err)
// 		return err
// 	}

// 	if err = updateAlertRuleState(actionName, ruleID, alert.Status.RuleStates[:]); err != nil {
// 		return err
// 	}

// 	alert, err = h.ClusterAlertGroup.Update(alert)
// 	if err != nil {
// 		logrus.Errorf("Error while updating alert:%v", err)
// 		return err
// 	}

// 	data := map[string]interface{}{}
// 	if err := access.ByID(request, request.Version, request.Type, request.ID, &data); err != nil {
// 		return err
// 	}
// 	request.WriteResponse(http.StatusOK, data)
// 	return nil
// }

// func (h *Handler) ProjectAlertRuleHandler(actionName string, action *types.Action, request *types.APIContext) error {
// 	parts := strings.Split(request.ID, "-")
// 	groupID := parts[0]
// 	ruleID := parts[1]
// 	parts = strings.Split(groupID, ":")
// 	ns := parts[0]
// 	groupName := parts[1]

// 	alert, err := h.ProjectAlertGroup.GetNamespaced(ns, groupName, metav1.GetOptions{})
// 	if err != nil {
// 		logrus.Errorf("Error while getting alert for %s :%v", request.ID, err)
// 		return err
// 	}

// 	if err = updateAlertRuleState(actionName, ruleID, alert.Status.RuleStates[:]); err != nil {
// 		return err
// 	}

// 	alert, err = h.ProjectAlertGroup.Update(alert)
// 	if err != nil {
// 		logrus.Errorf("Error while updating alert:%v", err)
// 		return err
// 	}

// 	data := map[string]interface{}{}
// 	if err := access.ByID(request, request.Version, request.Type, request.ID, &data); err != nil {
// 		return err
// 	}
// 	request.WriteResponse(http.StatusOK, data)
// 	return nil
// }

// func updateAlertRuleState(actionName, ruleID string, alertRules []v3.RuleState) error {
// 	for _, v := range alertRules {
// 		if v.RuleID == ruleID {
// 			switch actionName {
// 			case "activate":
// 				if v.State == "inactive" {
// 					v.State = "active"
// 				} else {
// 					return httperror.NewAPIError(httperror.ActionNotAvailable, "state is not inactive")
// 				}

// 			case "deactivate":
// 				if v.State == "active" {
// 					v.State = "inactive"
// 				} else {
// 					return httperror.NewAPIError(httperror.ActionNotAvailable, "state is not active")
// 				}

// 			case "mute":
// 				if v.State == "alerting" {
// 					v.State = "muted"
// 				} else {
// 					return httperror.NewAPIError(httperror.ActionNotAvailable, "state is not alerting")
// 				}

// 			case "unmute":
// 				if v.State == "muted" {
// 					v.State = "alerting"
// 				} else {
// 					return httperror.NewAPIError(httperror.ActionNotAvailable, "state is not muted")
// 				}
// 			}
// 		}
// 	}
// 	return nil
// }

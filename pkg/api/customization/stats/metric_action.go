package stats

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rancher/norman/parse"
	"github.com/rancher/norman/types"
	"github.com/rancher/norman/types/convert"
	"github.com/rancher/rancher/pkg/clustermanager"
	"github.com/rancher/types/config/dialer"
)

func newMetricHandler(dialerFactory dialer.Factory, clustermanager *clustermanager.Manager) *metricHandler {
	return &metricHandler{
		dialerFactory:  dialerFactory,
		clustermanager: clustermanager,
	}
}

type metricHandler struct {
	dialerFactory  dialer.Factory
	clustermanager *clustermanager.Manager
}

func (h *metricHandler) action(actionName string, action *types.Action, apiContext *types.APIContext) error {
	clusterName := getID(apiContext.Request.URL.Path, clusterPrefix)
	userContext, err := h.clustermanager.UserContext(clusterName)
	if err != nil {
		return fmt.Errorf("get usercontext failed, %v", err)
	}

	prometheusQuery, err := NewPrometheusQuery(userContext, clusterName, h.clustermanager, h.dialerFactory)
	if err != nil {
		return err
	}

	switch actionName {
	case queryAction:
		var queryMetricInput QueryMetricInput
		actionInput, err := parse.ReadBody(apiContext.Request)
		if err != nil {
			return err
		}
		if err = convert.ToObj(actionInput, &queryMetricInput); err != nil {
			return err
		}

		start, end, step, err := parseTimeParams(queryMetricInput.From, queryMetricInput.To, queryMetricInput.Interval)
		if err != nil {
			return err
		}

		query := InitPromQuery("", start, end, step, queryMetricInput.Expr, "", nil, false)
		seriesSlice, err := prometheusQuery.QueryRange(query)
		if err != nil {
			return err
		}

		if seriesSlice == nil {
			apiContext.WriteResponse(http.StatusNoContent, nil)
			return nil
		}

		data := map[string]interface{}{
			"type":   "queryMetricOutput",
			"series": seriesSlice,
		}

		res, err := json.Marshal(data)
		if err != nil {
			return fmt.Errorf("marshal query stats result failed, %v", err)
		}
		apiContext.Response.Write(res)

	case listMetricNameAction:
		names, err := prometheusQuery.GetLabelValues("__name__")
		if err != nil {
			return fmt.Errorf("get metric list failed, %v", err)
		}
		data := map[string]interface{}{
			"type":  "metricNamesOutput",
			"names": names,
		}
		apiContext.WriteResponse(http.StatusOK, data)
	}
	return nil

}

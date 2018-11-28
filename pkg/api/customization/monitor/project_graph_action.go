package monitor

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rancher/rancher/pkg/ref"

	mgmtclientv3 "github.com/rancher/types/client/management/v3"

	"github.com/rancher/norman/api/access"
	"github.com/rancher/norman/parse"
	"github.com/rancher/norman/types"
	"github.com/rancher/norman/types/convert"
	"github.com/rancher/rancher/pkg/clustermanager"
	monitorutil "github.com/rancher/rancher/pkg/monitoring"
	"github.com/rancher/types/apis/management.cattle.io/v3"
	"github.com/rancher/types/config/dialer"
)

func NewProjectGraphHandler(dialerFactory dialer.Factory, clustermanager *clustermanager.Manager) *projectGraphHandler {
	return &projectGraphHandler{
		dialerFactory:  dialerFactory,
		clustermanager: clustermanager,
	}
}

type projectGraphHandler struct {
	dialerFactory  dialer.Factory
	clustermanager *clustermanager.Manager
}

func (h *projectGraphHandler) QuerySeriesAction(actionName string, action *types.Action, apiContext *types.APIContext) error {
	var queryGraphInput v3.QueryGraphInput
	actionInput, err := parse.ReadBody(apiContext.Request)
	if err != nil {
		return err
	}

	if err = convert.ToObj(actionInput, &queryGraphInput); err != nil {
		return err
	}

	inputParser := newProjectGraphInputParser(queryGraphInput)
	if err = inputParser.parse(); err != nil {
		return err
	}

	clusterName := inputParser.ClusterName
	userContext, err := h.clustermanager.UserContext(clusterName)
	if err != nil {
		return fmt.Errorf("get usercontext failed, %v", err)
	}

	check := newAuthChecker(apiContext.Request.Context(), userContext, inputParser.Input, inputParser.ProjectName)
	if err = check.check(); err != nil {
		return err
	}

	token, err := getAuthToken(userContext, monitorutil.ClusterAppName, monitorutil.CattleNamespaceName) //todo
	if err != nil {
		return err
	}

	reqContext, cancel := context.WithTimeout(context.Background(), prometheusReqTimeout)
	defer cancel()

	prometheusQuery, err := NewPrometheusQuery(reqContext, userContext, clusterName, token, h.clustermanager, h.dialerFactory)
	if err != nil {
		return err
	}

	var graphs []mgmtclientv3.ProjectMonitorGraph
	err = access.List(apiContext, apiContext.Version, mgmtclientv3.ProjectMonitorGraphType, &types.QueryOptions{Conditions: inputParser.Conditions}, &graphs)
	if err != nil {
		return err
	}

	var queries []*PrometheusQuery
	for _, graph := range graphs {
		g := graph
		_, projectName := ref.Parse(graph.ProjectID)
		refName := getRefferenceGraphName(projectName, graph.Name)
		monitorMetrics, err := graph2Metrics(userContext, clusterName, g.ResourceType, refName, graph.MetricsSelector, graph.DetailsMetricsSelector, inputParser.Input.MetricParams, inputParser.Input.IsDetails)
		if err != nil {
			return err
		}

		queries = append(queries, metrics2PrometheusQuery(monitorMetrics, inputParser.Start, inputParser.End, inputParser.Step, isInstanceGraph(g.GraphType))...)
	}

	seriesSlice, err := prometheusQuery.Do(queries)
	if err != nil {
		return fmt.Errorf("query series failed, %v", err)
	}

	if seriesSlice == nil {
		apiContext.WriteResponse(http.StatusNoContent, nil)
		return nil
	}

	collection := v3.QueryProjectGraphOutput{Type: "collection"}
	for k, v := range seriesSlice {
		name := k

		queryGraph := v3.QueryProjectGraph{
			GraphName: name,
			Series:    parseResponse(v),
		}
		collection.Data = append(collection.Data, queryGraph)
	}

	res, err := json.Marshal(collection)
	if err != nil {
		return fmt.Errorf("marshal query series result failed, %v", err)
	}
	apiContext.Response.Write(res)
	return nil
}

func parseResponse(seriesSlice []*TimeSeries) []*v3.TimeSeries {
	var series []*v3.TimeSeries
	for _, v := range seriesSlice {
		series = append(series, &v3.TimeSeries{
			Name:   v.Name,
			Points: v.Points,
		})
	}
	return series
}

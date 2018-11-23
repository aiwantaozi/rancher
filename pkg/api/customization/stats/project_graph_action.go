package stats

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rancher/rancher/pkg/ref"

	"github.com/rancher/norman/parse"
	"github.com/rancher/norman/types"
	"github.com/rancher/norman/types/convert"
	"github.com/rancher/rancher/pkg/clustermanager"
	monitorutil "github.com/rancher/rancher/pkg/monitoring"
	clusterv3 "github.com/rancher/types/apis/cluster.cattle.io/v3"
	"github.com/rancher/types/config/dialer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func newProjectGraphHandler(dialerFactory dialer.Factory, clustermanager *clustermanager.Manager) *projectGraphHandler {
	return &projectGraphHandler{
		dialerFactory:  dialerFactory,
		clustermanager: clustermanager,
	}
}

type projectGraphHandler struct {
	dialerFactory  dialer.Factory
	clustermanager *clustermanager.Manager
}

func (h *projectGraphHandler) querySeriesAction(actionName string, action *types.Action, apiContext *types.APIContext) error {
	var queryGraphInput QueryGraphInput
	actionInput, err := parse.ReadBody(apiContext.Request)
	if err != nil {
		return err
	}

	if err = convert.ToObj(actionInput, &queryGraphInput); err != nil {
		return err
	}

	projectName := getID(apiContext.Request.URL.Path, projectPrefix)
	clusterName, _ := ref.Parse(projectName)

	userContext, err := h.clustermanager.UserContext(clusterName)
	if err != nil {
		return fmt.Errorf("get usercontext failed, %v", err)
	}

	inputParser := newQueryGraphInputParser(apiContext.Request.Context(), userContext, queryGraphInput, false, projectName)
	if err = inputParser.parse(); err != nil {
		return err
	}

	prometheusQuery, err := NewPrometheusQuery(userContext, clusterName, h.clustermanager, h.dialerFactory)
	if err != nil {
		return err
	}

	graphs, err := userContext.ClusterV3.MonitorGraphs(monitorutil.CattleNamespaceName).List(metav1.ListOptions{LabelSelector: labels.Set(queryGraphInput.GraphSelector).AsSelector().String()})
	if err != nil {
		return fmt.Errorf("list monitor graph with label %+v failed, %v", queryGraphInput.GraphSelector, err)
	}

	var queries []*PrometheusQuery
	graphMap := make(map[string]*clusterv3.MonitorGraph)
	for _, graph := range graphs.Items {
		g := graph
		graphMap[getRefferenceGraphName(graph.Namespace, graph.Name)] = &g
		monitorMetrics, err := graph2Metrics(userContext, clusterName, &graph, inputParser.Input.MetricParams, inputParser.Input.IsDetails)
		if err != nil {
			return err
		}

		queries = append(queries, metrics2PrometheusQuery(monitorMetrics, inputParser.Start, inputParser.End, inputParser.Step, inputParser.Input.IsInstanceQuery)...)
	}

	seriesSlice, err := prometheusQuery.Do(queries)
	if err != nil {
		return fmt.Errorf("query series failed, %v", err)
	}

	if seriesSlice == nil {
		apiContext.WriteResponse(http.StatusNoContent, nil)
		return nil
	}

	timeSpliceMap := make(map[string]TimeResponseSeriesSlice)
	for _, v := range seriesSlice {
		graphName, _ := parseID(v.ID)
		timeSpliceMap[graphName] = append(timeSpliceMap[graphName], v)

	}

	collection := QueryGraphOutputCollection{Type: "collection"}
	for k, v := range timeSpliceMap {
		queryGraph := QueryGraph{
			Graph:  graphMap[k].Spec,
			Series: v,
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

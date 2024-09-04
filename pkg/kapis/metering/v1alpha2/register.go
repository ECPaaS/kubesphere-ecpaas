/*
Copyright(c) 2024-present Accton. All rights reserved. www.accton.com.tw
*/

package v1alpha2

import (
	"net/http"

	"kubesphere.io/kubesphere/pkg/models/openpitrix"

	"github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/cache"

	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/informers"
	monitoringv1alpha3 "kubesphere.io/kubesphere/pkg/kapis/monitoring/v1alpha3"
	model "kubesphere.io/kubesphere/pkg/models/monitoring"
	resourcev1alpha3 "kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/resource"
	meteringclient "kubesphere.io/kubesphere/pkg/simple/client/metering"
	"kubesphere.io/kubesphere/pkg/simple/client/monitoring"
)

const (
	groupName = "metering.kubesphere.io"
	respOK    = "ok"
)

var GroupVersion = schema.GroupVersion{Group: groupName, Version: "v1alpha2"}

func AddToContainer(c *restful.Container, k8sClient kubernetes.Interface, meteringClient monitoring.Interface, factory informers.InformerFactory, cache cache.Cache, meteringOptions *meteringclient.Options, opClient openpitrix.Interface, rtClient runtimeclient.Client) error {
	ws := runtime.NewWebService(GroupVersion)

	h := newHandler(k8sClient, meteringClient, factory, resourcev1alpha3.NewResourceGetter(factory, cache), meteringOptions, opClient, rtClient)

	ws.Route(ws.GET("/workspaces/{workspace}").
		To(h.HandleWorkspaceMeterQuery).
		Doc("Get workspace-level meter data of a specific workspace.").
		Param(ws.QueryParameter("operation", "Metering operation.").DataType("string").Required(false).DefaultValue(monitoringv1alpha3.OperationQuery)).
		Param(ws.PathParameter("workspace", "Workspace name.").DataType("string").Required(true)).
		Param(ws.QueryParameter("metrics_filter", "The metric name filter consists of a regexp pattern. It specifies which metric data to return. For example, the following filter matches both workspace CPU usage and memory usage: `meter_workspace_cpu_usage|meter_workspace_memory_usage`.").DataType("string").Required(false)).
		Param(ws.QueryParameter("storageclass", "The name of the storageclass.").DataType("string").Required(false)).
		Param(ws.QueryParameter("pvc_filter", "The PVC filter consists of a regexp pattern. It specifies which PVC data to return.").DataType("string").Required(false)).
		Param(ws.QueryParameter("start", "Start time of query. Use **start** and **end** to retrieve metric data over a time span. It is a string with Unix time format, eg. 1559347200. ").DataType("string").Required(false)).
		Param(ws.QueryParameter("end", "End time of query. Use **start** and **end** to retrieve metric data over a time span. It is a string with Unix time format, eg. 1561939200. ").DataType("string").Required(false)).
		Param(ws.QueryParameter("step", "Time interval. Retrieve metric data at a fixed interval within the time range of start and end. It requires both **start** and **end** are provided. The format is [0-9]+[smhdwy]. Defaults to 10m (i.e. 10 min).").DataType("string").DefaultValue("10m").Required(false)).
		Param(ws.QueryParameter("time", "A timestamp in Unix time format. Retrieve metric data at a single point in time. Defaults to now. Time and the combination of start, end, step are mutually exclusive.").DataType("string").Required(false)).
		Param(ws.QueryParameter("type", "Additional operations. Currently available types is statistics. It retrieves the total number of namespaces, devops projects, members and roles in this workspace at the moment.").DataType("string").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.WorkspaceMetersTag}).
		Writes(model.Metrics{}).
		Returns(http.StatusOK, respOK, model.Metrics{})).
		Produces(restful.MIME_JSON)

	c.Add(ws)
	return nil
}

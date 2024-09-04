/*
Copyright(c) 2024-present Accton. All rights reserved. www.accton.com.tw
*/

package v1alpha4

import (
	"net/http"

	"github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/cache"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/authorization/authorizer"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/models/iam/am"
	"kubesphere.io/kubesphere/pkg/models/iam/im"
	"kubesphere.io/kubesphere/pkg/models/metering"
	"kubesphere.io/kubesphere/pkg/models/monitoring"
	"kubesphere.io/kubesphere/pkg/models/openpitrix"
	resourcev1alpha3 "kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/resource"
	"kubesphere.io/kubesphere/pkg/simple/client/auditing"
	"kubesphere.io/kubesphere/pkg/simple/client/events"
	"kubesphere.io/kubesphere/pkg/simple/client/logging"
	meteringclient "kubesphere.io/kubesphere/pkg/simple/client/metering"
	monitoringclient "kubesphere.io/kubesphere/pkg/simple/client/monitoring"
)

const (
	GroupName = "tenant.kubesphere.io"
)

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1alpha4"}

func Resource(resource string) schema.GroupResource {
	return GroupVersion.WithResource(resource).GroupResource()
}

func AddToContainer(c *restful.Container, factory informers.InformerFactory, k8sclient kubernetes.Interface,
	ksclient kubesphere.Interface, evtsClient events.Client, loggingClient logging.Client,
	auditingclient auditing.Client, am am.AccessManagementInterface, im im.IdentityManagementInterface, authorizer authorizer.Authorizer,
	monitoringclient monitoringclient.Interface, cache cache.Cache, meteringOptions *meteringclient.Options, opClient openpitrix.Interface) error {

	ws := runtime.NewWebService(GroupVersion)
	handler := NewTenantHandler(factory, k8sclient, ksclient, evtsClient, loggingClient, auditingclient, am, im, authorizer, monitoringclient, resourcev1alpha3.NewResourceGetter(factory, cache), meteringOptions, opClient)

	ws.Route(ws.GET("/metering").
		To(handler.QueryMetering).
		Doc("Get meterings against the cluster.").
		Param(ws.QueryParameter("level", "Metering level.").DataType("string").Required(true)).
		Param(ws.QueryParameter("node", "Node name.").DataType("string").Required(false)).
		Param(ws.QueryParameter("workspace", "Workspace name.").DataType("string").Required(false)).
		Param(ws.QueryParameter("namespace", "Namespace name.").DataType("string").Required(false)).
		Param(ws.QueryParameter("kind", "Workload kind. One of deployment, daemonset, statefulset.").DataType("string").Required(false)).
		Param(ws.QueryParameter("workload", "Workload name.").DataType("string").Required(false)).
		Param(ws.QueryParameter("pod", "Pod name.").DataType("string").Required(false)).
		Param(ws.QueryParameter("applications", "Appliction names, format app_name[:app_version](such as nginx:v1, nignx) which are joined by \"|\" ").DataType("string").Required(false)).
		Param(ws.QueryParameter("services", "Services which are joined by \"|\".").DataType("string").Required(false)).
		Param(ws.QueryParameter("metrics_filter", "The metric name filter consists of a regexp pattern. It specifies which metric data to return. For example, the following filter matches both workspace CPU usage and memory usage: `meter_workspace_cpu_usage|meter_workspace_memory_usage`.").DataType("string").Required(false)).
		Param(ws.QueryParameter("resources_filter", "The workspace filter consists of a regexp pattern. It specifies which workspace data to return.").DataType("string").Required(false)).
		Param(ws.QueryParameter("start", "Start time of query. Use **start** and **end** to retrieve metric data over a time span. It is a string with Unix time format, eg. 1559347200. ").DataType("string").Required(false)).
		Param(ws.QueryParameter("end", "End time of query. Use **start** and **end** to retrieve metric data over a time span. It is a string with Unix time format, eg. 1561939200. ").DataType("string").Required(false)).
		Param(ws.QueryParameter("step", "Time interval. Retrieve metric data at a fixed interval within the time range of start and end. It requires both **start** and **end** are provided. The format is [0-9]+[smhdwy]. Defaults to 10m (i.e. 10 min).").DataType("string").DefaultValue("10m").Required(false)).
		Param(ws.QueryParameter("time", "A timestamp in Unix time format. Retrieve metric data at a single point in time. Defaults to now. Time and the combination of start, end, step are mutually exclusive.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_metric", "Sort workspaces by the specified metric. Not applicable if **start** and **end** are provided.").DataType("string").Required(false)).
		Param(ws.QueryParameter("sort_type", "Sort order. One of asc, desc.").DefaultValue("desc.").DataType("string").Required(false)).
		Param(ws.QueryParameter("page", "The page number. This field paginates result data of each metric, then returns a specific page. For example, setting **page** to 2 returns the second page. It only applies to sorted metric data.").DataType("integer").Required(false)).
		Param(ws.QueryParameter("limit", "Page size, the maximum number of results in a single page. Defaults to 5.").DataType("integer").Required(false).DefaultValue("5")).
		Param(ws.QueryParameter("storageclass", "The name of the storageclass.").DataType("string").Required(false)).
		Param(ws.QueryParameter("pvc_filter", "The PVC filter consists of a regexp pattern. It specifies which PVC data to return.").DataType("string").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.WorkspaceMetersTag}).
		Writes(monitoring.Metrics{}).
		Returns(http.StatusOK, api.StatusOK, monitoring.Metrics{}))

	ws.Route(ws.GET("/metering/price").
		To(handler.HandlePriceInfoQuery).
		Doc("Get resoure price.").
		Writes(metering.PriceInfo{}).
		Returns(http.StatusOK, api.StatusOK, metering.PriceInfo{}))

	c.Add(ws)
	return nil
}

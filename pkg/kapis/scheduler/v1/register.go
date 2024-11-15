/*
Copyright(c) 2024-present Accton. All rights reserved. www.accton.com.tw
*/

package v1

import (
	"net/http"

	"github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
)

const (
	GroupName    = "scheduler.ecpaas.io"
	schedulerTag = "Scheduler"
)

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1"}

func Resource(resource string) schema.GroupResource {
	return GroupVersion.WithResource(resource).GroupResource()
}

func AddToContainer(container *restful.Container, k8sclient kubernetes.Interface, ksclient kubesphere.Interface) error {
	webservice := runtime.NewWebService(GroupVersion)
	handler := newHandler(k8sclient, ksclient)

	webservice.Route(webservice.GET("/scheduler").
		To(handler.ListSchedulerName).
		Doc("List all of scheduler name").
		Notes("This API provides multiple schedulerName options that can be specified in the Job resource's schedulerName field under the spec definition. "+
			"Users can choose from the available schedulers to manage how their job is scheduled in the cluster.").
		Returns(http.StatusOK, api.StatusOK, SchedulerNameResponse{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{schedulerTag}))

	webservice.Route(webservice.GET("/yunikorn/queues").
		To(handler.ListYuniKornQueues).
		Doc("List YouniKorn's queues").
		Notes("This API provides the available YuniKorn leaf queues, which can be specified in the Job resource's labels, e.g., queue: root.system.high-priority").
		Returns(http.StatusOK, api.StatusOK, YunikornQueuesResponse{}).
		Returns(http.StatusNotFound, api.StatusNotFound, nil).
		Returns(http.StatusInternalServerError, api.StatusInternalServerError, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{schedulerTag}))

	container.Add(webservice)

	return nil
}

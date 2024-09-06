/*
Copyright(c) 2024-present Accton. All rights reserved. www.accton.com.tw
*/

package v1alpha2

import (
	"github.com/emicklei/go-restful"
	"k8s.io/client-go/kubernetes"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/models/openpitrix"

	"kubesphere.io/kubesphere/pkg/informers"
	monitorhle "kubesphere.io/kubesphere/pkg/kapis/monitoring/v1alpha3"
	resourcev1alpha3 "kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/resource"
	meteringclient "kubesphere.io/kubesphere/pkg/simple/client/metering"
	"kubesphere.io/kubesphere/pkg/simple/client/monitoring"
)

type meterHandler interface {
	HandleWorkspaceMeterQuery(req *restful.Request, resp *restful.Response)
}

func newHandler(k kubernetes.Interface, m monitoring.Interface, f informers.InformerFactory, resourceGetter *resourcev1alpha3.ResourceGetter, meteringOptions *meteringclient.Options, opClient openpitrix.Interface, rtClient runtimeclient.Client) meterHandler {
	return monitorhle.NewHandler(k, m, nil, f, resourceGetter, meteringOptions, opClient, rtClient)
}

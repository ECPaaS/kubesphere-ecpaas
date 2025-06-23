/*
Copyright(c) 2025-present Accton. All rights reserved. www.accton.com.tw
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
	"kubesphere.io/kubesphere/pkg/constants"
)

const (
	GroupName = "file.ecpaas.io"
)

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1"}

func AddToContainer(container *restful.Container, k8sclient kubernetes.Interface) error {
	webservice := runtime.NewWebService(GroupVersion)
	handler := newHandler(k8sclient)

	formData := webservice.FormParameter("file", "File Stream form-data").Required(true)
	formData.DataType("file")
	webservice.Route(webservice.POST("/upload").
		To(handler.UploadFile).
		Doc("Upload file to PVC").
		Consumes("multipart/form-data").
		Param(formData).
		Param(webservice.FormParameter("namespace", "The namespace where the PVC is located").Required(true)).
		Param(webservice.FormParameter("pvcName", "PVC name").Required(true)).
		Param(webservice.FormParameter("targetPath", "Specify the target path where the file is stored in the PVC").Required(true)).
		Returns(http.StatusOK, api.StatusOK, FileUploadResponse{}).
		Returns(http.StatusBadRequest, api.StatusBadRequest, nil).
		Returns(http.StatusInternalServerError, api.StatusInternalServerError, nil).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.FileTag}))

	container.Add(webservice)

	return nil
}

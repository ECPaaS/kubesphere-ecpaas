/*
Copyright(c) 2023-present Accton. All rights reserved. www.accton.com.tw
*/

package v1

import (
	"net/http"

	"github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
	"github.com/minio/minio-go/v7"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/server/errors"
)

const (
	GroupName = "virtualization.ecpaas.io"
)

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1alpha1"}

func Resource(resource string) schema.GroupResource {
	return GroupVersion.WithResource(resource).GroupResource()
}

func AddToContainer(container *restful.Container, minioClient *minio.Client, k8sclient kubernetes.Interface, ksclient kubesphere.Interface) error {
	webservice := runtime.NewWebService(GroupVersion)
	handler := newHandler(minioClient, k8sclient, ksclient)

	webservice.Route(webservice.GET("/minio/images").
		To(handler.ListMinioObjects).
		Doc("List all Minio images").
		Returns(http.StatusOK, api.StatusOK, ImagesList{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.MinioImageTag}))

	webservice.Route(webservice.GET("/namespaces/{namespace}/minio/images").
		To(handler.ListMinioObjectsWithNs).
		Doc("List all Minio images with namespace").
		Param(webservice.PathParameter("namespace", "name of a namespace").Required(true)).
		Returns(http.StatusOK, api.StatusOK, ImagesList{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.MinioImageTag}))

	webservice.Route(webservice.GET("/minio/image/checkFileExist/{imageName}").
		To(handler.GetMinioObjectStatus).
		Doc("Check if Minio image exist").
		Param(webservice.PathParameter("imageName", "Image name").Required(true)).
		Returns(http.StatusOK, api.StatusOK, ObjectStatus{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.MinioImageTag}))

	webservice.Route(webservice.GET("/namespaces/{namespace}/minio/image/checkFileExist/{imageName}").
		To(handler.GetMinioObjectStatusWithNs).
		Doc("Check If Minio image exist with namespace").
		Param(webservice.PathParameter("imageName", "Image name").Required(true)).
		Param(webservice.PathParameter("namespace", "name of a namespace").Required(true)).
		Returns(http.StatusOK, api.StatusOK, ObjectStatus{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.MinioImageTag}))

	formData := webservice.FormParameter("uploadfile", "File Stream form-data").Required(true)
	formData.DataType("file")
	webservice.Route(webservice.POST("/minio/image").
		To(handler.UploadMinioObject).
		Doc("Upload Minio Image").
		Consumes("multipart/form-data").
		Param(formData).
		Returns(http.StatusOK, api.StatusOK, errors.None).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.MinioImageTag}))

	webservice.Route(webservice.POST("/namespaces/{namespace}/minio/image").
		To(handler.UploadMinioObjectWithNs).
		Doc("Upload Minio Image with namespace").
		Consumes("multipart/form-data").
		Param(formData).
		Param(webservice.PathParameter("namespace", "name of a namespace").Required(true)).
		Returns(http.StatusOK, api.StatusOK, errors.None).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.MinioImageTag}))

	webservice.Route(webservice.DELETE("/minio/image/{imageName}").
		To(handler.DeleteMinioObject).
		Doc("Delete Minio Image").
		Param(webservice.PathParameter("imageName", "Image name").Required(true)).
		Returns(http.StatusOK, api.StatusOK, errors.None).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.MinioImageTag}))

	webservice.Route(webservice.DELETE("/namespaces/{namespace}/minio/image/{imageName}").
		To(handler.DeleteMinioObjectWithNs).
		Doc("Delete Minio Image with namespace").
		Param(webservice.PathParameter("namespace", "name of a namespace").Required(true)).
		Param(webservice.PathParameter("imageName", "Image name").Required(true)).
		Returns(http.StatusOK, api.StatusOK, errors.None).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.MinioImageTag}))

	container.Add(webservice)

	return nil
}

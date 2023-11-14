/*
Copyright(c) 2023-present Accton. All rights reserved. www.accton.com.tw
*/

package v1

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/emicklei/go-restful"
	"github.com/minio/minio-go/v7"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/api"
	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	"kubesphere.io/kubesphere/pkg/server/errors"
)

var bucketName = "ecpaas-images"

type handler struct {
	minioClient *minio.Client
	k8sclient   kubernetes.Interface
	ksclient    kubesphere.Interface
}

func newHandler(minioClient *minio.Client, k8sclient kubernetes.Interface, ksclient kubesphere.Interface) *handler {
	return &handler{
		minioClient: minioClient,
		k8sclient:   k8sclient,
		ksclient:    ksclient,
	}
}

type ObjectStatus struct {
	// If object exisit or not
	FileHas bool `json:"fileHas" description:"Check file exist or not"`
}

type ObjectStatusData struct {
	// Object name
	Name         string `json:"name" description:"Image file name"`
	Location     string `json:"location" description:"Image URL location"`
	LastModified string `json:"lastModified" description:"The last modified time of the image"`
	Size         int64  `json:"size" description:"Size in bytes of the image"`
}

type ImagesList struct {
	Image []ObjectStatusData `json:"items"`
}

func (h *handler) ListMinioObjects(request *restful.Request, response *restful.Response) {

	images := ImagesList{}
	minioServiceName := "minio"

	serviceList, err := h.k8sclient.CoreV1().Services("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
		return
	}

	var minioService *v1.Service

	for _, service := range serviceList.Items {
		if service.Name == minioServiceName {
			minioService = &service
			break
		}
	}

	if minioService == nil {
		klog.Warning("Cannot find the minio service ", err)
		return
	}

	ip := minioService.Spec.ClusterIP
	port := minioService.Spec.Ports[0].Port

	// Check minio bucket "ecpaas-images" if not exist then create it.
	found, err := h.minioClient.BucketExists(context.Background(), bucketName)
	if err != nil {
		api.HandleInternalError(response, request, err)
		return
	}

	if !found {
		err = h.minioClient.MakeBucket(context.Background(), bucketName, minio.MakeBucketOptions{})
		if err != nil {
			api.HandleInternalError(response, request, err)
			return
		}
	}

	objectCh := h.minioClient.ListObjects(context.Background(), bucketName, minio.ListObjectsOptions{})
	for object := range objectCh {
		if object.Err != nil {
			api.HandleInternalError(response, request, object.Err)
			return
		}

		objInfo, err := h.minioClient.StatObject(context.Background(), bucketName, object.Key, minio.StatObjectOptions{})
		if err != nil {
			klog.Warning(err)
			continue
		}

		data := ObjectStatusData{}
		data.Name = objInfo.Key
		// "location": "http://minio.kubesphere-system.svc:9000/ecpaas-images",
		data.Location = "http://" + ip + ":" + strconv.Itoa(int(port)) + "/" + bucketName
		data.LastModified = objInfo.LastModified.Format(time.RFC3339)
		data.Size = objInfo.Size
		images.Image = append(images.Image, data)
	}

	response.WriteAsJson(images)
}

func (h *handler) ListMinioObjectsWithNs(request *restful.Request, response *restful.Response) {

	images := ImagesList{}
	minioServiceName := "minio"

	serviceList, err := h.k8sclient.CoreV1().Services("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
		return
	}

	var minioService *v1.Service

	for _, service := range serviceList.Items {
		if service.Name == minioServiceName {
			minioService = &service
			break
		}
	}

	if minioService == nil {
		klog.Warning("Cannot find the minio service ", err)
		return
	}

	ip := minioService.Spec.ClusterIP
	port := minioService.Spec.Ports[0].Port

	objectCh := h.minioClient.ListObjects(context.Background(), bucketName, minio.ListObjectsOptions{})
	for object := range objectCh {
		if object.Err != nil {
			api.HandleInternalError(response, request, object.Err)
			return
		}

		objInfo, err := h.minioClient.StatObject(context.Background(), bucketName, object.Key, minio.StatObjectOptions{})
		if err != nil {
			klog.Warning(err)
			continue
		}

		data := ObjectStatusData{}
		data.Name = objInfo.Key
		// "location": "http://minio.kubesphere-system.svc:9000/ecpaas-images",
		data.Location = "http://" + ip + ":" + strconv.Itoa(int(port)) + "/" + bucketName
		data.LastModified = objInfo.LastModified.Format(time.RFC3339)
		data.Size = objInfo.Size
		images.Image = append(images.Image, data)
	}

	response.WriteAsJson(images)
}

func (h *handler) GetMinioObjectStatus(request *restful.Request, response *restful.Response) {

	imageName := request.PathParameter("imageName")
	status := ObjectStatus{}

	_, err := h.minioClient.StatObject(context.Background(), bucketName, imageName, minio.StatObjectOptions{})
	if err != nil {
		status.FileHas = false
	} else {
		status.FileHas = true
	}

	response.WriteAsJson(status)
}

func (h *handler) GetMinioObjectStatusWithNs(request *restful.Request, response *restful.Response) {

	imageName := request.PathParameter("imageName")
	status := ObjectStatus{}

	_, err := h.minioClient.StatObject(context.Background(), bucketName, imageName, minio.StatObjectOptions{})
	if err != nil {
		status.FileHas = false
	} else {
		status.FileHas = true
	}

	response.WriteAsJson(status)
}

func (h *handler) UploadMinioObject(request *restful.Request, response *restful.Response) {

	// Check minio bucket "ecpaas-images" if not exist then create it.
	found, err := h.minioClient.BucketExists(context.Background(), bucketName)
	if err != nil {
		api.HandleInternalError(response, request, err)
		return
	}

	if !found {
		err = h.minioClient.MakeBucket(context.Background(), bucketName, minio.MakeBucketOptions{})
		if err != nil {
			api.HandleInternalError(response, request, err)
			return
		}
	}

	request.Request.ParseMultipartForm(500 << 20)
	file, header, err := request.Request.FormFile("uploadfile")
	if err != nil {
		api.HandleInternalError(response, request, err)
		return
	}
	filesize := file.(Sizer).Size()

	request.Request.MultipartReader()

	uploadInfo, err := h.minioClient.PutObject(context.Background(), bucketName, header.Filename,
		file, filesize, minio.PutObjectOptions{ContentType: "application/octet-stream"})

	if err != nil {
		api.HandleInternalError(response, request, err)
		return
	}

	response.WriteAsJson(uploadInfo)
}

func (h *handler) UploadMinioObjectWithNs(request *restful.Request, response *restful.Response) {

	// Check minio bucket "ecpaas-images" if not exist then create it.
	found, err := h.minioClient.BucketExists(context.Background(), bucketName)
	if err != nil {
		api.HandleInternalError(response, request, err)
		return
	}

	if !found {
		err = h.minioClient.MakeBucket(context.Background(), bucketName, minio.MakeBucketOptions{})
		if err != nil {
			api.HandleInternalError(response, request, err)
			return
		}
	}

	request.Request.ParseMultipartForm(500 << 20)
	file, header, err := request.Request.FormFile("uploadfile")
	if err != nil {
		api.HandleInternalError(response, request, err)
		return
	}
	filesize := file.(Sizer).Size()

	request.Request.MultipartReader()

	uploadInfo, err := h.minioClient.PutObject(context.Background(), bucketName, header.Filename,
		file, filesize, minio.PutObjectOptions{ContentType: "application/octet-stream"})

	if err != nil {
		api.HandleInternalError(response, request, err)
		return
	}

	response.WriteAsJson(uploadInfo)
}

func (h *handler) DeleteMinioObject(request *restful.Request, response *restful.Response) {

	imageName := request.PathParameter("imageName")

	err := h.minioClient.RemoveObject(context.Background(), bucketName, imageName, minio.RemoveObjectOptions{})
	if err != nil {
		api.HandleInternalError(response, request, err)
		return
	}

	response.WriteEntity(errors.None)
}

func (h *handler) DeleteMinioObjectWithNs(request *restful.Request, response *restful.Response) {

	imageName := request.PathParameter("imageName")

	err := h.minioClient.RemoveObject(context.Background(), bucketName, imageName, minio.RemoveObjectOptions{})
	if err != nil {
		api.HandleInternalError(response, request, err)
		return
	}

	response.WriteEntity(errors.None)
}

type Sizer interface {
	Size() int64
}

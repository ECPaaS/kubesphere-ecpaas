/*
Copyright(c) 2025-present Accton. All rights reserved. www.accton.com.tw
*/

package v1

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/emicklei/go-restful"
	"github.com/google/uuid"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	v1 "k8s.io/api/core/v1"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	pvcviewerv1alpha1 "kubesphere.io/api/pvcviewer/v1alpha1"
)

const (
	JobTimeout = 10      // Jobs that take longer than 10 minutes are considered timed out.
	JobStatusTick = 2    // Check status of the job every 2 seconds.
	JobBackoffLimit = 0  // Retry of the job isn't allowed and the status is directly marked as failed.
	JobTtlAfterFin = 60  // When the job is completed, it will be cleared after 60 seconds.

	PVCViewerInitialDelaySeconds = 2
	PVCViewerPeriodSeconds = 10
	PVCViewerServiceTargetPort = 8080
)

type fileHandler struct {
	ksclient  kubesphere.Interface
	k8sClient kubernetes.Interface
}

func newHandler(ksclient kubesphere.Interface, k8sclient kubernetes.Interface) fileHandler {
	return fileHandler{
		ksclient:  ksclient,
		k8sClient: k8sclient,
	}
}

type FileUploadResponse struct {
	Namespace  string `json:"namespace" description:"The namespace where the PVC is located"`
	PVCName    string `json:"pvcName" description:"PVC name"`
	TargetPath string `json:"targetPath" description:"Specify the target path where the file is stored in the PVC"`
	JobName    string `json:"jobName" description:"The name of the job that uploads the file"`
}

type DownloadModelRequest struct {
	Token     string `json:"token,omitempty" description:"Hugging Face Hub access token"`
	ModelName string `json:"modelName" description:"Model name on Hugging Face Hub. Consists of username and model name (e.g. unsloth/Llama-3.2-3B-Instruct)"`
	Namespace string `json:"namespace" description:"The namespace where the PVC is located"`
	PVCName   string `json:"pvcName" description:"PVC name"`
	Directory string `json:"directory,omitempty" description:"Directory to store model files in PVC. If empty, it will be stored in a directory with the same name as the model"`
}

type DownloadModelResponse struct {
	ModelName  string `json:"modelName" description:"Model name on Hugging Face Hub"`
	Namespace  string `json:"namespace" description:"The namespace where the PVC is located"`
	PVCName    string `json:"pvcName" description:"PVC name"`
	Directory  string `json:"directory" description:"Directory to store model files in PVC"`
	JobName    string `json:"jobName" description:"The name of the job that uploads the file"`
}

type ModifyPVCViewerRequest struct {
	Enable *bool  `json:"enable" description:"Enable or disable the PVC viewer"`
}

type ListPVCViewerResponse struct {
	TotalCount int                 `json:"total_count" description:"Total number of PVC viewer informations"`
	Items      []PVCViewerResponse `json:"items" description:"List of PVC viewer informations. Key is items[].id"`
}

type PVCViewerResponse struct {
	ID        string `json:"id" description:"PVC ID"`
	Enable    bool   `json:"enable" description:"Enable or disable the PVC viewer"`
	Ready     bool   `json:"ready" description:"Ready defines if the viewer is ready to be used"`
	Namespace string `json:"namespace" description:"The namespace where the PVC is located"`
	PVCName   string `json:"pvcName" description:"PVC name"`
	ServiceIP string `json:"serviceIp" description:"The node IP and port of the node where the PVC viewer is located"`
}

func (h *fileHandler) UploadFile(req *restful.Request, resp *restful.Response) {
	// Parse HTTP multipart/form-data data.
	// All uploaded files, regardless of size, will be written to disk for temporary storage.
	err := req.Request.ParseMultipartForm(0)
	if err != nil {
		klog.Error(err)
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Extracting file fields from a multipart/form-data form
	file, _, err := req.Request.FormFile("file")
	if err != nil {
		resp.WriteError(http.StatusBadRequest, err)
		return
	}
	defer file.Close()

	// Extract the values ​​of the "namespace", "pvcName" and "targetPath" fields
	// from the multipart/form-data form.
	namespace := req.Request.FormValue("namespace")
	pvcName := req.Request.FormValue("pvcName")
	targetPath := req.Request.FormValue("targetPath")

	missingFields := []string{}
	if namespace == "" {
		missingFields = append(missingFields, "namespace")
	}
	if pvcName == "" {
		missingFields = append(missingFields, "pvcName")
	}
	if targetPath == "" {
		missingFields = append(missingFields, "targetPath")
	}
	if len(missingFields) > 0 {
		resp.WriteErrorString(http.StatusBadRequest, "Missing required fields: " + strings.Join(missingFields, ", "))
		return
	}

	// Get the absolute path location of the temporary file
	var tmpFile string
	if f, ok := file.(*os.File); ok {
		tmpFile = f.Name()
	} else {
		err := fmt.Errorf("Uploaded file is not *os.File")
		klog.Error(err)
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Uploader job needs to be located on the same node as the "ks-apiserver" pod.
	// Make sure both can mount the same hostpath /tmp directory.
	pods, _ := h.k8sClient.CoreV1().Pods("kubesphere-system").List(context.Background(), metav1.ListOptions{
		LabelSelector: "app=ks-apiserver",
	})

	var nodeName string
	if len(pods.Items) == 0 {
		nodeName = ""
	} else {
		nodeName = pods.Items[0].Spec.NodeName
	}

	// Create uploader job
	jobName := "file-upload-" + uuid.New().String()[0:6]
	jobObj := GenerateUploaderJob(jobName, nodeName, pvcName, tmpFile, targetPath)
	_, err = h.k8sClient.BatchV1().Jobs(namespace).Create(context.Background(), jobObj, metav1.CreateOptions{})
	if err != nil {
		klog.Error(err)
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Wait for the uploader job to copy the file to the target path of the specified PVC 
	// and return the upload information.
	success, err := WaitForJobCompletion(h.k8sClient, namespace, jobName)
	if success {
		uploadInfo := FileUploadResponse{
			Namespace:  namespace,
			PVCName:    pvcName,
			TargetPath: targetPath,
			JobName:    jobName,
		}
		resp.WriteEntity(uploadInfo)
	} else {
		klog.Error(err)
		resp.WriteHeader(http.StatusInternalServerError)
	}

	// After the upload job is completed, delete the local temporary file.
	req.Request.MultipartForm.RemoveAll()
}

func (h *fileHandler) DownloadModel(req *restful.Request, resp *restful.Response) {
	// Extract the parameter fields in the request
	var model DownloadModelRequest
	err := req.ReadEntity(&model)
	if err != nil {
		klog.Error(err)
		resp.WriteError(http.StatusBadRequest, err)
		return
	}

	missingFields := []string{}
	if model.ModelName == "" {
		missingFields = append(missingFields, "modelName")
	}
	if model.Namespace == "" {
		missingFields = append(missingFields, "namespace")
	}
	if model.PVCName == "" {
		missingFields = append(missingFields, "pvcName")
	}
	if len(missingFields) > 0 {
		resp.WriteErrorString(http.StatusBadRequest, "Missing required fields: " + strings.Join(missingFields, ", "))
		return
	}

	// Disable saving model files to the currently mounted directory
	if model.Directory == "/" || model.Directory == "./" {
		resp.WriteErrorString(http.StatusBadRequest, "Unable to store model file in the mounted directory")
		return
	}

	// If the directory field is empty, it will be stored in a directory with the same name as the model.
	if model.Directory == "" {
		model.Directory = filepath.Base(model.ModelName)
	}

	// Create download model job
	jobName := "download-model-" + uuid.New().String()[0:6]
	jobObj := GenerateDownloadModelJob(jobName, model.PVCName, model.Token, model.ModelName, model.Directory)
	_, err = h.k8sClient.BatchV1().Jobs(model.Namespace).Create(context.Background(), jobObj, metav1.CreateOptions{})
	if err != nil {
		klog.Error(err)
		resp.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		downloadModelInfo := DownloadModelResponse{
			ModelName: model.ModelName,
			Namespace: model.Namespace,
			PVCName:   model.PVCName,
			Directory: model.Directory,
			JobName:   jobName,
		}
		resp.WriteEntity(downloadModelInfo)
	}
}

func (h *fileHandler) PVCViewer(req *restful.Request, resp *restful.Response) {
	namespaceName := req.PathParameter("namespace")
	pvcName := req.PathParameter("pvc")
	_, err := h.k8sClient.CoreV1().PersistentVolumeClaims(namespaceName).Get(context.Background(), pvcName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			resp.WriteHeader(http.StatusNotFound)
		} else {
			resp.WriteHeader(http.StatusInternalServerError)
		}
		klog.Error(err)
		return
	}

	// Extract the parameter fields in the request
	var pvcViewer ModifyPVCViewerRequest
	err = req.ReadEntity(&pvcViewer)
	if err != nil {
		klog.Error(err)
		resp.WriteError(http.StatusBadRequest, err)
		return
	}

	if pvcViewer.Enable == nil {
		resp.WriteErrorString(http.StatusBadRequest, "Missing enable field")
		return
	}

	if *pvcViewer.Enable {
		// Add PVC viewer custom resource
		pvcViewerObj := GeneratePVCViewerCR(pvcName)
		if _, err := h.ksclient.PvcviewerV1alpha1().PVCViewers(namespaceName).Create(context.Background(), pvcViewerObj, metav1.CreateOptions{}); err != nil {
			klog.Error(err)
			resp.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		// Delete PVC viewer custom resource
		if err := h.ksclient.PvcviewerV1alpha1().PVCViewers(namespaceName).Delete(context.Background(), pvcName, metav1.DeleteOptions{}); err != nil {
			klog.Error(err)
			resp.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	resp.WriteEntity(http.StatusOK)
}

func (h *fileHandler) ListPVCViewerInfo(req *restful.Request, resp *restful.Response) {
	klog.V(2).Infof("Listing PVC viewer information")
	namespaces, err := h.k8sClient.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
    if err != nil {
		klog.Error(err)
		resp.WriteHeader(http.StatusInternalServerError)
		return
    }

	responseSlice := make([]PVCViewerResponse, 0)
	for _, ns := range namespaces.Items {
		pvcs, err := h.k8sClient.CoreV1().PersistentVolumeClaims(ns.Name).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			klog.Error(err)
			continue
		}
		for _, pvc := range pvcs.Items {
			pvcViewerEnable := false
			pvcViewerReady := false
			pvcViewerServiceIP := ""

			pvcViewer, err := h.ksclient.PvcviewerV1alpha1().PVCViewers(ns.Name).Get(context.Background(), pvc.Name, metav1.GetOptions{})
			if err != nil {
				if !apierrors.IsNotFound(err) {
					klog.Error(err)
				}
			} else {
				pvcViewerEnable = true
				pvcViewerReady = pvcViewer.Status.Ready
				if pvcViewer.Status.ServiceIP != nil {
					pvcViewerServiceIP = *pvcViewer.Status.ServiceIP
				}
			}

			response := PVCViewerResponse{
				ID:        string(pvc.UID),
				Enable:    pvcViewerEnable,
				Ready:     pvcViewerReady,
				Namespace: ns.Name,
				PVCName:   pvc.Name,
				ServiceIP: pvcViewerServiceIP,
			}
			responseSlice = append(responseSlice, response)
		}
    }
	resp.WriteEntity(ListPVCViewerResponse{TotalCount: len(responseSlice), Items: responseSlice})
}

func (h *fileHandler) GetPVCViewerInfo(req *restful.Request, resp *restful.Response) {
	klog.V(2).Infof("Get PVC viewer information")
	namespaceName := req.PathParameter("namespace")
	pvcName := req.PathParameter("pvc")

	pvc, err := h.k8sClient.CoreV1().PersistentVolumeClaims(namespaceName).Get(context.Background(), pvcName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			resp.WriteHeader(http.StatusNotFound)
		} else {
			resp.WriteHeader(http.StatusInternalServerError)
		}
		klog.Error(err)
		return
	}

	pvcViewerEnable := false
	pvcViewerReady := false
	pvcViewerServiceIP := ""

	pvcViewer, err := h.ksclient.PvcviewerV1alpha1().PVCViewers(namespaceName).Get(context.Background(), pvcName, metav1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			klog.Error(err)
			resp.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		pvcViewerEnable = true
		pvcViewerReady = pvcViewer.Status.Ready
		if pvcViewer.Status.ServiceIP != nil {
			pvcViewerServiceIP = *pvcViewer.Status.ServiceIP
		}
	}

	response := PVCViewerResponse{
		ID:        string(pvc.UID),
		Enable:    pvcViewerEnable,
		Ready:     pvcViewerReady,
		Namespace: namespaceName,
		PVCName:   pvcName,
		ServiceIP: pvcViewerServiceIP,
	}
	resp.WriteEntity(response)
}

func WaitForJobCompletion(client kubernetes.Interface, namespace, jobName string) (bool, error) {
	timeout := time.After(time.Duration(JobTimeout) * time.Minute)
	tick := time.Tick(JobStatusTick * time.Second)
	for {
		select {
		// The number of minutes for JobTimeout determines how long a job takes
		// before it's considered to have timed out.
		case <-timeout:
			return false, fmt.Errorf("Upload job timeout")

		// The interval for checking the job status is determined by the number of seconds of JobStatusTick.
		case <-tick:
			job, err := client.BatchV1().Jobs(namespace).Get(context.Background(), jobName, metav1.GetOptions{})
			if err != nil {
				return false, fmt.Errorf("Failed to get upload job: %v", err)
			}
			if job.Status.Succeeded > 0 {
				return true, nil
			} else if job.Status.Failed > 0 {
				return false, fmt.Errorf("Upload job failed")
			}
		}
	}
}

func GenerateUploaderJob(name, nodeName, pvcName, tmpFile, targetPath string) *batchv1.Job {
	filename := filepath.Base(tmpFile)
	cmd := fmt.Sprintf("mkdir -p $(dirname /ecpaas/target/%s) && cp /ecpaas/tmp/%s /ecpaas/target/%s", targetPath, filename, targetPath)
	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: int32Ptr(JobBackoffLimit),
			TTLSecondsAfterFinished: int32Ptr(JobTtlAfterFin),
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					RestartPolicy: v1.RestartPolicyNever,
					NodeName: nodeName,
					Volumes: []v1.Volume{
						{Name: "upload-pvc", VolumeSource: v1.VolumeSource{PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{ClaimName: pvcName}}},
						{Name: "local-tmp", VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: tmpFile, Type: &[]v1.HostPathType{v1.HostPathFile}[0]}}},
					},
					Containers: []v1.Container{
						{
							Name:  "uploader",
							Image: "busybox:latest",
							Command: []string{"/bin/sh", "-c", cmd},
							VolumeMounts: []v1.VolumeMount{
								{Name: "upload-pvc", MountPath: "/ecpaas/target"},
								{Name: "local-tmp", MountPath: "/ecpaas/tmp/" + filename, ReadOnly: true},
							},
							TerminationMessagePolicy: v1.TerminationMessageFallbackToLogsOnError,
						},
					},
				},
			},
		},
	}
}

func GenerateDownloadModelJob(name, pvcName, token, modelName, directory string) *batchv1.Job {
	var cloneURL string
	if token == "" {
		cloneURL = fmt.Sprintf("https://huggingface.co/%s", modelName)
	} else {
		cloneURL = fmt.Sprintf("https://ecpaas:%s@huggingface.co/%s", token, modelName)
	}

	cmd := fmt.Sprintf("git clone --depth=1 %s /ecpaas/target/%s && cd /ecpaas/target/%s && git lfs pull", cloneURL, directory, directory)
	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: int32Ptr(JobBackoffLimit),
			TTLSecondsAfterFinished: int32Ptr(JobTtlAfterFin),
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					RestartPolicy: v1.RestartPolicyNever,
					Volumes: []v1.Volume{
						{Name: "model-volume", VolumeSource: v1.VolumeSource{PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{ClaimName: pvcName}}},
					},
					Containers: []v1.Container{
						{
							Name:  "git-lfs-container",
							Image: "jgpelaez/git-lfs",
							Command: []string{"/bin/sh", "-c", cmd},
							Env: []v1.EnvVar{
								{Name:  "GIT_ASKPASS", Value: "true"},
							},
							VolumeMounts: []v1.VolumeMount{
								{Name: "model-volume", MountPath: "/ecpaas/target"},
							},
							TerminationMessagePolicy: v1.TerminationMessageFallbackToLogsOnError,
						},
					},
				},
			},
		},
	}
}

func GeneratePVCViewerCR(pvcName string) *pvcviewerv1alpha1.PVCViewer {
	return &pvcviewerv1alpha1.PVCViewer{
		ObjectMeta: metav1.ObjectMeta{Name: pvcName},
		Spec: pvcviewerv1alpha1.PVCViewerSpec{
			PVC: pvcName,
			PodSpec: v1.PodSpec{
				Volumes: []v1.Volume{
					{Name: "viewer-volume", VolumeSource: v1.VolumeSource{PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{ClaimName: pvcName}}},
				},
				Containers: []v1.Container{
					{
						Name:  "pvc-viewer",
						Image: "filebrowser/filebrowser:v2.25.0",
						Env: []v1.EnvVar{
							{Name:  "FB_ADDRESS", Value: "0.0.0.0"},
							{Name:  "FB_PORT", Value: "8080"},
							{Name:  "FB_DATABASE", Value: "/tmp/filebrowser.db"},
							{Name:  "FB_NOAUTH", Value: "true"},
						},
						VolumeMounts: []v1.VolumeMount{
							{Name: "viewer-volume", MountPath: "/srv"},
						},
						WorkingDir: "/srv",
						Resources: v1.ResourceRequirements{},
						ReadinessProbe: &v1.Probe{
							InitialDelaySeconds: PVCViewerInitialDelaySeconds,
							PeriodSeconds: PVCViewerPeriodSeconds,
							Handler: v1.Handler{
								TCPSocket: &v1.TCPSocketAction{
									Port: intstr.FromInt(PVCViewerServiceTargetPort),
								},
							},
						},
					},
				},
			},
			Networking: pvcviewerv1alpha1.Networking{
				TargetPort: intstr.FromInt(PVCViewerServiceTargetPort),
			},
			RWOScheduling: true,
		},
	}
}

func int32Ptr(i int32) *int32 { return &i }

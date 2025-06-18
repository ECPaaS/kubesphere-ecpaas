/*
Copyright(c) 2025-present Accton. All rights reserved. www.accton.com.tw
*/

package v1alpha1

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/emicklei/go-restful"
	"github.com/google/uuid"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	v1 "k8s.io/api/core/v1"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	JobTimeout = 10      // Jobs that take longer than 10 minutes are considered timed out.
	JobStatusTick = 2    // Check status of the job every 2 seconds.
	JobBackoffLimit = 3  // If the job still cannot be executed after 3 retries, the status will be marked as failed.
	JobTtlAfterFin = 10  // When the job is completed, it will be cleared after 10 seconds.
)

type fileHandler struct {
	k8sClient kubernetes.Interface
}

func newHandler(k8sclient kubernetes.Interface) fileHandler {
	return fileHandler{
		k8sClient: k8sclient,
	}
}

type FileUploadInfo struct {
	Namespace  string `json:"namespace" description:"The namespace where the PVC is located"`
	PvcName    string `json:"pvcName" description:"PVC name"`
	TargetPath string `json:"targetPath" description:"Specify the target path where the file is stored in the PVC"`
	JobName    string `json:"jobName" description:"The name of the job that uploads the file"`
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
		err := errors.New("Uploaded file is not *os.File")
		klog.Error(err)
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Create uploader job
	jobName := "file-upload-" + uuid.New().String()[0:6]
	jobObj := GenerateUploaderJob(jobName, pvcName, namespace, tmpFile, targetPath)

	_, err = h.k8sClient.BatchV1().Jobs(namespace).Create(context.TODO(), jobObj, metav1.CreateOptions{})
	if err != nil {
		klog.Error(err)
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Wait for the uploader job to copy the file to the target path of the specified PVC 
	// and return the upload information.
	success, err := WaitForJobCompletion(h.k8sClient, namespace, jobName)
	if success {
		uploadInfo := FileUploadInfo{
			Namespace:  namespace,
			PvcName:    pvcName,
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

func WaitForJobCompletion(client kubernetes.Interface, namespace, jobName string) (bool, error) {
	timeout := time.After(time.Duration(JobTimeout) * time.Minute)
	tick := time.Tick(JobStatusTick * time.Second)
	for {
		select {
		// The number of minutes for JobTimeout determines how long a job takes
		// before it's considered to have timed out.
		case <-timeout:
			return false, errors.New("Upload job timeout")

		// The interval for checking the job status is determined by the number of seconds of JobStatusTick.
		case <-tick:
			job, err := client.BatchV1().Jobs(namespace).Get(context.TODO(), jobName, metav1.GetOptions{})
			if err != nil {
				return false, errors.New(fmt.Sprintf("Failed to get upload job: %v", err))
			}
			if job.Status.Succeeded > 0 {
				return true, nil
			} else if job.Status.Failed > 0 {
				return false, errors.New("Upload job failed")
			}
		}
	}
}

func GenerateUploaderJob(name, pvcName, namespace, tmpFile, targetPath string) *batchv1.Job {
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

func int32Ptr(i int32) *int32 { return &i }

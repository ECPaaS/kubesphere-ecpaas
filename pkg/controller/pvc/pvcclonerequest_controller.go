/*
Copyright(c) 2025-present Accton. All rights reserved. www.accton.com.tw
*/

package pvc

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"

	pvcv1 "kubesphere.io/api/pvc/v1"
)

const (
	controllerName = "pvc-clone-request-controller"
	requestPhaseCloning = "Cloning"
	requestPhaseCompleted = "Completed"
	podImage = "alpine:latest"
	containerName = "cloner"
	sourcePodSuffix = "-cloner-"
	sourcePodCommand = "apk add --no-cache rsync > /dev/null; echo -e '[source]\npath = /source' >> /etc/rsyncd.conf; rsync --daemon --no-detach"
	targetPodSuffix = "-creator-"
	targetPodCommand = "apk add --no-cache rsync > /dev/null; until rsync rsync://%s 2> /dev/null; do :; sleep 2; done; rsync -avh rsync://%s/source /target/ || :"
	serviceSuffix = "-service-"
	serviceSelector = "rsync-server"
	rsyncPort = 873
	pvcAnnotationKey = "ecpaas.io/pvc-clone-result"
)

// Reconciler reconciles a PVCCloneRequest object
type Reconciler struct {
	client.Client
	Logger                  logr.Logger
	Recorder                record.EventRecorder
	MaxConcurrentReconciles int
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	if r.Client == nil {
		r.Client = mgr.GetClient()
	}
	if r.Logger == nil {
		r.Logger = ctrl.Log.WithName("controllers").WithName(controllerName)
	}
	if r.Recorder == nil {
		r.Recorder = mgr.GetEventRecorderFor(controllerName)
	}
	if r.MaxConcurrentReconciles <= 0 {
		r.MaxConcurrentReconciles = 1
	}
	return ctrl.NewControllerManagedBy(mgr).
		Named(controllerName).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: r.MaxConcurrentReconciles,
		}).
		For(&pvcv1.PVCCloneRequest{}).
		Complete(r)
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	klog.V(2).Infof("Reconciling PVCCloneRequest %s/%s", req.Namespace, req.Name)

	rootCtx := context.Background()
	cloneRequest := &pvcv1.PVCCloneRequest{}
	if err := r.Get(rootCtx, req.NamespacedName, cloneRequest); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	switch cloneRequest.Status.Phase {
	case "":
		// New clone request, create target PVC, pods, service
		if err := r.createTargetPVC(cloneRequest); err != nil {
			klog.Infof("error creating target PVC, abort")
			r.changePhase(requestPhaseCompleted, cloneRequest)
			return ctrl.Result{}, err
		}
		if err := r.createClonePods(cloneRequest); err != nil {
			klog.Infof("error creating pods for cloning, abort")
			r.changePhase(requestPhaseCompleted, cloneRequest)
			r.updatePVCAnnotation(cloneRequest.Spec.TargetPVCName, cloneRequest.Spec.TargetPVCNamespace, "Clone failed, no data is cloned")
			return ctrl.Result{}, err
		}
		if err := r.createCloneService(cloneRequest); err != nil {
			klog.Infof("error creating service for cloning, abort")
			r.changePhase(requestPhaseCompleted, cloneRequest)
			r.updatePVCAnnotation(cloneRequest.Spec.TargetPVCName, cloneRequest.Spec.TargetPVCNamespace, "Clone failed, no data is cloned")
			return ctrl.Result{}, err
		}

		// Change phase to "Cloning"
		return ctrl.Result{}, r.changePhase(requestPhaseCloning, cloneRequest)
	case requestPhaseCloning:
		// If target pod completes, save log and mark completed.
		// If target pod being deleted, interrupt cloning and clean up.
		targetPod := &corev1.Pod{}
		targetPodName := cloneRequest.Spec.TargetPVCName + targetPodSuffix + string(cloneRequest.UID) // <targetPVC>-creator-<UID>
		if err := r.Get(rootCtx, types.NamespacedName{Name: targetPodName, Namespace: cloneRequest.Spec.TargetPVCNamespace}, targetPod); err != nil {
			if errors.IsNotFound(err) {
				klog.Infof("target pod for cloning is deleted, abort")
				r.changePhase(requestPhaseCompleted, cloneRequest)
				r.updatePVCAnnotation(cloneRequest.Spec.TargetPVCName, cloneRequest.Spec.TargetPVCNamespace, "Clone interrupted, no data is cloned")
				return ctrl.Result{}, nil
			} else {
				return ctrl.Result{}, err
			}
		}
		if targetPod.Status.Phase == corev1.PodSucceeded {
			// Save log
			if err := r.updateLogs(targetPodName, targetPod.Namespace, cloneRequest); err != nil {
				klog.Infof("error getting clone result, continue")
			}

			// Change phase to "Completed"
			return ctrl.Result{}, r.changePhase(requestPhaseCompleted, cloneRequest)
		} else {
			// Wait for cloning process
			duration := time.Duration(2)
			return ctrl.Result{RequeueAfter: duration}, nil
		}
	case requestPhaseCompleted:
		// Clean up everything
		r.deleteClonePods(cloneRequest)
		r.deleteCloneService(cloneRequest)
		r.Delete(rootCtx, cloneRequest)
		return ctrl.Result{}, nil
	default:
		// Log and change phase to "Completed"
		klog.Infof("invalid phase: \"%s\", change to \"Completed\" to clean up", cloneRequest.Status.Phase)
		return ctrl.Result{}, r.changePhase(requestPhaseCompleted, cloneRequest)
	}
}

func (r *Reconciler) createTargetPVC(cloneRequest *pvcv1.PVCCloneRequest) error {
	newPVC := &corev1.PersistentVolumeClaim{}
	newPVC.Name = cloneRequest.Spec.TargetPVCName
	newPVC.Namespace = cloneRequest.Spec.TargetPVCNamespace
	newPVC.Spec.AccessModes = createAccessModes(cloneRequest.Spec.AccessModes)
	newPVC.Spec.Resources.Requests = createResourceList(cloneRequest.Spec.Size)
	if err := r.Create(context.Background(), newPVC); err != nil {
		return err
	} else {
		return nil
	}
}

func createAccessModes(modes []string) []corev1.PersistentVolumeAccessMode {
	returnArray := make([]corev1.PersistentVolumeAccessMode, 0)
	for _, mode := range modes {
		returnArray = append(returnArray, corev1.PersistentVolumeAccessMode(mode))
	}
	return returnArray
}

func createResourceList(size int) corev1.ResourceList {
	returnMap := make(map[corev1.ResourceName]resource.Quantity, 0)
	returnMap[corev1.ResourceStorage] = resource.MustParse(fmt.Sprint(size) + "Gi")
	return returnMap
}

func (r *Reconciler) createClonePods(cloneRequest *pvcv1.PVCCloneRequest) error {
	// Source Pod
	sourcePod := &corev1.Pod{}
	sourcePod.Name = cloneRequest.Spec.SourcePVCName + sourcePodSuffix + string(cloneRequest.UID) // <sourcePVC>-cloner-<UID>
	sourcePod.Namespace = cloneRequest.Spec.SourcePVCNamespace
	sourcePod.Spec.Containers = []corev1.Container{
		{
			Name: containerName,
			Image: podImage,
			ImagePullPolicy: corev1.PullIfNotPresent,
			Command: []string{"/bin/sh", "-c"},
			Args: []string{sourcePodCommand},
			VolumeMounts: []corev1.VolumeMount{
				{
					Name: "source",
					MountPath: "/source",
				},
			},
		},
	}
	sourcePod.Spec.Volumes = []corev1.Volume{
		{
			Name: "source",
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: cloneRequest.Spec.SourcePVCName,
				},
			},
		},
	}
	sourcePod.Labels = map[string]string{
		serviceSelector: sourcePod.Name, // so the service can match this exact pod
	}
	if err := r.Create(context.Background(), sourcePod); err != nil {
		return err
	}

	// Target Pod
	targetPod := &corev1.Pod{}
	targetPod.Name = cloneRequest.Spec.TargetPVCName + targetPodSuffix + string(cloneRequest.UID) // <targetPVC>-creator-<UID>
	targetPod.Namespace = cloneRequest.Spec.TargetPVCNamespace
	targetPod.Spec.RestartPolicy = corev1.RestartPolicyNever
	// <pvc>-service-<UID>.<namespace>
	serviceNamespacedName := cloneRequest.Spec.SourcePVCName + serviceSuffix + string(cloneRequest.UID) + "." + cloneRequest.Spec.SourcePVCNamespace
	targetPod.Spec.Containers = []corev1.Container{
		{
			Name: containerName,
			Image: podImage,
			ImagePullPolicy: corev1.PullIfNotPresent,
			Command: []string{"/bin/sh", "-c"},
			Args: []string{fmt.Sprintf(targetPodCommand, serviceNamespacedName, serviceNamespacedName)},
			VolumeMounts: []corev1.VolumeMount{
				{
					Name: "target",
					MountPath: "/target",
				},
			},
		},
	}
	targetPod.Spec.Volumes = []corev1.Volume{
		{
			Name: "target",
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: cloneRequest.Spec.TargetPVCName,
				},
			},
		},
	}
	targetPod.Spec.NodeSelector = cloneRequest.Spec.NodeSelector
	if err := r.Create(context.Background(), targetPod); err != nil {
		return err
	}

	return nil
}

func (r *Reconciler) createCloneService(cloneRequest *pvcv1.PVCCloneRequest) error {
	service := &corev1.Service{}
	service.Name = cloneRequest.Spec.SourcePVCName + serviceSuffix + string(cloneRequest.UID) // <sourcePVC>-service-<UID>
	service.Namespace = cloneRequest.Spec.SourcePVCNamespace
	service.Spec.Type = corev1.ServiceTypeClusterIP
	service.Spec.Ports = []corev1.ServicePort{
		{
			Port: rsyncPort,
			TargetPort: intstr.IntOrString{IntVal: rsyncPort},
		},
	}
	service.Spec.Selector = map[string]string{
		serviceSelector: cloneRequest.Spec.SourcePVCName + sourcePodSuffix + string(cloneRequest.UID), // <sourcePVC>-cloner-<UID>
	}
	if err := r.Create(context.Background(), service); err != nil {
		return err
	} else {
		return nil
	}
}

func (r *Reconciler) deleteClonePods(cloneRequest *pvcv1.PVCCloneRequest) {
	// Source Pod
	sourcePod := &corev1.Pod{}
	sourcePodName := cloneRequest.Spec.SourcePVCName + sourcePodSuffix + string(cloneRequest.UID) // <sourcePVC>-cloner-<UID>
	err := r.Get(context.Background(), types.NamespacedName{Name: sourcePodName, Namespace: cloneRequest.Spec.SourcePVCNamespace}, sourcePod)
	if err == nil {
		if err := r.Delete(context.Background(), sourcePod); err != nil {
			klog.Infof("error deleting source pod \"%s\" in \"%s\": %s", sourcePod.Name, sourcePod.Namespace, err.Error())
		}
	} else if !errors.IsNotFound(err) {
		// Ignore Not Found error, log other errors
		klog.Infof("error getting source pod \"%s\" in \"%s\": %s", sourcePod.Name, sourcePod.Namespace, err.Error())
	}

	// Target Pod
	targetPod := &corev1.Pod{}
	targetPodName := cloneRequest.Spec.TargetPVCName + targetPodSuffix + string(cloneRequest.UID) // <targetPVC>-creator-<UID>
	err = r.Get(context.Background(), types.NamespacedName{Name: targetPodName, Namespace: cloneRequest.Spec.TargetPVCNamespace}, targetPod)
	if err == nil {
		if err := r.Delete(context.Background(), targetPod); err != nil {
			klog.Infof("error deleting target pod \"%s\" in \"%s\": %s", targetPod.Name, targetPod.Namespace, err.Error())
		}
	} else if !errors.IsNotFound(err) {
		// Ignore Not Found error, log other errors
		klog.Infof("error getting target pod \"%s\" in \"%s\": %s", targetPod.Name, targetPod.Namespace, err.Error())
	}
}

func (r *Reconciler) deleteCloneService(cloneRequest *pvcv1.PVCCloneRequest) {
	service := &corev1.Service{}
	serviceName := cloneRequest.Spec.SourcePVCName + serviceSuffix + string(cloneRequest.UID) // <sourcePVC>-service-<UID>
	err := r.Get(context.Background(), types.NamespacedName{Name: serviceName, Namespace: cloneRequest.Spec.SourcePVCNamespace}, service)
	if err == nil {
		if err := r.Delete(context.Background(), service); err != nil {
			klog.Infof("error deleting clone service \"%s\" in \"%s\": %s", service.Name, service.Namespace, err.Error())
		}
	} else if !errors.IsNotFound(err) {
		// Ignore Not Found error, log other errors
		klog.Infof("error getting clone service \"%s\" in \"%s\": %s", service.Name, service.Namespace, err.Error())
	}
}

func (r *Reconciler) updateLogs(podName string, podNamespace string, cloneRequest *pvcv1.PVCCloneRequest) error {
	restConfig := ctrl.GetConfigOrDie()
	kubeClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		klog.Infof("error getting client for logs, give up logs")
		return err
	}
	podLogOpts := &corev1.PodLogOptions{
		Container: containerName,
		Follow: false,
	}
	req := kubeClient.CoreV1().Pods(podNamespace).GetLogs(podName, podLogOpts)
	readCloser, err := req.Stream(context.Background())
	if err != nil {
		klog.Infof("error opening stream for logs, give up logs: %s", err.Error())
		return err
	}
	defer readCloser.Close()

	logData, err := io.ReadAll(readCloser)
	if err != nil {
		klog.Infof("error reading logs, give up logs: %s", err.Error())
		return err
	}

	err = r.updatePVCAnnotation(cloneRequest.Spec.TargetPVCName, cloneRequest.Spec.TargetPVCNamespace, string(logData))
	if err != nil {
		klog.Infof("error updating PVC, give up logs: %s", err.Error())
		return err
	}

	return nil
}

func (r *Reconciler) updatePVCAnnotation(name string, namespace string, content string) error {
	pvc := &corev1.PersistentVolumeClaim{}
	err := r.Get(context.Background(), types.NamespacedName{Name: name, Namespace: namespace}, pvc)
	if err != nil {
		klog.Infof("error getting PVC: %s", err.Error())
		return err
	}
	pvc.Annotations[pvcAnnotationKey] = content
	err = r.Update(context.Background(), pvc)
	if err != nil {
		klog.Infof("error updating PVC: %s", err.Error())
		return err
	}
	return nil
}

func (r *Reconciler) changePhase(phase string, cloneRequest *pvcv1.PVCCloneRequest) error {
	cloneRequest.Status.Phase = phase
	return r.Status().Update(context.Background(), cloneRequest)
}

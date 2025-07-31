/*
 Copyright(c) 2025-present Accton. All rights reserved. www.accton.com.tw
 */

package pvcviewer

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	pvcviewerv1 "kubesphere.io/api/pvcviewer/v1"
)

// PVCViewerReconciler reconciles a PVCViewer object
type PVCViewerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

const (
	// We use a resource prefix so that the names of generated resources like deployments are unique
	resourcePrefix = "pvcviewer-"

	nameLabelKey     = "app.kubernetes.io/name"
	instanceLabelKey = "app.kubernetes.io/instance"
	partOfLabelKey   = "app.kubernetes.io/part-of"
	partOfLabelValue = "pvc-viewer"

	servicePort         = int32(80)
)

// SetupWithManager sets up the controller with the Manager.
func (r *PVCViewerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if r.Client == nil {
		r.Client = mgr.GetClient()
	}
	if r.Scheme == nil {
		r.Scheme = mgr.GetScheme()
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&pvcviewerv1.PVCViewer{}).
		// This controller manages, i.e. creates these kinds for a PVCViewer
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Complete(r)
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *PVCViewerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	instance := &pvcviewerv1.PVCViewer{}
	if err := r.Get(ctx, req.NamespacedName, instance); err != nil {
		// Created objects are automatically garbage collected if parent is deleted
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	commonLabels := map[string]string{
		nameLabelKey:     instance.Name,
		instanceLabelKey: resourcePrefix + instance.Name,
		partOfLabelKey:   partOfLabelValue,
	}

	if !instance.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is being deleted
		// Do nothing as the resources are automatically garbage collected
		klog.Info("PVCViewer is being deleted")

		// Keep on reconciling status until the finalizer is removed
		if err := r.reconcileStatus(ctx, instance.Name, instance.Namespace, commonLabels); err != nil {
			klog.Errorf("Error while reconciling status: %v", err)
			return ctrl.Result{}, err
		}

		return reconcile.Result{}, nil
	}

	if err := r.reconcileDeployment(ctx, instance, commonLabels); err != nil {
		klog.Errorf("Error while reconciling deployment: %v", err)
		return ctrl.Result{}, err
	}

	if err := r.reconcileService(ctx, instance, commonLabels); err != nil {
		klog.Errorf("Error while reconciling service: %v", err)
		return ctrl.Result{}, err
	}

	if err := r.reconcileStatus(ctx, instance.Name, instance.Namespace, commonLabels); err != nil {
		klog.Errorf("Error while reconciling status: %v", err)
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// Creates or updates the deployment as defined by the viewer's podSpec
func (r *PVCViewerReconciler) reconcileDeployment(ctx context.Context, viewer *pvcviewerv1.PVCViewer, commonLabels map[string]string) error {
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      resourcePrefix + viewer.Name,
			Namespace: viewer.Namespace,
			Labels:    commonLabels,
		},
	}
	createDeployment := false
	if err := r.Get(ctx, types.NamespacedName{Name: deployment.Name, Namespace: deployment.Namespace}, deployment); err != nil {
		if !apierrs.IsNotFound(err) {
			return err
		}
		createDeployment = true
	}

	original := deployment.DeepCopy()

	var (
		// Do not change affinity or rwoClaims by default
		affinity = deployment.Spec.Template.Spec.Affinity
		// Affinity is only to be set when rwo scheduling is enabled and the deployment is to be newly created
		determineAffinity = viewer.Spec.RWOScheduling && createDeployment
	)

	if determineAffinity {
		if newAffinity, err := r.generateAffinity(ctx, viewer); err != nil {
			return err
		} else if newAffinity != nil {
			// Only set the affinity if it is not nil - we wouldn't win anything by restarting without affinity
			affinity = newAffinity
		}
	}

	deployment.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: commonLabels,
	}
	deployment.Spec.Template = corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: commonLabels,
		},
		Spec: viewer.Spec.PodSpec,
	}
	// We're using a recreate strategy to ensure that the pod is restarted when the affinity change.
	// Otherwise, we could be mounting the same PVC to multiple pods, preventing the pod from starting.
	deployment.Spec.Strategy = appsv1.DeploymentStrategy{
		Type: appsv1.RecreateDeploymentStrategyType,
	}
	deployment.Spec.Template.Spec.Affinity = affinity

	if err := ctrl.SetControllerReference(viewer, deployment, r.Scheme); err != nil {
		return err
	}

	if createDeployment {
		klog.Info("Creating Deployment")
		return r.Create(ctx, deployment)
	}
	klog.V(2).Info("Updating Deployment")
	return r.Patch(ctx, deployment, client.MergeFrom(original))
}

// Creates or updates the service as defined by the viewer's service
func (r *PVCViewerReconciler) reconcileService(ctx context.Context, viewer *pvcviewerv1.PVCViewer, commonLabels map[string]string) error {
	if viewer.Spec.Networking == (pvcviewerv1.Networking{}) {
		return nil
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      resourcePrefix + viewer.Name,
			Namespace: viewer.Namespace,
			Labels:    commonLabels,
		},
	}
	createService := false
	if err := r.Get(ctx, types.NamespacedName{Name: service.Name, Namespace: service.Namespace}, service); err != nil {
		if !apierrs.IsNotFound(err) {
			return err
		}
		createService = true
	}

	original := service.DeepCopy()

	service.Spec.Type = "NodePort"
	service.Spec.Selector = commonLabels
	service.Spec.Ports = []corev1.ServicePort{
		{
			Name:       "http",
			Port:       servicePort,
			TargetPort: viewer.Spec.Networking.TargetPort,
		},
	}

	if err := ctrl.SetControllerReference(viewer, service, r.Scheme); err != nil {
		return err
	}

	if createService {
		klog.Info("Creating Service")
		return r.Create(ctx, service)
	}
	klog.V(2).Info("Updating Service")
	return r.Patch(ctx, service, client.MergeFrom(original))
}

// Computes and updates the status of the PVCViewer
func (r *PVCViewerReconciler) reconcileStatus(ctx context.Context, viewerName string, viewerNamespace string, commonLabels map[string]string) error {
	viewer := &pvcviewerv1.PVCViewer{}
	if err := r.Get(ctx, types.NamespacedName{Name: viewerName, Namespace: viewerNamespace}, viewer); err != nil {
		return err
	}

	viewer.Status.ServiceIP = r.generateServiceIP(ctx, viewer, commonLabels)
	if viewer.Status.ServiceIP == nil {
		viewer.Status.Ready = false
		return nil
	}

	deployment := &appsv1.Deployment{}
	if err := r.Get(ctx, types.NamespacedName{Name: resourcePrefix + viewer.Name, Namespace: viewer.Namespace}, deployment); err != nil {
		klog.Info("Could not find Deployment for status update")
		viewer.Status.Ready = false
	} else {
		viewer.Status.Ready = *deployment.Spec.Replicas == deployment.Status.ReadyReplicas
		// Append the latest condition, if it is not already in the list
		if len(deployment.Status.Conditions) > 0 {
			clen := len(viewer.Status.Conditions)
			if clen == 0 || viewer.Status.Conditions[clen-1] != deployment.Status.Conditions[0] {
				viewer.Status.Conditions = append(viewer.Status.Conditions, deployment.Status.Conditions[0])
			}
		}
	}

	klog.V(2).Info("Updating status")
	return r.Client.Status().Update(ctx, viewer)
}

// Generates the affinity to be used for the deployment
// In case no affinity should be used (e.g. RWOScheduling is disabled) or updated, nil is returned
func (r *PVCViewerReconciler) generateAffinity(ctx context.Context, viewer *pvcviewerv1.PVCViewer) (*corev1.Affinity, error) {
	// Check if the viewer's PVC is RWO access mode
	pvc := &corev1.PersistentVolumeClaim{}
	if err := r.Get(ctx, types.NamespacedName{Name: viewer.Spec.PVC, Namespace: viewer.Namespace}, pvc); err != nil {
		if apierrs.IsNotFound(err) {
			klog.Info("Omitting Affinity: PVC not found")
			// Should we return an error here or suppress it and let the Deployment fail?
			// Latter might be better and more visible to the user
			return nil, nil
		}
		return nil, err
	}

	if len(pvc.Spec.AccessModes) != 1 || pvc.Spec.AccessModes[0] != corev1.ReadWriteOnce {
		klog.Info("Omitting Affinity: PVC is not RWO")
		return nil, nil
	}

	// Get all pods in namespace and filter by RWO PVCs
	podList := &corev1.PodList{}
	if err := r.List(ctx, podList, client.InNamespace(viewer.Namespace)); err != nil {
		return nil, err
	}
	var nodeName *string
	for _, pod := range podList.Items {
		// Skip pods this controller created
		if partOf, ok := pod.Labels[partOfLabelKey]; ok && partOf == partOfLabelValue {
			continue
		}
		for _, volume := range pod.Spec.Volumes {
			if volume.PersistentVolumeClaim != nil && volume.PersistentVolumeClaim.ClaimName != "" {
				if volume.PersistentVolumeClaim.ClaimName == pvc.Name {
					if nodeName != nil {
						// Rather than throwing an error, we just omit the affinity, leaving the current deployment's affinity unchanged
						klog.Info("Omitting Affinity: Viewer references RWO volumes on multiple nodes",
							"nodes", []string{*nodeName, pod.Spec.NodeName})
						return nil, nil
					}
					if pod.Spec.NodeName == "" {
						klog.Info("Omitting Affinity: Viewer references RWO volume on pod without nodeName")
						return nil, nil
					}
					nodeName = &pod.Spec.NodeName
				}
			}
		}
	}

	if nodeName == nil {
		klog.Info("Omitting Affinity: PVC not used by other Pods")
		return nil, nil
	}

	// Generate Affinity using the node name
	affinity := &corev1.Affinity{
		NodeAffinity: &corev1.NodeAffinity{
			PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{
				{
					Weight: 100,
					Preference: corev1.NodeSelectorTerm{
						MatchExpressions: []corev1.NodeSelectorRequirement{
							{
								Key:      "kubernetes.io/hostname",
								Operator: "In",
								Values:   []string{*nodeName},
							},
						},
					},
				},
			},
		},
	}
	return affinity, nil
}

func (r *PVCViewerReconciler) generateServiceIP(ctx context.Context, viewer *pvcviewerv1.PVCViewer, commonLabels map[string]string) *string {
	service := &corev1.Service{}
	if err := r.Get(ctx, types.NamespacedName{Name: resourcePrefix + viewer.Name, Namespace: viewer.Namespace}, service); err != nil {
		klog.Info("Could not find Service for status update")
		return nil
	} else {
		// Find the Pod corresponding to the Deployment
		podList := &corev1.PodList{}
		if err := r.List(ctx, podList, client.InNamespace(viewer.Namespace),
			client.MatchingLabels(commonLabels)); err != nil || len(podList.Items) == 0 {
			klog.Info("No Pod found for Deployment")
			return nil
		} else {
			pod := podList.Items[0]
			nodeName := pod.Spec.NodeName

			// Query Node IP based on Node name
			node := &corev1.Node{}
			if err := r.Get(ctx, types.NamespacedName{Name: nodeName}, node); err != nil {
				klog.Infof("Failed to get Node: %v", err)
				return nil
			} else {
				var nodeIP string
				// Get only the internal IP of the node
				for _, addr := range node.Status.Addresses {
					if addr.Type == corev1.NodeInternalIP {
						nodeIP = addr.Address
						break
					}
				}

				// Combining Node IP and Node port
				if nodeIP != "" && len(service.Spec.Ports) > 0 {
					nodePort := service.Spec.Ports[0].NodePort
					ip := fmt.Sprintf("%s:%d", nodeIP, nodePort)
					return &ip
				} else {
					klog.Info("Node IP or port not found")
					return nil
				}
			}
		}
	}
}
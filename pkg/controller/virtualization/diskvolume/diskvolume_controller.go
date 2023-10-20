/*
Copyright(c) 2023-present Accton. All rights reserved. www.accton.com.tw
*/

package diskvolume

import (
	"context"
	"fmt"
	"reflect"

	"github.com/go-logr/logr"
	"github.com/spf13/pflag"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	storagev1 "k8s.io/api/storage/v1"

	virtzv1alpha1 "kubesphere.io/api/virtualization/v1alpha1"
	"kubevirt.io/client-go/kubecli"
	cdiv1 "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"
)

const (
	controllerName        = "diskvolume-controller"
	successSynced         = "Synced"
	messageResourceSynced = "DiskVolume synced successfully"
	pvcNamePrefix         = "tpl-" // tpl: template
)

// Reconciler reconciles a disk volume object
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
		For(&virtzv1alpha1.DiskVolume{}).
		Complete(r)
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	klog.V(2).Infof("Reconciling DiskVolume %s/%s", req.Namespace, req.Name)

	rootCtx := context.Background()
	dv := &virtzv1alpha1.DiskVolume{}
	if err := r.Get(rootCtx, req.NamespacedName, dv); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	clientConfig := kubecli.DefaultClientConfig(&pflag.FlagSet{})

	// get the kubevirt client, using which kubevirt resources can be managed.
	virtClient, err := kubecli.GetKubevirtClientFromClientConfig(clientConfig)
	if err != nil {
		klog.Infof("Cannot obtain KubeVirt client: %v\n", err)
		return ctrl.Result{}, err
	}

	// get default storage class name
	scName := ""
	scList := &storagev1.StorageClassList{}
	if err := r.List(rootCtx, scList); err != nil {
		return ctrl.Result{}, err
	}
	for _, sc := range scList.Items {
		if sc.Annotations["storageclass.kubernetes.io/is-default-class"] == "true" {
			scName = sc.Name
			break
		}
	}
	if scName == "" {
		return ctrl.Result{}, fmt.Errorf("no default storage class found")
	}

	dv_instance := dv.DeepCopy()

	status := &dv_instance.Status
	if !status.Created {

		if dv_instance.Spec.PVCName == "" {
			dv_instance.Spec.PVCName = pvcNamePrefix + dv_instance.Name
		}

		// create pvc for blank disk
		if dv_instance.Spec.Source.Blank != nil {
			err := r.createPVC(dv_instance, scName)
			if err != nil {
				statusErr := err.(*errors.StatusError)
				if statusErr.ErrStatus.Reason == metav1.StatusReasonAlreadyExists {
					klog.Infof("PVC %s/%s already exists", dv_instance.Namespace, dv_instance.Spec.PVCName)
				} else {
					klog.Infof("Cannot create PVC: %v\n", err)
					return ctrl.Result{}, err
				}
			}
		}
		// clone pvc for image disk
		if dv_instance.Spec.Source.Image != nil {
			err := r.clonePVC(virtClient, dv_instance, scName)
			if err != nil {
				klog.Infof("Cannot clone PVC: %v\n", err)
				return ctrl.Result{}, err
			}
		}

		status.Created = true

	}

	if !reflect.DeepEqual(dv, dv_instance) {
		if err := r.Update(rootCtx, dv_instance); err != nil {
			return ctrl.Result{}, err
		}
	}

	if dv_instance.OwnerReferences != nil {
		status.Owner = dv_instance.OwnerReferences[0].Name
	} else if dv_instance.Labels[virtzv1alpha1.VirtualizationDiskVolumeOwner] != "" {
		status.Owner = dv_instance.Labels[virtzv1alpha1.VirtualizationDiskVolumeOwner]
	} else {
		status.Owner = ""
	}

	// update status
	status.Ready = true
	if err := r.Status().Update(rootCtx, dv_instance); err != nil {
		return ctrl.Result{}, err
	}

	// update event
	r.Recorder.Event(dv, corev1.EventTypeNormal, successSynced, messageResourceSynced)
	return ctrl.Result{}, nil

}

func (r *Reconciler) createPVC(dv_instance *virtzv1alpha1.DiskVolume, scName string) error {
	klog.Infof("Creating pvc %s/%s", dv_instance.Namespace, dv_instance.Spec.PVCName)

	blockOwnerDeletion := true
	controller := true

	pvc := &corev1.PersistentVolumeClaim{}
	pvc.Name = dv_instance.Spec.PVCName
	pvc.Namespace = dv_instance.Namespace
	pvc.Spec.AccessModes = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}
	pvc.Spec.Resources = corev1.ResourceRequirements{}
	pvc.Spec.Resources.Requests = corev1.ResourceList{}
	pvc.Spec.Resources.Requests[corev1.ResourceStorage] = dv_instance.Spec.Resources.Requests[corev1.ResourceStorage]
	pvc.Spec.StorageClassName = &scName
	// owner reference
	pvc.OwnerReferences = []metav1.OwnerReference{
		{
			APIVersion:         dv_instance.APIVersion,
			BlockOwnerDeletion: &blockOwnerDeletion,
			Controller:         &controller,
			Kind:               dv_instance.Kind,
			Name:               dv_instance.Name,
			UID:                dv_instance.UID,
		},
	}

	if err := r.Create(context.Background(), pvc); err != nil {
		return err
	}

	klog.Infof("PVC %s/%s created", pvc.Namespace, pvc.Name)

	return nil
}

func (r *Reconciler) clonePVC(virtClient kubecli.KubevirtClient, dv_instance *virtzv1alpha1.DiskVolume, scName string) error {
	klog.Infof("Cloning pvc %s/%s", dv_instance.Namespace, dv_instance.Spec.Source.Image.Name)

	blockOwnerDeletion := true
	controller := true

	dv := &cdiv1.DataVolume{
		TypeMeta: metav1.TypeMeta{
			Kind: "DataVolume",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        dv_instance.Spec.PVCName,
			Namespace:   dv_instance.Namespace,
			Annotations: dv_instance.Annotations,
			Labels:      dv_instance.Labels,
		},
		Spec: cdiv1.DataVolumeSpec{
			PVC: &corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{
					corev1.ReadWriteOnce,
				},
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: dv_instance.Spec.Resources.Requests[corev1.ResourceStorage],
					},
				},
			},
			Source: &cdiv1.DataVolumeSource{
				PVC: &cdiv1.DataVolumeSourcePVC{
					Name:      dv_instance.Spec.Source.Image.Name,
					Namespace: dv_instance.Spec.Source.Image.Namespace,
				},
			},
		},
	}
	dv.OwnerReferences = []metav1.OwnerReference{
		{
			APIVersion:         dv_instance.APIVersion,
			BlockOwnerDeletion: &blockOwnerDeletion,
			Controller:         &controller,
			Kind:               dv_instance.Kind,
			Name:               dv_instance.Name,
			UID:                dv_instance.UID,
		},
	}

	if _, err := virtClient.CdiClient().CdiV1beta1().DataVolumes(dv_instance.Namespace).Create(context.Background(), dv, metav1.CreateOptions{}); err != nil {
		klog.Infof("Cannot create DataVolume: %v\n", err)
		return err
	}

	klog.Infof("DataVolume %s/%s created", dv.Namespace, dv.Name)

	return nil
}

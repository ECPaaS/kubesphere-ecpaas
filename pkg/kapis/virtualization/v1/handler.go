package virtualization

import (
	"net/http"

	"github.com/emicklei/go-restful"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog"

	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"

	virtz "kubesphere.io/kubesphere/pkg/models/virtualization"
)

type virtzhandler struct {
	virtz virtz.Interface
}

func newHandler(ksclient kubesphere.Interface) virtzhandler {
	return virtzhandler{
		virtz: virtz.New(ksclient),
	}
}

func (h virtzhandler) CreateVirtualMahcine(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")

	var virtz_vm virtz.VirtualMachine
	err := req.ReadEntity(&virtz_vm)
	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	_, err = h.virtz.CreateVirtualMachine(namespace, &virtz_vm)
	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	resp.WriteHeader(http.StatusOK)
}

func (h virtzhandler) DeleteVirtualMachine(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")
	vmName := req.PathParameter("virtualmachine")

	_, err := h.virtz.DeleteVirtualMachine(namespace, vmName)
	if err != nil {
		klog.Error(err)
		if errors.IsNotFound(err) {
			resp.WriteError(http.StatusNotFound, err)
			return
		}
		resp.WriteError(http.StatusInternalServerError, err)
		return
	}

	resp.WriteHeader(http.StatusOK)
}

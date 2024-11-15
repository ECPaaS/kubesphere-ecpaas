/*
Copyright(c) 2024-present Accton. All rights reserved. www.accton.com.tw
*/

package v1

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/emicklei/go-restful"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
)

type handler struct {
	k8sclient kubernetes.Interface
	ksclient  kubesphere.Interface
}

func newHandler(k8sclient kubernetes.Interface, ksclient kubesphere.Interface) *handler {
	return &handler{
		k8sclient: k8sclient,
		ksclient:  ksclient,
	}
}

type YunikornQueuesResponse struct {
	TotalCount int             `json:"total_count" description:"Total number of queues"`
	Items      []YunikornQueue `json:"items" description:"Available yunikorn queues"`
}

type YunikornQueue struct {
	Queue string `json:"queue"  description:"Available yunikorn queue"`
}

type YunikornPatition struct {
	Name string `json:"name"  description:"Partition name"`
}

type YunikornQueues struct {
	QueueName     string           `json:"queuename"  description:"Queue name"`
	IsLeaf        bool             `json:"isLeaf"  description:"Is leaf queue"`
	Children      []YunikornQueues `json:"children"  description:"queue's children"`
	ChildrenNames []string         `json:"childrenNames"  description:"Array of children's name"`
}

type SchedulerNameResponse struct {
	TotalCount int             `json:"total_count" description:"Total number of scheduler"`
	Items      []SchedulerName `json:"items" description:"Available schedulers name"`
}

type SchedulerName struct {
	Name string `json:"name"  description:"Available scheduler name"`
}

func (h *handler) ListSchedulerName(request *restful.Request, response *restful.Response) {
	schedulers := []SchedulerName{
		{
			Name: "default-scheduler",
		},
	}

	// Yunikorn scheduler
	if isYunikornAvailable(h) {
		schedulers = append(schedulers, SchedulerName{
			Name: "yunikorn",
		})
	}

	schedulerResponse := SchedulerNameResponse{
		TotalCount: len(schedulers),
		Items:      schedulers,
	}

	response.WriteAsJson(schedulerResponse)
}

func isYunikornAvailable(h *handler) bool {

	yunikornServiceDNS, err := getYuniKornServiceName(h)
	if err != nil {
		return false
	}

	req, _ := http.NewRequest(http.MethodGet, "http://"+yunikornServiceDNS+"/ws/v1/partitions", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		klog.Error(err.Error())
		return false
	}

	defer resp.Body.Close()

	return resp.StatusCode == 200
}

func (h *handler) ListYuniKornQueues(request *restful.Request, response *restful.Response) {

	yunikornServiceDNS, err := getYuniKornServiceName(h)
	if err != nil {
		klog.Error(err.Error())
		if errors.IsNotFound(err) {
			response.WriteError(http.StatusNotFound, err)
		}
		response.WriteError(http.StatusInternalServerError, err)
	}

	partitionNames, err := getAllPartition(yunikornServiceDNS)
	if err != nil {
		klog.Error(err.Error())
		if errors.IsNotFound(err) {
			response.WriteError(http.StatusNotFound, err)
		}
		response.WriteError(http.StatusInternalServerError, err)
	}

	queues, err := getAllLeafQueues(yunikornServiceDNS, partitionNames)

	if err != nil {
		klog.Error(err.Error())
		if errors.IsNotFound(err) {
			response.WriteError(http.StatusNotFound, err)
		}
		response.WriteError(http.StatusInternalServerError, err)
	}
	yunikornQueue := []YunikornQueue{}

	for _, queuename := range queues {
		yunikornQueue = append(yunikornQueue, YunikornQueue{Queue: queuename})
	}

	queuesResponse := YunikornQueuesResponse{
		TotalCount: len(yunikornQueue),
		Items:      yunikornQueue,
	}

	response.WriteAsJson(queuesResponse)
}

func getAllPartition(yunikornServiceDNS string) ([]string, error) {
	req, _ := http.NewRequest(http.MethodGet, "http://"+yunikornServiceDNS+"/ws/v1/partitions", nil)
	resp, err := http.DefaultClient.Do(req)
	partitionNames := []string{}

	if err != nil {
		klog.Error(err.Error())
		return partitionNames, err
	}

	defer resp.Body.Close()

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		klog.Error(err.Error())
		return partitionNames, err
	}

	partitions := []YunikornPatition{}

	err = json.Unmarshal(content, &partitions)

	if err != nil {
		klog.Error(err.Error())
		return partitionNames, err
	}

	for _, partition := range partitions {
		partitionNames = append(partitionNames, partition.Name)
	}
	return partitionNames, nil
}

func getAllLeafQueues(yunikornServiceDNS string, partitionNames []string) ([]string, error) {

	queues := []string{}

	for _, partition := range partitionNames {
		req, _ := http.NewRequest(http.MethodGet, "http://"+yunikornServiceDNS+"/ws/v1/partition/"+partition+"/queues", nil)
		resp, err := http.DefaultClient.Do(req)

		if err != nil {
			klog.Error(err.Error())
			return queues, err
		}

		defer resp.Body.Close()

		content, err := io.ReadAll(resp.Body)
		if err != nil {
			klog.Error(err.Error())
			return queues, err
		}

		yunikornQueues := YunikornQueues{}

		err = json.Unmarshal(content, &yunikornQueues)

		if err != nil {
			klog.Error(err.Error())
			return queues, err
		}

		if !yunikornQueues.IsLeaf && len(yunikornQueues.Children) > 0 {
			queues = getLeafQueues(yunikornQueues.Children, queues)
		}
	}
	return queues, nil
}

func getLeafQueues(children []YunikornQueues, leafQueues []string) []string {

	for _, child := range children {
		if child.IsLeaf && len(child.Children) == 0 {
			leafQueues = append(leafQueues, child.QueueName)
		} else {
			leafQueues = getLeafQueues(child.Children, leafQueues)
		}
	}
	return leafQueues
}

func getYuniKornServiceName(h *handler) (string, error) {

	yunikornService, err := h.k8sclient.CoreV1().Services("yunikorn").Get(context.Background(), "yunikorn-service", metav1.GetOptions{})

	if err != nil {
		return "", err
	}

	var yunikornCorePort int
	for _, port := range yunikornService.Spec.Ports {
		if port.Name == "yunikorn-core" {
			yunikornCorePort = int(port.Port)
		}
	}

	yunikornServiceDNS := "yunikorn-service.yunikorn:" + strconv.Itoa(yunikornCorePort)
	return yunikornServiceDNS, nil
}

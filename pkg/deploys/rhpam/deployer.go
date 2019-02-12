package rhpam

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/golang/glog"
	brokerapi "github.com/integr8ly/managed-service-broker/pkg/broker"
	"github.com/integr8ly/managed-service-broker/pkg/clients/openshift"
	"github.com/integr8ly/managed-service-broker/pkg/deploys/rhpam/pkg/apis/rhpam/v1alpha1"
	k8sClient "github.com/operator-framework/operator-sdk/pkg/k8sclient"
	"github.com/operator-framework/operator-sdk/pkg/util/k8sutil"
	"github.com/pkg/errors"
	"k8s.io/api/authentication/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type RhpamDeployer struct {
	k8sClient kubernetes.Interface
	osClient  *openshift.ClientFactory
}

func NewDeployer(k8sClient kubernetes.Interface, osClient *openshift.ClientFactory) *RhpamDeployer {
	return &RhpamDeployer{
		k8sClient: k8sClient,
		osClient:  osClient,
	}
}

func (rd *RhpamDeployer) GetCatalogEntries() []*brokerapi.Service {
	glog.Infof("Getting rhpam catalog entries")
	return getCatalogServicesObj()
}

func (rd *RhpamDeployer) Deploy(req *brokerapi.ProvisionRequest, async bool) (*brokerapi.ProvisionResponse, error) {
	glog.Infof("Deploying rhpam from deployer, id: %s", req.InstanceId)

	// Namespace
	ns, err := rd.k8sClient.CoreV1().Namespaces().Create(getNamespaceObj("rhpam-" + req.InstanceId))
	if err != nil {
		glog.Errorf("failed to create rhpam namespace: %+v", err)
		return &brokerapi.ProvisionResponse{
			Code: http.StatusInternalServerError,
		}, errors.Wrap(err, "failed to create namespace for rhpam service")
	}

	namespace := ns.ObjectMeta.Name

	// ServiceAccount
	_, err = rd.k8sClient.CoreV1().ServiceAccounts(namespace).Create(getServiceAccountObj())
	if err != nil {
		glog.Errorf("failed to create rhpam service account: %+v", err)
		return &brokerapi.ProvisionResponse{
			Code: http.StatusInternalServerError,
		}, errors.Wrap(err, "failed to create service account for rhpam service")
	}

	//Role
	_, err = rd.k8sClient.RbacV1beta1().Roles(namespace).Create(getRoleObj())
	if err != nil {
		glog.Errorf("failed to create rhpamrole: %+v", err)
		return &brokerapi.ProvisionResponse{
			Code: http.StatusInternalServerError,
		}, errors.Wrap(err, "failed to create role for rhpam service")
	}

	// RoleBindings
	err = rd.createRoleBindings(namespace, req.OriginatingUserInfo, rd.k8sClient, rd.osClient)
	if err != nil {
		glog.Errorln(err)
		return &brokerapi.ProvisionResponse{
			Code: http.StatusInternalServerError,
		}, err
	}

	// DeploymentConfig
	err = rd.createRhpamOperator(namespace, rd.osClient)
	if err != nil {
		glog.Errorln(err)
		return &brokerapi.ProvisionResponse{
			Code: http.StatusInternalServerError,
		}, err
	}

	// Rhpam custom resource
	cr := rd.createRhpamCustomResourceTemplate(namespace)
	if err := rd.createRhpamCustomResource(namespace, cr); err != nil {
		glog.Errorln(err)
		return &brokerapi.ProvisionResponse{
			Code: http.StatusInternalServerError,
		}, err
	}

	return &brokerapi.ProvisionResponse{
		Code:         http.StatusAccepted,
		DashboardURL: "https://rhpam-bc-" + rd.getRouteHostname(namespace),
		Operation:    "deploy",
	}, nil
}

func (rd *RhpamDeployer) RemoveDeploy(req *brokerapi.DeprovisionRequest, async bool) (*brokerapi.DeprovisionResponse, error) {
	ns := "rhpam-" + req.InstanceId
	err := rd.k8sClient.CoreV1().Namespaces().Delete(ns, &metav1.DeleteOptions{})
	if err != nil && !strings.Contains(err.Error(), "not found") {
		glog.Errorf("failed to delete %s namespace: %+v", ns, err)
		return &brokerapi.DeprovisionResponse{}, errors.Wrap(err, fmt.Sprintf("failed to delete namespace %s", ns))
	} else if err != nil && strings.Contains(err.Error(), "not found") {
		glog.Infof("rhpam namespace already deleted")
	}
	return &brokerapi.DeprovisionResponse{Operation: "remove"}, nil
}

func (rd *RhpamDeployer) ServiceInstanceLastOperation(req *brokerapi.LastOperationRequest) (*brokerapi.LastOperationResponse, error) {
	glog.Infof("Getting last operation for %s", req.InstanceId)

	namespace := "rhpam-" + req.InstanceId
	switch req.Operation {
	case "deploy":
		cr, err := getRhpam(namespace)
		if err != nil {
			return nil, err
		}
		if cr == nil {
			return nil, apiErrors.NewNotFound(v1alpha1.SchemeGroupResource, req.InstanceId)
		}

		if cr.Status.Phase == v1alpha1.PhaseComplete {
			return &brokerapi.LastOperationResponse{
				State:       brokerapi.StateSucceeded,
				Description: "rhpam deployed successfully",
			}, nil
		}

		return &brokerapi.LastOperationResponse{
			State:       brokerapi.StateInProgress,
			Description: "rhpam is deploying",
		}, nil

	case "remove":
		_, err := rd.k8sClient.CoreV1().Namespaces().Get(namespace, metav1.GetOptions{})
		if err != nil && apiErrors.IsNotFound(err) {
			return &brokerapi.LastOperationResponse{
				State:       brokerapi.StateSucceeded,
				Description: "rhpam has been deleted",
			}, nil
		}

		return &brokerapi.LastOperationResponse{
			State:       brokerapi.StateInProgress,
			Description: "rhpam is deleting",
		}, nil
	default:
		cr, err := getRhpam(namespace)
		if err != nil {
			return nil, err
		}
		if cr == nil {
			return nil, apiErrors.NewNotFound(v1alpha1.SchemeGroupResource, req.InstanceId)
		}

		return &brokerapi.LastOperationResponse{
			State:       brokerapi.StateFailed,
			Description: "unknown operation: " + req.Operation,
		}, nil
	}
}

func (rd *RhpamDeployer) createRoleBindings(namespace string, userInfo v1.UserInfo, k8sclient kubernetes.Interface, osClientFactory *openshift.ClientFactory) error {
	for _, sysRoleBinding := range getSystemRoleBindings(namespace) {
		_, err := k8sclient.RbacV1beta1().RoleBindings(namespace).Create(&sysRoleBinding)
		if err != nil && !strings.Contains(err.Error(), "already exists") {
			return errors.Wrapf(err, "failed to create rolebinding for %s", sysRoleBinding.ObjectMeta.Name)
		}
	}

	_, err := k8sclient.RbacV1beta1().RoleBindings(namespace).Create(getInstallRoleBindingObj())
	if err != nil {
		return errors.Wrap(err, "failed to create install role binding for rhpam service")
	}

	authClient, err := osClientFactory.AuthClient()
	if err != nil {
		return errors.Wrap(err, "failed to create an openshift authorization client")
	}

	_, err = authClient.RoleBindings(namespace).Create(getViewRoleBindingObj())
	if err != nil {
		return errors.Wrap(err, "failed to create view role binding for rhpam service")
	}

	_, err = authClient.RoleBindings(namespace).Create(getEditRoleBindingObj())
	if err != nil {
		return errors.Wrap(err, "failed to create edit role binding for rhpam service")
	}

	_, err = authClient.RoleBindings(namespace).Create(getUserViewRoleBindingObj(namespace, userInfo.Username))
	if err != nil {
		return errors.Wrap(err, "failed to create user view role binding for rhpam service")
	}

	return nil
}

func (rd *RhpamDeployer) createRhpamOperator(namespace string, osClientFactory *openshift.ClientFactory) error {
	dcClient, err := osClientFactory.AppsClient()
	if err != nil {
		return errors.Wrap(err, "failed to create an openshift deployment config client")
	}

	_, err = dcClient.DeploymentConfigs(namespace).Create(getDeploymentConfigObj())
	if err != nil {
		return errors.Wrap(err, "failed to create deployment config for rhpam service")
	}

	return nil
}

// Create the rhpam custom resource template
func (rd *RhpamDeployer) createRhpamCustomResourceTemplate(namespace string) *v1alpha1.RhpamDev {
	return getRhpamObj(namespace)
}

// Get route hostname for rhpam
func (rd *RhpamDeployer) getRouteHostname(namespace string) string {
	routeHostname := namespace
	routeSuffix, exists := os.LookupEnv("ROUTE_SUFFIX")
	if exists {
		routeHostname = routeHostname + "." + routeSuffix
	}
	return routeHostname
}

// Create the rhpam custom resource
func (rd *RhpamDeployer) createRhpamCustomResource(namespace string, rhpam *v1alpha1.RhpamDev) error {
	client, _, err := k8sClient.GetResourceClient(v1alpha1.ApiVersion(), "RhpamDev", namespace)
	if err != nil {
		return err
	}

	_, err = client.Create(k8sutil.UnstructuredFromRuntimeObject(rhpam))
	if err != nil {
		return err
	}

	return nil
}

// Get rhpam resource in namespace
func getRhpam(namespace string) (*v1alpha1.RhpamDev, error) {
	client, _, err := k8sClient.GetResourceClient(v1alpha1.ApiVersion(), "RhpamDev", namespace)
	if err != nil {
		return nil, err
	}

	u, err := client.List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	rl := v1alpha1.NewRhpamDevList()
	if err := k8sutil.RuntimeObjectIntoRuntimeObject(u, rl); err != nil {
		return nil, errors.Wrap(err, "failed to get the rhpam resources")
	}

	for _, r := range rl.Items {
		return &r, nil
	}

	return nil, nil
}

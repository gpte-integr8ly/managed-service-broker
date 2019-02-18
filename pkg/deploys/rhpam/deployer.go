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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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

	// Role
	_, err = rd.k8sClient.RbacV1beta1().Roles(namespace).Create(getRoleObj())
	if err != nil {
		glog.Errorf("failed to create rhpamrole: %+v", err)
		return &brokerapi.ProvisionResponse{
			Code: http.StatusInternalServerError,
		}, errors.Wrap(err, "failed to create role for rhpam service")
	}

	// Role for user
	_, err = rd.k8sClient.RbacV1beta1().Roles(namespace).Create(getUserRoleObj())
	if err != nil {
		glog.Errorf("failed to create rhpamrole: %+v", err)
		return &brokerapi.ProvisionResponse{
			Code: http.StatusInternalServerError,
		}, errors.Wrap(err, "failed to create role for rhpam service user")
	}

	// RoleBindings
	err = rd.createRoleBindings(namespace, req.OriginatingUserInfo, rd.k8sClient, rd.osClient)
	if err != nil {
		glog.Errorln(err)
		return &brokerapi.ProvisionResponse{
			Code: http.StatusInternalServerError,
		}, err
	}

	// Clusterrole
	_, err = rd.k8sClient.RbacV1beta1().ClusterRoles().Create(getClusterRoleObj(namespace))
	if err != nil {
		glog.Errorf("failed to create rhpam clusterrole: %+v", err)
		return &brokerapi.ProvisionResponse{
			Code: http.StatusInternalServerError,
		}, errors.Wrap(err, "failed to create clusterrole for rhpam service")
	}

	// ClusterRoleBinding
	_, err = rd.k8sClient.RbacV1beta1().ClusterRoleBindings().Create(getClusterRoleBindingObj(namespace))
	if err != nil {
		glog.Errorf("failed to create rhpam clusterrolebinding: %+v", err)
		return &brokerapi.ProvisionResponse{
			Code: http.StatusInternalServerError,
		}, errors.Wrap(err, "failed to create clusterrolebinding for rhpam service")
	}

	// DeploymentConfig
	err = rd.createRhpamOperator(namespace, rd.osClient)
	if err != nil {
		glog.Errorln(err)
		return &brokerapi.ProvisionResponse{
			Code: http.StatusInternalServerError,
		}, err
	}

	// Rhpam custom resources
	rhpamdevCr := rd.createRhpamDevCustomResourceTemplate(namespace)
	if err := rd.createRhpamDevCustomResource(namespace, rhpamdevCr); err != nil {
		glog.Errorln(err)
		return &brokerapi.ProvisionResponse{
			Code: http.StatusInternalServerError,
		}, err
	}

	rhpamuserCr := rd.createRhpamUserCustomResourceTemplate(namespace)
	if err := rd.createRhpamUserCustomResource(namespace, rhpamuserCr); err != nil {
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
	//TODO: remove clusterrole, clusterrolebinding, rhpadev user objec, rhpamdev object
	glog.Info("Deleting rhpamuser resources in namespace ", ns)
	err := rd.deleteRhpamUserCustomResources(ns)
	if err != nil {
		return &brokerapi.DeprovisionResponse{}, errors.Wrap(err, fmt.Sprintf("failed to delete rhpamuser resources in namespace %s", ns))
	}

	glog.Info("Deleting rhpamdev resources in namespace ", ns)
	err = rd.deleteRhpamDevCustomResources(ns)
	if err != nil {
		return &brokerapi.DeprovisionResponse{}, errors.Wrap(err, fmt.Sprintf("failed to delete rhpamdev resources in namespace %s", ns))
	}

	// Clusterrole
	err = rd.k8sClient.RbacV1beta1().ClusterRoleBindings().Delete("rhpam-dev-operator-"+ns, &metav1.DeleteOptions{})
	if err != nil {
		glog.Errorf("failed to delete clusterrolebinding: %+v", err)
		return &brokerapi.DeprovisionResponse{}, errors.Wrap(err, "failed to delete clusterrolebinding for rhpam service")
	}

	// ClusterRoleBinding
	err = rd.k8sClient.RbacV1beta1().ClusterRoles().Delete("rhpam-dev-operator-"+ns, &metav1.DeleteOptions{})
	if err != nil {
		glog.Errorf("failed to delete rhpam clusterrole: %+v", err)
		return &brokerapi.DeprovisionResponse{}, errors.Wrap(err, "failed to delete clusterrole for rhpam service")
	}

	err = rd.k8sClient.CoreV1().Namespaces().Delete(ns, &metav1.DeleteOptions{})
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
		cr, err := getRhpamDev(namespace)
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
		cr, err := getRhpamDev(namespace)
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

	_, err = k8sclient.RbacV1beta1().RoleBindings(namespace).Create(getUserRoleBindingObj(namespace, userInfo.Username))
	if err != nil {
		return errors.Wrap(err, "failed to create user role binding for rhpam service")
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

	_, err = authClient.RoleBindings(namespace).Create(getUserEditRoleBindingObj(namespace, userInfo.Username))
	if err != nil {
		return errors.Wrap(err, "failed to create user edit role binding for rhpam service")
	}

	return nil
}

func (rd *RhpamDeployer) createClusterRoleBinding(namespace string, k8sclient kubernetes.Interface) error {
	_, err := k8sclient.RbacV1beta1().ClusterRoleBindings().Create(getClusterRoleBindingObj(namespace))
	if err != nil {
		return errors.Wrap(err, "failed to create install cluster role binding for rhpam service")
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

// Create the rhpam dev custom resource template
func (rd *RhpamDeployer) createRhpamDevCustomResourceTemplate(namespace string) *v1alpha1.RhpamDev {
	return getRhpamDevObj(namespace)
}

// Create the rhpam user custom resource template
func (rd *RhpamDeployer) createRhpamUserCustomResourceTemplate(namespace string) *v1alpha1.RhpamUser {
	return getRhpamUserObj(namespace)
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

// Create the rhpam dev custom resource
func (rd *RhpamDeployer) createRhpamDevCustomResource(namespace string, rhpam *v1alpha1.RhpamDev) error {
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

// Delete the rhpam dev Custom Resource
func (rd *RhpamDeployer) deleteRhpamDevCustomResource(namespace string, rhpam string) error {
	client, _, err := k8sClient.GetResourceClient(v1alpha1.ApiVersion(), "RhpamDev", namespace)
	if err != nil {
		return err
	}
	err = client.Delete(rhpam, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}

// delete the rhpam dev Custom Resources
func (rd *RhpamDeployer) deleteRhpamDevCustomResources(namespace string) error {
	client, _, err := k8sClient.GetResourceClient(v1alpha1.ApiVersion(), "RhpamDev", namespace)
	if err != nil {
		return err
	}
	list, err := client.List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	// cast to unstructured
	rhpamdevs := list.(*unstructured.UnstructuredList)
	for _, u := range rhpamdevs.Items {
		if err := rd.deleteRhpamDevCustomResource(namespace, u.GetName()); err != nil {
			return err
		}
	}
	return nil
}

// Delete the rhpam user Custom Resource
func (rd *RhpamDeployer) deleteRhpamUserCustomResource(namespace string, rhpamuser string) error {
	client, _, err := k8sClient.GetResourceClient(v1alpha1.ApiVersion(), "RhpamUser", namespace)
	if err != nil {
		return err
	}
	err = client.Delete(rhpamuser, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}

// delete the rhpam dev Custom Resources
func (rd *RhpamDeployer) deleteRhpamUserCustomResources(namespace string) error {
	client, _, err := k8sClient.GetResourceClient(v1alpha1.ApiVersion(), "RhpamUser", namespace)
	if err != nil {
		return err
	}
	list, err := client.List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	// cast to unstructured
	rhpamusers := list.(*unstructured.UnstructuredList)
	for _, u := range rhpamusers.Items {
		if err := rd.deleteRhpamUserCustomResource(namespace, u.GetName()); err != nil {
			return err
		}
	}
	return nil
}

// Create the rhpam user custom resource
func (rd *RhpamDeployer) createRhpamUserCustomResource(namespace string, rhpam *v1alpha1.RhpamUser) error {
	client, _, err := k8sClient.GetResourceClient(v1alpha1.ApiVersion(), "RhpamUser", namespace)
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
func getRhpamDev(namespace string) (*v1alpha1.RhpamDev, error) {
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

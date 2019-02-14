package rhpam

import (
	"os"

	brokerapi "github.com/integr8ly/managed-service-broker/pkg/broker"
	"github.com/integr8ly/managed-service-broker/pkg/deploys/rhpam/pkg/apis/rhpam/v1alpha1"
	appsv1 "github.com/openshift/api/apps/v1"
	authv1 "github.com/openshift/api/authorization/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1beta1 "k8s.io/api/rbac/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	RHPAM_OPERATOR_IMAGE_STREAMS_NAMESPACE string = "openshift"
	RHPAM_OPERATOR_IMAGE_STREAM_NAME       string = "rhpam-dev-operator:v0.0.2"
)

var (
	SSO_NAMESPACE                string = lookupEnv("SSO_NAMESPACE")
	SSO_ADMIN_CREDENTIALS_SECRET string = lookupEnv("SSO_ADMIN_CREDENTIALS_SECRET")
)

func getCatalogServicesObj() []*brokerapi.Service {
	return []*brokerapi.Service{
		{
			Name:        "rhpam-dev",
			ID:          "rhpam-service-id",
			Description: "rhpam-dev",
			Metadata: map[string]string{
				"serviceName": "rhpam-dev",
				"serviceType": "rhpam-dev",
			},
			Plans: []brokerapi.ServicePlan{
				{
					Name:        "default-rhpam",
					ID:          "default-rhpam",
					Description: "default rhpam plan",
					Free:        true,
					Schemas: &brokerapi.Schemas{
						ServiceBinding: &brokerapi.ServiceBindingSchema{
							Create: &brokerapi.RequestResponseSchema{},
						},
						ServiceInstance: &brokerapi.ServiceInstanceSchema{
							Create: &brokerapi.InputParametersSchema{},
						},
					},
				},
			},
		},
	}
}

func getNamespaceObj(id string) *corev1.Namespace {
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: id,
		},
	}
}

// Rhpam operator service account
func getServiceAccountObj() *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name: "rhpam-dev-operator",
		},
	}
}

// Rhpam operator clusterrole
func getClusterRoleObj(namespace string) *rbacv1beta1.ClusterRole {
	return &rbacv1beta1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: "rhpam-dev-operator-" + namespace,
		},
		Rules: []rbacv1beta1.PolicyRule{
			{
				APIGroups: []string{"rhpam.integreatly.org"},
				Resources: []string{"rhpamdevs", "rhpamdevs/finalizers", "rhpamusers", "rhpamusers/finalizers"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"pods", "services", "endpoints", "persistentvolumeclaims", "configmaps", "secrets", "serviceaccounts"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{"apps.openshift.io"},
				Resources: []string{"deploymentconfigs"},
				Verbs:     []string{"get", "list", "watch"},
			},
		},
	}
}

func getClusterRoleBindingObj(namespace string) *rbacv1beta1.ClusterRoleBinding {
	return &rbacv1beta1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "rhpam-dev-operator-" + namespace,
		},
		RoleRef: rbacv1beta1.RoleRef{
			Name:     "rhpam-dev-operator-" + namespace,
			Kind:     "ClusterRole",
			APIGroup: "rbac.authorization.k8s.io",
		},
		Subjects: []rbacv1beta1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "rhpam-dev-operator",
				Namespace: namespace,
			},
		},
	}
}

// Rhpam operator role
func getRoleObj() *rbacv1beta1.Role {
	return &rbacv1beta1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name: "rhpam-dev-operator",
		},
		Rules: []rbacv1beta1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"pods", "services", "endpoints", "persistentvolumeclaims", "configmaps", "secrets", "serviceaccounts"},
				Verbs:     []string{"create", "delete", "deletecollection", "get", "list", "update", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"events"},
				Verbs:     []string{"get", "list"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"namespaces"},
				Verbs:     []string{"get"},
			},
			{
				APIGroups: []string{"apps.openshift.io"},
				Resources: []string{"deploymentconfigs"},
				Verbs:     []string{"create", "delete", "deletecollection", "get", "list", "update", "watch"},
			},
			{
				APIGroups: []string{"route.openshift.io"},
				Resources: []string{"routes"},
				Verbs:     []string{"create", "delete", "deletecollection", "get", "list", "update", "watch"},
			},
			{
				APIGroups: []string{"rhpam.integreatly.org"},
				Resources: []string{"rhpamdevs", "rhpamdevs/finalizers", "rhpamusers", "rhpamusers/finalizers"},
				Verbs:     []string{"create", "delete", "deletecollection", "get", "list", "update", "watch"},
			},
		},
	}
}

// System specific role bindings
func getSystemRoleBindings(namespace string) []rbacv1beta1.RoleBinding {
	return []rbacv1beta1.RoleBinding{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "system:deployers",
			},
			Subjects: []rbacv1beta1.Subject{
				{
					Kind:      "ServiceAccount",
					Name:      "deployer",
					Namespace: namespace,
				},
			},
			RoleRef: rbacv1beta1.RoleRef{
				Kind:     "ClusterRole",
				Name:     "system:deployer",
				APIGroup: "rbac.authorization.k8s.io",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "system:image-builders",
			},
			Subjects: []rbacv1beta1.Subject{
				{
					Kind:      "ServiceAccount",
					Name:      "builder",
					Namespace: namespace,
				},
			},
			RoleRef: rbacv1beta1.RoleRef{
				Kind:     "ClusterRole",
				Name:     "system:image-builder",
				APIGroup: "rbac.authorization.k8s.io",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "system:image-pullers",
			},
			Subjects: []rbacv1beta1.Subject{
				{
					Kind:      "Group",
					Name:      "system:serviceaccounts:" + namespace,
					Namespace: namespace,
				},
			},
			RoleRef: rbacv1beta1.RoleRef{
				Kind:     "ClusterRole",
				Name:     "system:image-puller",
				APIGroup: "rbac.authorization.k8s.io",
			},
		},
	}
}

// Rhpam specific role bindings
func getInstallRoleBindingObj() *rbacv1beta1.RoleBinding {
	return &rbacv1beta1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "rhpam-dev-operator:install",
		},
		Subjects: []rbacv1beta1.Subject{
			{
				Kind: "ServiceAccount",
				Name: "rhpam-dev-operator",
			},
		},
		RoleRef: rbacv1beta1.RoleRef{
			Kind:     "Role",
			Name:     "rhpam-dev-operator",
			APIGroup: "rbac.authorization.k8s.io",
		},
	}
}

func getViewRoleBindingObj() *authv1.RoleBinding {
	return &authv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "rhpam-dev-operator:view",
		},
		Subjects: []corev1.ObjectReference{
			{
				Kind: "ServiceAccount",
				Name: "rhpam-dev-operator",
			},
		},
		RoleRef: corev1.ObjectReference{
			Name: "view",
		},
	}
}

func getEditRoleBindingObj() *authv1.RoleBinding {
	return &authv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "rhpam-dev-operator:edit",
		},
		Subjects: []corev1.ObjectReference{
			{
				Kind: "ServiceAccount",
				Name: "rhpam-dev-operator",
			},
		},
		RoleRef: corev1.ObjectReference{
			Name: "edit",
		},
	}
}

func getUserViewRoleBindingObj(namespace, username string) *authv1.RoleBinding {
	return &authv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "rhpam-dev-operator:view-",
			Namespace:    namespace,
		},
		RoleRef: corev1.ObjectReference{
			Name: "view",
		},
		Subjects: []corev1.ObjectReference{
			{
				Kind: "User",
				Name: username,
			},
		},
	}
}

// Rhpam operator deployment config
func getDeploymentConfigObj() *appsv1.DeploymentConfig {
	return &appsv1.DeploymentConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: "rhpam-dev-operator",
		},
		Spec: appsv1.DeploymentConfigSpec{
			Strategy: appsv1.DeploymentStrategy{
				Type: "Recreate",
			},
			Replicas: 1,
			Selector: map[string]string{
				"name": "rhpam-dev-operator",
			},
			Template: &corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"name": "rhpam-dev-operator",
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: "rhpam-dev-operator",
					Containers: []corev1.Container{
						{
							Name:            "rhpam-dev-operator",
							Image:           " ",
							ImagePullPolicy: "IfNotPresent",
							Env: []corev1.EnvVar{
								{
									Name: "WATCH_NAMESPACE",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
								},
								{
									Name: "POD_NAME",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.name",
										},
									},
								},
								{
									Name:  "OPERATOR_NAME",
									Value: "rhpam-dev-operator",
								},
								{
									Name:  "SSO_NAMESPACE",
									Value: SSO_NAMESPACE,
								},
								{
									Name:  "SSO_ADMIN_CREDENTIALS_SECRET",
									Value: SSO_ADMIN_CREDENTIALS_SECRET,
								},
							},
							Ports: []corev1.ContainerPort{
								{
									Name:          "metrics",
									ContainerPort: 60000,
								},
							},
							Command: []string{"rhpam-dev-operator"},
							ReadinessProbe: &corev1.Probe{
								InitialDelaySeconds: 4,
								PeriodSeconds:       10,
								FailureThreshold:    1,
								Handler: corev1.Handler{
									Exec: &corev1.ExecAction{
										Command: []string{"stat", "/tmp/operator-sdk-ready"},
									},
								},
							},
						},
					},
				},
			},
			Triggers: appsv1.DeploymentTriggerPolicies{
				appsv1.DeploymentTriggerPolicy{
					ImageChangeParams: &appsv1.DeploymentTriggerImageChangeParams{
						Automatic: true,
						ContainerNames: []string{
							"rhpam-dev-operator",
						},
						From: corev1.ObjectReference{
							Kind:      "ImageStreamTag",
							Name:      RHPAM_OPERATOR_IMAGE_STREAM_NAME,
							Namespace: RHPAM_OPERATOR_IMAGE_STREAMS_NAMESPACE,
						},
					},
					Type: "ImageChange",
				},
				appsv1.DeploymentTriggerPolicy{
					Type: "ConfigChange",
				},
			},
		},
	}
}

// Rhpam dev Custom Resource
func getRhpamDevObj(deployNamespace string) *v1alpha1.RhpamDev {
	return &v1alpha1.RhpamDev{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RhpamDev",
			APIVersion: "rhpam.integreatly.org/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace:    deployNamespace,
			GenerateName: "rhpamdev-",
		},
		Spec: v1alpha1.RhpamDevSpec{
			Domain: lookupEnv("ROUTE_SUFFIX"),
		},
	}
}

// Rhpam user Custom Resource
func getRhpamUserObj(deployNamespace string) *v1alpha1.RhpamUser {
	return &v1alpha1.RhpamUser{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RhpamUser",
			APIVersion: "rhpam.integreatly.org/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace:    deployNamespace,
			GenerateName: "rhpamuser-",
		},
		Spec: v1alpha1.RhpamUserSpec{
			Roles: []*v1alpha1.Role{&v1alpha1.Role{Name: "group1"}, &v1alpha1.Role{Name: "group2"}},
			Users: []*v1alpha1.User{&v1alpha1.User{Username: "user1", Password: "password", Roles: []string{"user", "kie-server", "group1"}},
				&v1alpha1.User{Username: "user2", Password: "password", Roles: []string{"user", "kie-server", "group2"}}},
		},
	}
}

func lookupEnv(env string) string {
	value, _ := os.LookupEnv(env)
	return value
}

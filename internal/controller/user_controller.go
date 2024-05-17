/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"time"

	certv1 "k8s.io/api/certificates/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	usermanagementv1 "github.com/ashu8778/kubernetes-user-management/tree/main/k8s-controller/api/v1"
)

// UserReconciler reconciles a User object
type UserReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=usermanagement.github.com,resources=users,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=usermanagement.github.com,resources=users/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=usermanagement.github.com,resources=users/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources="*",verbs="*"
//+kubebuilder:rbac:groups=extensions,resources="*",verbs="*"
//+kubebuilder:rbac:groups=apps,resources="*",verbs="*"
//+kubebuilder:rbac:groups=batch,resources="*",verbs="*"
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources="*",verbs="*"
//+kubebuilder:rbac:groups=certificates.k8s.io,resources="*",verbs="*"
//+kubebuilder:rbac:groups=certificates.k8s.io,resources=certificatesigningrequests/approval,verbs=get;update;patch
//+kubebuilder:rbac:groups=extensions,resources="*",verbs="*"
//+kubebuilder:rbac:groups=github.com,resources="*",verbs="*"

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the User object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.2/pkg/reconcile
func (r *UserReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	// TODO: log events
	_ = log.FromContext(ctx)
	fmt.Println("---")

	// Check if user resource exists
	user := &usermanagementv1.User{}
	err := r.Get(ctx, req.NamespacedName, user)
	if err != nil {
		// If not found, return - no need to create any dependent resource
		if errors.IsNotFound(err) {
			fmt.Printf("%v user not found. ns: %v\n", req.Name, req.Namespace)
			return ctrl.Result{}, nil
		}
		fmt.Printf("Error retriving user: %v in ns: %v\n", req.Name, req.Namespace)
		return ctrl.Result{}, err
	}

	// User resource exists

	// Create private key and csr for user
	privateKey, err := r.generateKey()
	if err != nil {
		return ctrl.Result{}, err
	}

	// Generate a csr - to be used to create k8s csr resource
	csrPem, err := r.generateCSR(privateKey, user)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Check if csr already created and approved
	if user.Status.CertificateStatus != "Approved" {
		// Check if csr resource exists, if doesen't exist & csr not approved yet - create csr in kubernetes using csrPem
		csrResource := &certv1.CertificateSigningRequest{}
		err = r.Get(ctx, req.NamespacedName, csrResource)
		if err != nil {
			if errors.IsNotFound(err) {
				err := r.createCsrK8s(ctx, user, csrPem)
				if err != nil {
					return ctrl.Result{}, err
				}
				fmt.Println(req.Name, "CSR created")
			} else {
				return ctrl.Result{}, err
			}
		}

		// Approve the csr
		err = r.autoApproveCsr(ctx, req, user)
		if err != nil {
			return ctrl.Result{}, err
		}

		csrResource = &certv1.CertificateSigningRequest{}
		err = r.Get(ctx, req.NamespacedName, csrResource)
		if err != nil {
			return ctrl.Result{}, err
		}

	} else {
		fmt.Println(req.Name, "CSR already approved")
	}

	// If user resource is found - create role for user
	err = r.createRole(ctx, user)
	if err != nil {
		return ctrl.Result{}, err
	}

	// create role binding for user
	err = r.createRoleBinding(ctx, user)
	if err != nil {
		return ctrl.Result{}, err
	}

	if user.Status.CertificateStatus == "Pending" {
		return ctrl.Result{Requeue: true, RequeueAfter: 10 * time.Second}, nil
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *UserReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&usermanagementv1.User{}).
		Complete(r)
}

// Create role using user crd resource spec permissions
func (r *UserReconciler) createRole(ctx context.Context, user *usermanagementv1.User) error {
	// Role specifications
	blockOwnerDeletion := true
	role := rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      user.Name,
			Namespace: user.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				{APIVersion: user.APIVersion, Kind: user.Kind, Name: user.Name, BlockOwnerDeletion: &blockOwnerDeletion, UID: user.UID},
			},
		},
		Rules: user.Spec.RoleRules,
	}

	// Update the role with the user rolr spec
	err := r.Update(ctx, &role)
	if err != nil {
		return err
	}
	fmt.Printf("%v role updated. ns: %v\n", role.Name, role.Namespace)
	return nil
}

// Create RoleBinding
func (r *UserReconciler) createRoleBinding(ctx context.Context, user *usermanagementv1.User) error {
	// RoleBinding specifications
	blockOwnerDeletion := true
	roleBinding := rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      user.Name,
			Namespace: user.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				{APIVersion: user.APIVersion, Kind: user.Kind, Name: user.Name, BlockOwnerDeletion: &blockOwnerDeletion, UID: user.UID},
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     user.Name,
		},
		Subjects: []rbacv1.Subject{
			{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "User",
				Name:     user.Name,
			},
		},
	}

	// Update the role binding with the user rolr spec
	err := r.Update(ctx, &roleBinding)
	if err != nil {
		return err
	}
	fmt.Printf("%v rolebinding updated. ns: %v\n", roleBinding.Name, roleBinding.Namespace)
	return nil
}

// Generates private key for the user resource
func (r *UserReconciler) generateKey() (*rsa.PrivateKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Printf("Error generating private key: %v\n", err)
		return nil, err
	}
	return privateKey, nil
}

// Get the private key, return raw certificatesigningrequest.
func (r *UserReconciler) generateCSR(privateKey *rsa.PrivateKey, user *usermanagementv1.User) ([]byte, error) {

	// Create a CSR template
	csrTemplate := x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName:   user.Name,
			Organization: []string{"my-organisation"},
		},
		SignatureAlgorithm: x509.SHA256WithRSA,
	}

	// Create a CSR from the template
	csr, err := x509.CreateCertificateRequest(rand.Reader, &csrTemplate, privateKey)
	if err != nil {
		fmt.Printf("Error creating CSR: %v\n", err)
		return nil, err
	}

	// Encode the CSR in PEM format
	csrPem := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE REQUEST",
		Bytes: csr,
	})

	return csrPem, nil
}

// Create k8s csr resource
func (r *UserReconciler) createCsrK8s(ctx context.Context, user *usermanagementv1.User, csrPem []byte) error {
	// one month expiry
	expirationSeconds := int32(2592000)
	// Kubernetes csr spec
	blockOwnerDeletion := true
	csr := &certv1.CertificateSigningRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      user.Name,
			Namespace: user.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				{APIVersion: user.APIVersion, Kind: user.Kind, Name: user.Name, BlockOwnerDeletion: &blockOwnerDeletion, UID: user.UID},
			},
		},
		Spec: certv1.CertificateSigningRequestSpec{
			Request:    csrPem,
			SignerName: "kubernetes.io/kube-apiserver-client",
			Usages: []certv1.KeyUsage{
				"digital signature",
				"key encipherment",
				"client auth",
			},
			ExpirationSeconds: &expirationSeconds,
		},
	}
	err := r.Create(ctx, csr)
	return err
}

// Auto approves k8s csr created by user crd
func (r *UserReconciler) autoApproveCsr(ctx context.Context, req ctrl.Request, user *usermanagementv1.User) error {

	// Get the clientset to approve the certificate
	clientset, err := getK8sClientset()
	if err != nil {
		return err
	}

	// Get fresh obj - Check if csr resource exists
	csrResource := &certv1.CertificateSigningRequest{}
	err = r.Get(ctx, req.NamespacedName, csrResource)
	if err != nil {
		return err
	}

	// Update the certificateStatus in user crd to pending if csr with same name already exists
	if csrResource.ObjectMeta.OwnerReferences[0].UID != user.ObjectMeta.UID {
		fmt.Println("CSR/Certificate creation pending. Already existing csr with same name.")
		user.Status.CertificateStatus = "Pending"
		err = r.Status().Update(ctx, user)
		if err != nil {
			return err
		}
		return nil
	}

	// Approve the CSR if not approved or denied yet
	if len(csrResource.Status.Conditions) != 0 {

		if csrResource.Status.Conditions[len(csrResource.Status.Conditions)-1].Type == certv1.CertificateApproved || csrResource.Status.Conditions[len(csrResource.Status.Conditions)-1].Type == certv1.CertificateDenied {

			fmt.Println(req.Name, "CSR already approved or denied")
			return nil
		}
	} else {

		csrResource.Status.Conditions = append(csrResource.Status.Conditions, certv1.CertificateSigningRequestCondition{
			Type:           certv1.CertificateApproved,
			Status:         corev1.ConditionTrue,
			Reason:         "ControllerApproved",
			Message:        "This CSR was approved by the controller.",
			LastUpdateTime: metav1.NewTime(time.Now()),
		})

		// Update the CSR
		_, err = clientset.CertificatesV1().CertificateSigningRequests().UpdateApproval(ctx, req.Name, csrResource, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
		// Update the certificateStatus to approved
		user.Status.CertificateStatus = "Approved"
		err = r.Status().Update(ctx, user)
		if err != nil {
			return err
		}
		fmt.Println(req.Name, "CSR is approved.")
	}

	return nil
}

// Generates k8s clientset
func getK8sClientset() (*kubernetes.Clientset, error) {

	// Get the path to the kubeconfig file. If it's not provided, assume it's in the default location.
	kubeconfig := filepath.Join(
		os.Getenv("HOME"), ".kube", "config",
	)

	// Check if the controller is running inside cluster
	config, err := rest.InClusterConfig()
	if err != nil {
		// Load kubeconfig file - controller is not running inside cluster
		fmt.Println("Controller not running inside cluster. Use local config")
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}
	}

	// Create the Kubernetes clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

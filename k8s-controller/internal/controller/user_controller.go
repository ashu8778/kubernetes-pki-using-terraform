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
	"fmt"

	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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

	// Check if resource exists
	user := &usermanagementv1.User{}
	err := r.Get(ctx, req.NamespacedName, user)
	if err != nil {
		// If not found, return - no need to create any resource
		if errors.IsNotFound(err) {
			fmt.Printf("Resouce user: %v in ns: %v not found\n", req.Name, req.Namespace)
			return ctrl.Result{}, nil
		}
		fmt.Printf("Error retriving user: %v in ns: %v\n", req.Name, req.Namespace)
		return ctrl.Result{}, err
	}

	// If found - create role for user

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
	err = r.Update(ctx, &role)
	if err != nil {
		return ctrl.Result{}, err
	}

	fmt.Printf("Resouce role: %v in ns: %v updated\n", role.Name, role.Namespace)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *UserReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&usermanagementv1.User{}).
		Complete(r)
}

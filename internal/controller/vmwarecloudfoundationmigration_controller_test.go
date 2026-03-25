/*
Copyright 2026.

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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	configv1 "github.com/openshift/api/config/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	migrationv1alpha1 "github.com/openshift/vcf-migration-operator/api/v1alpha1"
)

var _ = Describe("VmwareCloudFoundationMigration Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default", // TODO(user):Modify as needed
		}
		vmwarecloudfoundationmigration := &migrationv1alpha1.VmwareCloudFoundationMigration{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind VmwareCloudFoundationMigration")
			err := k8sClient.Get(ctx, typeNamespacedName, vmwarecloudfoundationmigration)
			if err != nil && errors.IsNotFound(err) {
				resource := &migrationv1alpha1.VmwareCloudFoundationMigration{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: migrationv1alpha1.VmwareCloudFoundationMigrationSpec{
						State: migrationv1alpha1.MigrationStatePending,
						TargetVCenterCredentialsSecret: migrationv1alpha1.SecretReference{
							Name:      "target-vcenter-creds",
							Namespace: "default",
						},
						FailureDomains: []configv1.VSpherePlatformFailureDomainSpec{
							{
								Name:   "target-fd-1",
								Region: "target-region",
								Zone:   "target-zone-1",
								Server: "vcenter-target.example.com",
								Topology: configv1.VSpherePlatformTopology{
									Datacenter:     "TargetDC",
									ComputeCluster: "/TargetDC/host/TargetCluster",
									Datastore:      "/TargetDC/datastore/TargetDatastore",
									Networks:       []string{"VM Network"},
									ResourcePool:   "/TargetDC/host/TargetCluster/Resources",
									Template:       "/TargetDC/vm/rhcos-template",
									Folder:         "/TargetDC/vm/my-cluster-infra-id",
								},
							},
						},
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			// TODO(user): Cleanup logic after each test, like removing the resource instance.
			resource := &migrationv1alpha1.VmwareCloudFoundationMigration{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance VmwareCloudFoundationMigration")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})
		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &VmwareCloudFoundationMigrationReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			// TODO(user): Add more specific assertions depending on your controller's reconciliation logic.
			// Example: If you expect a certain status condition after reconciliation, verify it here.
		})
	})
})

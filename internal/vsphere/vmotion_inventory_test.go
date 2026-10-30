package vsphere

import (
	"context"
	"testing"

	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/simulator"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

func TestListVMotionVMInventoryUsesBIOSUUIDAndInventoryPath(t *testing.T) {
	simulator.Test(func(ctx context.Context, client *vim25.Client) {
		session := newTestSession(ctx, t, client)
		vms, err := session.Finder.VirtualMachineList(ctx, "*")
		if err != nil {
			t.Fatalf("listing simulator VMs: %v", err)
		}
		if len(vms) == 0 {
			t.Fatal("simulator has no VMs")
		}
		var vm mo.VirtualMachine
		if err := vms[0].Properties(ctx, vms[0].Reference(), []string{"config.uuid", "resourcePool"}, &vm); err != nil {
			t.Fatalf("getting simulator VM BIOS UUID: %v", err)
		}
		if vm.Config == nil || vm.Config.Uuid == "" || vm.ResourcePool == nil {
			t.Fatal("simulator VM has no BIOS UUID or resource pool")
		}

		wantPath, err := find.InventoryPath(ctx, client, vms[0].Reference())
		if err != nil {
			t.Fatalf("getting simulator VM inventory path: %v", err)
		}
		resourcePool := object.NewResourcePool(client, *vm.ResourcePool)
		owner, err := resourcePool.Owner(ctx)
		if err != nil {
			t.Fatalf("getting simulator resource-pool owner: %v", err)
		}
		ownerPath, err := find.InventoryPath(ctx, client, owner.Reference())
		if err != nil {
			t.Fatalf("getting simulator compute-resource path: %v", err)
		}
		inventory, err := ListVMotionVMInventory(ctx, session, "source.example.com", []VMotionSourceFailureDomain{{Name: "source-a", ComputeResourcePath: ownerPath}}, []string{vm.Config.Uuid})
		if err != nil {
			t.Fatalf("ListVMotionVMInventory: %v", err)
		}
		if len(inventory.VMs) != 1 {
			t.Fatalf("ListVMotionVMInventory returned %d VMs, want exactly one", len(inventory.VMs))
		}
		got := inventory.VMs[0]
		if got.BIOSUUID != vm.Config.Uuid || got.InventoryPath != wantPath || got.NetworkAdapterCount != 1 {
			t.Fatalf("ListVMotionVMInventory result = %+v, want UUID %q, path %q, and one NIC", got, vm.Config.Uuid, wantPath)
		}
		wantOwner := VMotionManagedObjectReference{VCenter: "source.example.com", Datacenter: session.Datacenter, Type: owner.Reference().Type, Value: owner.Reference().Value}
		if got.ResourcePoolRef.Value != vm.ResourcePool.Value || got.ResourcePoolOwner != wantOwner ||
			inventory.SourceFailureDomainReferences["source-a"] != wantOwner {
			t.Fatalf("resource-pool references = pool %+v, owner %+v, FD refs %+v; want pool %s, owner %+v", got.ResourcePoolRef, got.ResourcePoolOwner, inventory.SourceFailureDomainReferences, vm.ResourcePool.Value, wantOwner)
		}

		inventory, err = ListVMotionVMInventory(ctx, session, "source.example.com", []VMotionSourceFailureDomain{{Name: "source-a", ComputeResourcePath: ownerPath}}, []string{"11111111-1111-4111-8111-111111111111"})
		if err != nil {
			t.Fatalf("ListVMotionVMInventory for absent UUID: %v", err)
		}
		if len(inventory.VMs) != 0 {
			t.Fatalf("ListVMotionVMInventory for absent UUID returned %d VMs, want 0", len(inventory.VMs))
		}
		duplicateInventory, err := ListVMotionVMInventory(ctx, session, "source.example.com", []VMotionSourceFailureDomain{{Name: "source-a", ComputeResourcePath: ownerPath}, {Name: "source-b", ComputeResourcePath: ownerPath}}, []string{vm.Config.Uuid})
		if err != nil {
			t.Fatalf("listing duplicate refs for controller preflight: %v", err)
		}
		if duplicateInventory.SourceFailureDomainReferences["source-a"] != duplicateInventory.SourceFailureDomainReferences["source-b"] {
			t.Fatalf("duplicate source paths resolved to different references: %+v", duplicateInventory.SourceFailureDomainReferences)
		}
		if _, err := ListVMotionVMInventory(ctx, session, "source.example.com", []VMotionSourceFailureDomain{{Name: "source-other", ComputeResourcePath: "/OtherDC/host/Cluster"}}, []string{vm.Config.Uuid}); err == nil {
			t.Fatal("compute resource outside the configured datacenter must be rejected")
		}
	})
}

func TestListVMotionVMInventoryUsesResolvedNestedDatacenter(t *testing.T) {
	simulator.Test(func(ctx context.Context, client *vim25.Client) {
		session := newTestSession(ctx, t, client)
		dc, err := session.Finder.Datacenter(ctx, session.Datacenter)
		if err != nil {
			t.Fatal(err)
		}
		root := object.NewFolder(client, client.ServiceContent.RootFolder)
		tenant, err := root.CreateFolder(ctx, "tenant")
		if err != nil {
			t.Fatal(err)
		}
		move, err := tenant.MoveInto(ctx, []types.ManagedObjectReference{dc.Reference()})
		if err != nil {
			t.Fatal(err)
		}
		if err := move.Wait(ctx); err != nil {
			t.Fatal(err)
		}

		// Resolve again after the move so cached Finder paths cannot hide a scope bug.
		session.Finder = find.NewFinder(client, true)
		dc, err = session.Finder.Datacenter(ctx, session.Datacenter)
		if err != nil {
			t.Fatal(err)
		}
		session.Finder.SetDatacenter(dc)
		vms, err := session.Finder.VirtualMachineList(ctx, "*")
		if err != nil || len(vms) == 0 {
			t.Fatalf("finding nested simulator VMs: %v", err)
		}
		var vm mo.VirtualMachine
		if err := vms[0].Properties(ctx, vms[0].Reference(), []string{"config.uuid", "resourcePool"}, &vm); err != nil {
			t.Fatal(err)
		}
		if vm.Config == nil || vm.ResourcePool == nil {
			t.Fatal("VM identity or resource pool missing")
		}
		owner, err := object.NewResourcePool(client, *vm.ResourcePool).Owner(ctx)
		if err != nil {
			t.Fatal(err)
		}
		ownerPath, err := find.InventoryPath(ctx, client, owner.Reference())
		if err != nil {
			t.Fatal(err)
		}
		wantVMPath, err := find.InventoryPath(ctx, client, vms[0].Reference())
		if err != nil {
			t.Fatal(err)
		}
		inventory, err := ListVMotionVMInventory(ctx, session, "source.example.com", []VMotionSourceFailureDomain{{Name: "nested-fd", ComputeResourcePath: ownerPath}}, []string{vm.Config.Uuid})
		if err != nil {
			t.Fatalf("ListVMotionVMInventory for %q resolved as %q: %v", session.Datacenter, dc.InventoryPath, err)
		}
		if len(inventory.VMs) != 1 || inventory.VMs[0].BIOSUUID != vm.Config.Uuid || inventory.VMs[0].InventoryPath != wantVMPath {
			t.Fatalf("nested VM discovery lost identity/path: %+v", inventory)
		}
		wantOwner := VMotionManagedObjectReference{VCenter: "source.example.com", Datacenter: session.Datacenter, Type: owner.Reference().Type, Value: owner.Reference().Value}
		if inventory.SourceFailureDomainReferences["nested-fd"] != wantOwner || inventory.VMs[0].ResourcePoolOwner != wantOwner {
			t.Fatalf("nested FD/pool owner mismatch: %+v", inventory)
		}

		otherDC, err := root.CreateDatacenter(ctx, "OtherDC")
		if err != nil {
			t.Fatal(err)
		}
		folders, err := otherDC.Folders(ctx)
		if err != nil {
			t.Fatal(err)
		}
		otherCompute, err := folders.HostFolder.CreateCluster(ctx, "OtherCluster", types.ClusterConfigSpecEx{})
		if err != nil {
			t.Fatal(err)
		}
		otherPath, err := find.InventoryPath(ctx, client, otherCompute.Reference())
		if err != nil {
			t.Fatal(err)
		}
		if _, err := ListVMotionVMInventory(ctx, session, "source.example.com", []VMotionSourceFailureDomain{{Name: "other-fd", ComputeResourcePath: otherPath}}, []string{vm.Config.Uuid}); err == nil {
			t.Fatal("accepted a real compute resource from a different datacenter")
		}
	})
}

package vsphere

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/session"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"
)

// VMotionVM is a read-only vSphere VM identity and observed placement.
type VMotionVM struct {
	// BIOSUUID is VirtualMachine.config.uuid, not config.instanceUuid.
	BIOSUUID string
	// InventoryPath is the VM's absolute path within the source datacenter.
	InventoryPath string
	// Host is the observed ESXi host path when returned by an inspection call.
	Host string
	// PowerState is the observed vSphere power state when returned by an inspection call.
	PowerState string
	// ResourcePoolRef is the VM's resourcePool managed-object reference.
	ResourcePoolRef VMotionManagedObjectReference
	// ResourcePoolOwner is the resource pool's owning compute-resource reference.
	ResourcePoolOwner VMotionManagedObjectReference
	// NetworkAdapterCount is the number of VirtualEthernetCard devices.
	NetworkAdapterCount int
}

// ListVMotionVMInventory resolves each source failure-domain compute-resource
// path and finds VMs in the same datacenter by BIOS UUID. Duplicate references
// and VM matches are retained so callers can fail closed on ambiguity.
func ListVMotionVMInventory(ctx context.Context, s *Session, vcenter string, sourceFailureDomains []VMotionSourceFailureDomain, biosUUIDs []string) (VMotionInventory, error) {
	result := VMotionInventory{SourceFailureDomainReferences: make(map[string]VMotionManagedObjectReference, len(sourceFailureDomains))}
	if s == nil || s.Client == nil || s.Finder == nil {
		return VMotionInventory{}, fmt.Errorf("listing VMotion VM inventory: vSphere session is not configured")
	}
	if vcenter == "" {
		return VMotionInventory{}, fmt.Errorf("listing VMotion VM inventory: source vCenter must not be empty")
	}

	datacenter, err := s.Finder.Datacenter(ctx, s.Datacenter)
	if err != nil {
		return VMotionInventory{}, fmt.Errorf("finding source datacenter %q: %w", s.Datacenter, err)
	}
	scopePrefix := strings.TrimRight(datacenter.InventoryPath, "/") + "/host/"
	for _, failureDomain := range sourceFailureDomains {
		if failureDomain.Name == "" || failureDomain.ComputeResourcePath == "" {
			return VMotionInventory{}, fmt.Errorf("resolving source failure domain %q: name and compute-resource path must not be empty", failureDomain.Name)
		}
		resource, err := s.Finder.ComputeResource(ctx, failureDomain.ComputeResourcePath)
		if err != nil {
			return VMotionInventory{}, fmt.Errorf("resolving compute resource %q for source failure domain %q: %w", failureDomain.ComputeResourcePath, failureDomain.Name, err)
		}
		if resource.InventoryPath != failureDomain.ComputeResourcePath || !strings.HasPrefix(resource.InventoryPath, scopePrefix) {
			return VMotionInventory{}, fmt.Errorf("resolving source failure domain %q: compute resource %q is outside datacenter %q or does not match its configured path", failureDomain.Name, resource.InventoryPath, s.Datacenter)
		}
		ref := resource.Reference()
		result.SourceFailureDomainReferences[failureDomain.Name] = VMotionManagedObjectReference{
			VCenter: vcenter, Datacenter: s.Datacenter, Type: ref.Type, Value: ref.Value,
		}
	}
	if len(biosUUIDs) == 0 {
		return result, nil
	}
	wanted := make(map[string]struct{}, len(biosUUIDs))
	for _, uuid := range biosUUIDs {
		wanted[strings.ToLower(uuid)] = struct{}{}
	}
	containerView, err := view.NewManager(s.Client.Client).CreateContainerView(ctx, datacenter.Reference(), []string{"VirtualMachine"}, true)
	if err != nil {
		return VMotionInventory{}, fmt.Errorf("creating source datacenter VM inventory view: %w", err)
	}
	defer func() { _ = containerView.Destroy(ctx) }()

	var inventory []mo.VirtualMachine
	if err := containerView.Retrieve(ctx, []string{"VirtualMachine"}, []string{"config.uuid", "config.hardware.device", "resourcePool"}, &inventory); err != nil {
		return VMotionInventory{}, fmt.Errorf("listing source datacenter VM identities: %w", err)
	}

	result.VMs = make([]VMotionVM, 0, len(biosUUIDs))
	for i := range inventory {
		vm := &inventory[i]
		if vm.Config == nil || vm.Config.Uuid == "" {
			continue
		}
		if _, ok := wanted[strings.ToLower(vm.Config.Uuid)]; !ok {
			continue
		}
		inventoryPath, err := find.InventoryPath(ctx, s.Client.Client, vm.Reference())
		if err != nil {
			return VMotionInventory{}, fmt.Errorf("getting inventory path for VM BIOS UUID %q: %w", vm.Config.Uuid, err)
		}
		networkAdapterCount := 0
		for _, device := range vm.Config.Hardware.Device {
			if _, ok := device.(types.BaseVirtualEthernetCard); ok {
				networkAdapterCount++
			}
		}
		vmInventory := VMotionVM{
			BIOSUUID: vm.Config.Uuid, InventoryPath: inventoryPath, NetworkAdapterCount: networkAdapterCount,
		}
		if vm.ResourcePool != nil {
			vmInventory.ResourcePoolRef = VMotionManagedObjectReference{
				VCenter: vcenter, Datacenter: s.Datacenter, Type: vm.ResourcePool.Type, Value: vm.ResourcePool.Value,
			}
			owner, err := object.NewResourcePool(s.Client.Client, *vm.ResourcePool).Owner(ctx)
			if err != nil {
				return VMotionInventory{}, fmt.Errorf("getting resource-pool owner for VM BIOS UUID %q: %w", vm.Config.Uuid, err)
			}
			ownerRef := owner.Reference()
			if ownerRef.Type != "" && ownerRef.Value != "" {
				ownerPath, err := find.InventoryPath(ctx, s.Client.Client, ownerRef)
				if err != nil {
					return VMotionInventory{}, fmt.Errorf("getting resource-pool owner path for VM BIOS UUID %q: %w", vm.Config.Uuid, err)
				}
				if !strings.HasPrefix(ownerPath, scopePrefix) {
					return VMotionInventory{}, fmt.Errorf("getting resource-pool owner for VM BIOS UUID %q: owner %q is outside datacenter %q", vm.Config.Uuid, ownerPath, s.Datacenter)
				}
				vmInventory.ResourcePoolOwner = VMotionManagedObjectReference{
					VCenter: vcenter, Datacenter: s.Datacenter, Type: ownerRef.Type, Value: ownerRef.Value,
				}
			}
		}
		result.VMs = append(result.VMs, vmInventory)
	}
	sort.Slice(result.VMs, func(i, j int) bool {
		if result.VMs[i].BIOSUUID == result.VMs[j].BIOSUUID {
			return result.VMs[i].InventoryPath < result.VMs[j].InventoryPath
		}
		return strings.ToLower(result.VMs[i].BIOSUUID) < strings.ToLower(result.VMs[j].BIOSUUID)
	})
	return result, nil
}

// ListDatacenters connects to vCenter and returns the names of all datacenters.
// It does not require a datacenter in Params; Server, Username, and Password are used.
func ListDatacenters(ctx context.Context, p Params) ([]string, error) {
	if p.Server == "" || p.Username == "" || p.Password == "" {
		return nil, fmt.Errorf("server, username, and password are required")
	}

	host := SanitizeServer(p.Server)
	u := &url.URL{
		Scheme: "https",
		Host:   host,
		Path:   vim25.Path,
	}
	u.User = url.UserPassword(p.Username, p.Password)

	soapClient := soap.NewClient(u, p.Insecure)
	vimClient, err := vim25.NewClient(ctx, soapClient)
	if err != nil {
		return nil, fmt.Errorf("creating vim25 client for %s: %w", host, err)
	}

	client := &govmomi.Client{
		Client:         vimClient,
		SessionManager: session.NewManager(vimClient),
	}
	if err := client.Login(ctx, u.User); err != nil {
		return nil, fmt.Errorf("logging in: %w", err)
	}
	defer func() { _ = client.Logout(ctx) }()

	finder := find.NewFinder(vimClient, true)
	dcs, err := finder.DatacenterList(ctx, "*")
	if err != nil {
		return nil, fmt.Errorf("listing datacenters: %w", err)
	}

	names := make([]string, 0, len(dcs))
	for _, dc := range dcs {
		names = append(names, dc.Name())
	}
	return names, nil
}

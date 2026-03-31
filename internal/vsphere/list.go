package vsphere

import (
	"context"
	"fmt"
	"net/url"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/session"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/soap"
)

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
	defer client.Logout(ctx)

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

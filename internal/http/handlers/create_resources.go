package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/icloudeng/platform-installer/internal/database/entities"
	"github.com/icloudeng/platform-installer/internal/http/validators"
	"github.com/icloudeng/platform-installer/internal/resources/db"
	"github.com/icloudeng/platform-installer/internal/resources/jobs"
	"github.com/icloudeng/platform-installer/internal/resources/proxmox"
	"github.com/icloudeng/platform-installer/internal/resources/terraform"
	"github.com/icloudeng/platform-installer/internal/resources/utilities"
	"github.com/icloudeng/platform-installer/internal/structs"
)

func createResourceJob(ctx *gin.Context, json *resourcesBody) *entities.Job {
	// If domain key doesn't exist in metadata platform
	// then auto fill with the passed domain resource
	metadata := json.Platform.Metadata
	domain := fmt.Sprintf("%s.%s", json.Domain.Subdomain, json.Domain.Zone)

	_, domain_exists := metadata["domain"]
	if !domain_exists {
		json.Platform.Metadata["domain"] = domain
	}

	// Chech if platform the password corresponse to an existing platform folder
	if !validators.ValidatePlatformMetadata(ctx, *json.Platform) {
		return nil
	}

	// Skip apply when both resources exist
	var should_skip_apply = false
	_vm := terraform.Resources.GetProxmoxVmQemuResource(json.Ref)
	_domain := terraform.Resources.GetOvhDomainZoneResource(json.Ref)

	if _vm != nil && _domain != nil {
		should_skip_apply = true
	}

	// Failure when resources exists on POST request
	if ctx.Request.Method == "POST" {
		if _vm != nil || _domain != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": ResourceExistsError,
				"resource": map[string]interface{}{
					"vm":     _vm,
					"domain": _domain,
				},
			})

			return nil
		}
	}

	// Check if VM Id doesn't exist
	// if json.Vm.Vmid != 0 {
	// 	if exists := proxmox.VmQemuIDExists(json.Vm.Vmid); exists {
	// 		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
	// 			"error": "VM ID already exists !",
	// 		})
	// 		return nil
	// 	}
	// }

	// If Target Node is set to auto,
	// then selected automatic node based on resourse Availability
	target_node := json.Vm.TargetNode
	if target_node == "auto" {
		nodeStatus, err := proxmox.SelectNodeWithMostResources()
		if err != nil || nodeStatus == nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "No enough proxmox resources",
			})
			return nil
		}

		json.Vm.TargetNode = nodeStatus.Node
	}

	mxDomain := autoComposeMxDomain(domain, json)
	if mxDomain != nil {
		json.MxDomainValue = mxDomain
	}

	if json.Platform != nil && mxDomain != nil {
		json.Platform.Metadata["mx_domain"] = utilities.Helpers.ConcatenateSubdomain(
			mxDomain.Subdomain,
			mxDomain.Zone,
		)
	}

	json.Vm.Description = fmt.Sprintf("https://%s", domain)

	task := jobs.ResourcesJob{
		Ref:           json.Ref,
		PostBody:      json,
		ResourceState: true,
		Description:   "Resources creation",
		Handler:       ctx.Request.URL.String(),
		Method:        ctx.Request.Method,
		Task: func(ctx context.Context, job *entities.Job) error {
			if should_skip_apply {
				return nil
			}

			// Reselect Targe node
			if target_node == "auto" {
				if nodeStatus, err := proxmox.SelectNodeWithMostResources(); nodeStatus != nil && err == nil {
					json.Vm.TargetNode = nodeStatus.Node
					db.Jobs.JobUpdatePostBody(job, json)
				}
			}

			// Reset unmutable vm fields
			structs.ResetUnmutableProxmoxVmQemu(&structs.ResetProxmoxVmQemuFields{
				Vm:       json.Vm,
				Platform: *json.Platform,
				Ref:      json.Ref,
				JobID:    job.ID,
			})
			// Create or update resources
			terraform.Resources.WriteOvhDomainZoneResource(json.Ref, json.Domain)
			// MX Domain
			if mxDomain != nil {
				terraform.Resources.WriteOvhDomainZoneResource(
					fmt.Sprintf("mx-%s", json.Ref),
					mxDomain,
				)
			}
			terraform.Resources.WriteProxmoxVmQemuResource(json.Ref, json.Vm)

			// Terraform Apply changes
			return terraform.Exec.Apply(true)
		},
	}

	return jobs.ResourcesJobTask(task)

}

func autoComposeMxDomain(resourceDomain string, json *resourcesBody) *structs.DomainZoneRecord {
	if json.MxDomain != nil && len(*json.MxDomain) > 1 {
		mx_value := *json.MxDomain
		if mx_value == "auto" || mx_value == resourceDomain {
			subdomain, rootDomain := utilities.Helpers.ExtractSubdomainAndRootDomain(resourceDomain)
			return &structs.DomainZoneRecord{
				Zone:      rootDomain,
				Subdomain: utilities.RemoveFirstSegment(subdomain),
				Fieldtype: "MX",
				Ttl:       3600,
				Target:    fmt.Sprintf("1 %s.", resourceDomain),
			}
		} else {
			subdomain, rootDomain := utilities.Helpers.ExtractSubdomainAndRootDomain(mx_value)
			return &structs.DomainZoneRecord{
				Zone:      rootDomain,
				Subdomain: subdomain,
				Fieldtype: "MX",
				Ttl:       3600,
				Target:    fmt.Sprintf("1 %s.", resourceDomain),
			}
		}
	}

	return nil
}

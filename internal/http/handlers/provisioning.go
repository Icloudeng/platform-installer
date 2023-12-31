package handlers

import (
	"context"
	"net/http"

	"github.com/icloudeng/platform-installer/internal/database/entities"
	"github.com/icloudeng/platform-installer/internal/http/validators"
	"github.com/icloudeng/platform-installer/internal/resources/jobs"
	"github.com/icloudeng/platform-installer/internal/resources/provisioning"
	"github.com/icloudeng/platform-installer/internal/structs"

	"github.com/gin-gonic/gin"
)

type (
	provisioningHandler struct{}
)

var Provisioning provisioningHandler

func (provisioningHandler) CreatePlatformProvisioning(c *gin.Context) {
	body := &structs.Provisioning{
		Platform: &structs.Platform{
			Metadata: map[string]interface{}{},
		},
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if platform name corresponse to an existing platform folder
	if !validators.ValidatePlatformMetadata(c, *body.Platform) {
		return
	}

	// Validate ref and bind resource state platform values
	if !validators.ValidatePlatformProvisionAndBindResourceState(body) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "Cannot find resource linked to passed reference name!",
		})
		return
	}

	ref := body.MachineIp
	if len(body.Ref) > 0 {
		ref = body.Ref
	}

	task := jobs.ResourcesJob{
		Ref:           ref,
		PostBody:      body,
		Description:   "Platform Provisioning",
		ResourceState: len(body.Ref) > 0,
		Handler:       c.Request.URL.String(),
		Method:        c.Request.Method,
		Task: func(ctx context.Context, job *entities.Job) error {
			provisioning.CreatePlatformProvisioning(*body, job.ID)

			return nil
		},
	}

	job := jobs.ResourcesJobTask(task)

	c.JSON(http.StatusOK, gin.H{"data": body, "job": job})
}

func (provisioningHandler) CreateConfigurationProvisioning(c *gin.Context) {
	body := &structs.Provisioning{
		Platform: &structs.Platform{
			Metadata: map[string]interface{}{},
		},
	}

	if err := c.ShouldBindJSON(body); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Chech if platform the password corresponse to an existing platform folder
	if !validators.ValidateConfigurationMetadata(c, *body.Platform) {
		return
	}

	// Validate ref and bind resource state platform values
	if !validators.ValidatePlatformProvisionAndBindResourceState(body) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "Cannot find resource linked to passed reference name!",
		})
		return
	}

	task := jobs.ResourcesJob{
		Ref:           body.MachineIp,
		PostBody:      body,
		Description:   "Platform Configuration Provisioning",
		ResourceState: false,
		Handler:       c.Request.URL.String(),
		Method:        c.Request.Method,
		Task: func(ctx context.Context, job *entities.Job) error {
			provisioning.CreateConfigurationProvisioning(*body, job.ID)

			return nil
		},
	}

	job := jobs.ResourcesJobTask(task)

	c.JSON(http.StatusOK, gin.H{"data": body, "job": job})
}

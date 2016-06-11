package compute

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// OSImages represents a page of OSImage results.
type OSImages struct {
	// The current page of network domains.
	Images []OSImage `json:"osImage"`

	// The current page number.
	PageNumber int `json:"pageNumber"`

	// The number of OS images in the current page of results.
	PageCount int `json:"pageCount"`

	// The total number of OS images that match the requested filter criteria (if any).
	TotalCount int `json:"totalCount"`

	// The maximum number of OS images per page.
	PageSize int `json:"pageSize"`
}

// OSImage represents a virtual machine image.
type OSImage struct {
	ID              string               `json:"id"`
	Name            string               `json:"name"`
	Description     string               `json:"description"`
	DataCenterID    string               `json:"datacenterId"`
	OperatingSystem OperatingSystem      `json:"operatingSystem"`
	CPU             VirtualMachineCPU    `json:"cpu"`
	MemoryGB        int                  `json:"memoryGb"`
	Disks           []VirtualMachineDisk `json:"disk"`
	CreateTime      string               `json:"createTime"`
	OSImageKey      string               `json:"osImageKey"`
}

// OperatingSystem a well-known operating system for virtual machines.
type OperatingSystem struct {
	// The operating system Id.
	ID string `json:"id"`

	// The operating system type.
	Family string `json:"family"`

	// The operating system display-name.
	DisplayName string `json:"displayName"`
}

// VirtualMachineCPU represents the CPU configuration for a virtual machine.
type VirtualMachineCPU struct {
	Count          int    `json:"count"`
	Speed          string `json:"speed"`
	CoresPerSocket int    `json:"coresPerSocket"`
}

// VirtualMachineDisk represents the disk configuration for a virtual machine.
type VirtualMachineDisk struct {
	ID         string `json:"id"`
	SCSIUnitID int    `json:"scsiId"`
	SizeGB     int    `json:"sizeGb"`
	Speed      string `json:"speed"`
}

// FindOSImage finds an OS image by name in a given data centre.
func (client *Client) FindOSImage(name string, dataCenterID string) (image *OSImage, err error) {
	organizationID, err := client.getOrganizationID()
	if err != nil {
		return nil, err
	}

	requestURI := fmt.Sprintf("%s/image/osImage?name=%s&datacenterId=%s", organizationID, url.QueryEscape(name), url.QueryEscape(dataCenterID))
	request, err := client.newRequestV22(requestURI, http.MethodGet, nil)
	if err != nil {
		return nil, err
	}

	responseBody, statusCode, err := client.executeRequest(request)
	if err != nil {
		return nil, err
	}

	if statusCode != http.StatusOK {
		var apiResponse *APIResponse

		apiResponse, err = readAPIResponseAsJSON(responseBody, statusCode)
		if err != nil {
			return nil, err
		}

		return nil, fmt.Errorf("Request to list VLANs failed with status code %d (%s): %s", statusCode, apiResponse.ResponseCode, apiResponse.Message)
	}

	images := &OSImages{}
	err = json.Unmarshal(responseBody, images)
	if err != nil {
		return nil, err
	}

	if images.PageCount == 0 {
		return nil, nil
	}

	if images.PageCount != 1 {
		return nil, fmt.Errorf("Found multiple images (%d) matching '%s' in data centre '%s'.", images.TotalCount, name, dataCenterID)
	}

	return &images.Images[0], err
}

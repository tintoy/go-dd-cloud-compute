package compute

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// OSImage represents a DD-provided virtual machine image.
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

// ToEntityReference creates an EntityReference representing the OSImage.
func (image *OSImage) ToEntityReference() EntityReference {
	return EntityReference{
		ID:   image.ID,
		Name: image.Name,
	}
}

var _ NamedEntity = &OSImage{}

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

// GetOSImage retrieves a specific OS image by Id.
func (client *Client) GetOSImage(id string) (image *OSImage, err error) {
	organizationID, err := client.getOrganizationID()
	if err != nil {
		return nil, err
	}

	requestURI := fmt.Sprintf("%s/image/osImage/%s", organizationID, id)
	request, err := client.newRequestV22(requestURI, http.MethodGet, nil)
	if err != nil {
		return nil, err
	}
	responseBody, statusCode, err := client.executeRequest(request)
	if err != nil {
		return nil, err
	}

	if statusCode != http.StatusOK {
		var apiResponse *APIResponseV2

		apiResponse, err = readAPIResponseAsJSON(responseBody, statusCode)
		if err != nil {
			return nil, err
		}

		if apiResponse.ResponseCode == ResponseCodeResourceNotFound {
			return nil, nil // Not an error, but was not found.
		}

		return nil, apiResponse.ToError("Request to retrieve OS image '%s' failed with status code %d (%s): %s", id, statusCode, apiResponse.ResponseCode, apiResponse.Message)
	}

	image = &OSImage{}
	err = json.Unmarshal(responseBody, image)
	if err != nil {
		return nil, err
	}

	return image, nil
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
		var apiResponse *APIResponseV2

		apiResponse, err = readAPIResponseAsJSON(responseBody, statusCode)
		if err != nil {
			return nil, err
		}

		return nil, fmt.Errorf("Request to find OS image '%s' in data centre '%s' failed with status code %d (%s): %s", name, dataCenterID, statusCode, apiResponse.ResponseCode, apiResponse.Message)
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

// ListOSImagesInDatacenter lists all OS images in a given data centre.
func (client *Client) ListOSImagesInDatacenter(dataCenterID string, paging *Paging) (images *OSImages, err error) {
	organizationID, err := client.getOrganizationID()
	if err != nil {
		return nil, err
	}

	requestURI := fmt.Sprintf("%s/image/osImage?datacenterId=%s&%s",
		organizationID,
		url.QueryEscape(dataCenterID),
		paging.EnsurePaging().toQueryParameters(),
	)
	request, err := client.newRequestV22(requestURI, http.MethodGet, nil)
	if err != nil {
		return nil, err
	}

	responseBody, statusCode, err := client.executeRequest(request)
	if err != nil {
		return nil, err
	}

	if statusCode != http.StatusOK {
		var apiResponse *APIResponseV2

		apiResponse, err = readAPIResponseAsJSON(responseBody, statusCode)
		if err != nil {
			return nil, err
		}

		return nil, fmt.Errorf("Request to list OS images in data centre '%s' failed with status code %d (%s): %s", dataCenterID, statusCode, apiResponse.ResponseCode, apiResponse.Message)
	}

	images = &OSImages{}
	err = json.Unmarshal(responseBody, images)

	return
}

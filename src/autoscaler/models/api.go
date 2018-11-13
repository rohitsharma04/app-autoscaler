package models

type InstanceCreationRequestBody struct {
	OrgGuid   string `json:"organization_guid"`
	SpaceGuid string `json:"space_guid"`
}

type BindingRequestBody struct {
	AppId  string `json:"app_guid"`
	Policy string `json:"parameters"`
}

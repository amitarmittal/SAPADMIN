package responsedto

type GetLicenseStatusRespDto struct {
	Status           string `json:"status"`
	ErrorDescription string `json:"errorDescription"`
	LicenseStatus    bool   `json:"licenseStatus"`
}

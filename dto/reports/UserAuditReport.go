package reports

type UserAuditReport struct {
	Time      int64       `json:"time"`
	UserName  string      `json:"user_name"`
	Operation string      `json:"operation"`
	Payload   interface{} `json:"payload"`
	IP        string      `json:"ip"`
}

type UserAuditReportRespDto struct {
	Status           string            `json:"status"`
	ErrorDescription string            `json:"errorDescription"`
	UserAuditReport  []UserAuditReport `json:"user_audit_report"`
}

type UserAuditReportReqDto struct {
	UserName  string  `json:"user_name"`
	StartTime float64 `json:"start_time"`
	EndTime   float64 `json:"end_time"`
}

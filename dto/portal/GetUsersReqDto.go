package dto

type GetUsersReqDto struct {
	Page            int    `json:"page"`            // optional. Empty value will bring latest results sort by date descending.
	PageSize        int    `json:"pageSize"`        // optional. Empty value will bring 50 records. Value can't be more than 50.
	PartialUserName string `json:"partialUserName"` // optional.
}

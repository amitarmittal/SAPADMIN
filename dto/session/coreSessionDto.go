package dto

type B2BSessionDto struct {
	OperatorId string `bson:"operatorId"`
	PartnerId  string `bson:"partnerId"`
	UserId     string `bson:"userId"`
	UserName   string `bson:"userName"`
	Token      string `bson:"token"`
	UserKey    string `bson:"userKey"`
	CreatedAt  int64  `bson:"createdAt"`
	ExpireAt   int64  `bson:"expireAt"`
	BaseURL    string `bson:"baseURL"`
	UserIp     string `bson:"userIp"`
}

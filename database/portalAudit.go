package database

import (
	"Sp/dto/models"
	"Sp/dto/reports"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Insert Portal Audit Document
func InsertPortalAudit(audit models.Audit) error {
	//log.Println("InsertProvider: Adding Documnet for providerId - ", provider.ProviderId)
	audit.Time = time.Now().Round(0).UTC()
	_, err := PortalAuditCollection.InsertOne(Ctx, audit)
	if err != nil {
		log.Println("InsertPortalAudit: INSERT Portal Audit failed with error - ", err.Error())
		return err
	}
	return nil
}

// Get Audit Document
func GetAudit(operatorId string) (models.Audit, error) {
	audit := models.Audit{}
	err := PortalAuditCollection.FindOne(Ctx, bson.M{"operator_id": operatorId}).Decode(&audit)
	if err != nil {
		log.Println("GetAudit: Provider NOT FOUND - ", err.Error())
		log.Println("GetAudit: operator_id - ", operatorId)
		return audit, err
	}
	return audit, nil
}

func GetUserAuditReport(operatorId string, reqDto reports.UserAuditReportReqDto) ([]models.Audit, error) {
	var audits []models.Audit
	var filter = bson.M{}
	if reqDto.UserName != "" {
		filter["user_name"] = reqDto.UserName
	}
	if reqDto.StartTime > 0 && reqDto.EndTime > 0 {
		filter["time"] = bson.M{"$gte": time.Unix(int64(reqDto.StartTime), 0), "$lte": time.Unix(int64(reqDto.EndTime), 0)}
	}
	if operatorId != "" {
		filter["operator_id"] = operatorId
	}
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"time", -1}})
	cursor, err := PortalAuditCollection.Find(Ctx, filter, findOptions)
	if err != nil {
		log.Println("GetUserAuditReport: Find failed with error - ", err.Error())
		return audits, err
	}
	defer cursor.Close(Ctx)
	for cursor.Next(Ctx) {
		var audit models.Audit
		err := cursor.Decode(&audit)
		if err != nil {
			log.Println("GetUserAuditReport: cursor.Decode failed with error - ", err.Error())
			return audits, err
		}
		audits = append(audits, audit)
	}
	return audits, nil

}

package sports

import "go.mongodb.org/mongo-driver/bson/primitive"

type EventDto struct {
	ID              primitive.ObjectID `bson:"_id,omitempty"`
	EventKey        string             `bson:"event_key"`
	ProviderId      string             `bson:"provider_id"`
	ProviderName    string             `bson:"provider_name"`
	SportId         string             `bson:"sport_id"`
	SportName       string             `bson:"sport_name"`
	CompetitionId   string             `bson:"competition_id"`
	CompetitionName string             `bson:"competition_name"`
	EventId         string             `bson:"event_id"`
	EventName       string             `bson:"event_name"`
	OpenDate        int64              `bson:"open_date"`
}

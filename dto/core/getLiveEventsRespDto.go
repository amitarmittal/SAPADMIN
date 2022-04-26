package dto

import (
	"encoding/json"
	"log"
)

type GetLiveEventsRespDto struct {
	Status           string     `json:"status"`
	ErrorDescription string     `json:"errorDescription"`
	Events           []EventDto `json:"sports"`
}

func GetLiveEventsTestData() GetLiveEventsRespDto {
	getLiveEventsRespDto := GetLiveEventsRespDto{}
	getLiveEventsRespDto.Status = "RS_OK"
	getLiveEventsRespDto.ErrorDescription = ""
	getLiveEventsRespDto.Events = []EventDto{}

	liveEventDto := EventDto{}
	err := json.Unmarshal([]byte(CricketEventLive), &liveEventDto)
	if err != nil {
		log.Println("GetTestData: Unmarshal failed to convert test data into event - ", err.Error())
	} else {
		getLiveEventsRespDto.Events = append(getLiveEventsRespDto.Events, liveEventDto)
		getLiveEventsRespDto.Events = append(getLiveEventsRespDto.Events, liveEventDto)
	}
	err = json.Unmarshal([]byte(SoccerEventLive), &liveEventDto)
	if err != nil {
		log.Println("GetTestData: Unmarshal failed to convert test data into event - ", err.Error())
	} else {
		getLiveEventsRespDto.Events = append(getLiveEventsRespDto.Events, liveEventDto)
		getLiveEventsRespDto.Events = append(getLiveEventsRespDto.Events, liveEventDto)
		getLiveEventsRespDto.Events = append(getLiveEventsRespDto.Events, liveEventDto)
		getLiveEventsRespDto.Events = append(getLiveEventsRespDto.Events, liveEventDto)
		getLiveEventsRespDto.Events = append(getLiveEventsRespDto.Events, liveEventDto)
		getLiveEventsRespDto.Events = append(getLiveEventsRespDto.Events, liveEventDto)
	}
	err = json.Unmarshal([]byte(TennisEventLive), &liveEventDto)
	if err != nil {
		log.Println("GetTestData: Unmarshal failed to convert test data into event - ", err.Error())
	} else {
		getLiveEventsRespDto.Events = append(getLiveEventsRespDto.Events, liveEventDto)
		getLiveEventsRespDto.Events = append(getLiveEventsRespDto.Events, liveEventDto)
		getLiveEventsRespDto.Events = append(getLiveEventsRespDto.Events, liveEventDto)
		getLiveEventsRespDto.Events = append(getLiveEventsRespDto.Events, liveEventDto)
		getLiveEventsRespDto.Events = append(getLiveEventsRespDto.Events, liveEventDto)
		getLiveEventsRespDto.Events = append(getLiveEventsRespDto.Events, liveEventDto)
		getLiveEventsRespDto.Events = append(getLiveEventsRespDto.Events, liveEventDto)
		getLiveEventsRespDto.Events = append(getLiveEventsRespDto.Events, liveEventDto)
	}
	/*
		resData, err := json.Marshal(getLiveEventsRespDto)
		if err != nil {
			log.Println("GetTestData: Marshal failed to convert test data into json string - ", err.Error())
		}
		resStr := string(resData)
		log.Println("GetTestData: response String is - ", resStr)
	*/
	return getLiveEventsRespDto
}

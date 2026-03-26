package processor

import (
	"github.com/free5gc/openapi/models"
)

// UeIdReq - Request data for UE ID retrieval (TS 29.522)
type UeIdReq struct {
	AfId     string          `json:"afId" yaml:"afId" bson:"afId,omitempty"`
	Gpsi     string          `json:"gpsi,omitempty" yaml:"gpsi" bson:"gpsi,omitempty"`
	IpAddr   *models.IpAddr  `json:"ipAddr,omitempty" yaml:"ipAddr" bson:"ipAddr,omitempty"`
	MacAddr  string          `json:"macAddr,omitempty" yaml:"macAddr" bson:"macAddr,omitempty"`
	SuppFeat string          `json:"suppFeat,omitempty" yaml:"suppFeat" bson:"suppFeat,omitempty"`
}

// UeIdInfo - Response data for UE ID retrieval (TS 29.522)
type UeIdInfo struct {
	UeIdResults []UeIdResult `json:"ueIdResults" yaml:"ueIdResults" bson:"ueIdResults,omitempty"`
	SuppFeat    string       `json:"suppFeat,omitempty" yaml:"suppFeat" bson:"suppFeat,omitempty"`
}

// UeIdResult - Individual UE ID translation result
type UeIdResult struct {
	Gpsi       string `json:"gpsi,omitempty" yaml:"gpsi" bson:"gpsi,omitempty"`
	ExternalId string `json:"externalId,omitempty" yaml:"externalId" bson:"externalId,omitempty"`
}

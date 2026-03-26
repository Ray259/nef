package sbi

import (
	"net/http"

	"github.com/free5gc/nef/internal/logger"
	"github.com/free5gc/openapi"
	"github.com/free5gc/nef/internal/sbi/processor"
	"github.com/free5gc/util/metrics/sbi"
	"github.com/gin-gonic/gin"
)

func (s *Server) getUEIdentifierRoutes() []Route {
	return []Route{
		{
			Method:  http.MethodPost,
			Pattern: "/retrieve",
			APIFunc: s.apiPostRetrieveUEId,
		},
	}
}

func (s *Server) apiPostRetrieveUEId(gc *gin.Context) {
	var ueIdReq processor.UeIdReq
	reqBody, err := gc.GetRawData()
	if err != nil {
		logger.SBILog.Errorf("Get Request Body error: %+v", err)
		pd := openapi.ProblemDetailsSystemFailure(err.Error())
		gc.Set(sbi.IN_PB_DETAILS_CTX_STR, pd.Cause)
		gc.JSON(http.StatusInternalServerError, pd)
		return
	}

	err = openapi.Deserialize(&ueIdReq, reqBody, "application/json")
	if err != nil {
		logger.SBILog.Errorf("Deserialize Request Body error: %+v", err)
		pd := openapi.ProblemDetailsMalformedReqSyntax(err.Error())
		gc.Set(sbi.IN_PB_DETAILS_CTX_STR, pd.Cause)
		gc.JSON(http.StatusBadRequest, pd)
		return
	}

	s.Processor().RetrieveUEId(gc, &ueIdReq)
}

package processor

import (
	"net/http"

	"github.com/free5gc/openapi"
	"github.com/free5gc/util/metrics/sbi"
	"github.com/gin-gonic/gin"
)

func (p *Processor) RetrieveUEId(gc *gin.Context, ueIdReq *UeIdReq) {
	if ueIdReq.Gpsi == "" {
		pd := openapi.ProblemDetailsMalformedReqSyntax("Missing Gpsi in request")
		gc.Set(sbi.IN_PB_DETAILS_CTX_STR, pd.Cause)
		gc.JSON(http.StatusBadRequest, pd)
		return
	}

	// 1. Resolve identity from UDM
	idTranslation, pd, err := p.Consumer().GetIdTranslationResult(ueIdReq.Gpsi)

	if err != nil {
		pd := openapi.ProblemDetailsSystemFailure(err.Error())
		gc.Set(sbi.IN_PB_DETAILS_CTX_STR, pd.Cause)
		gc.JSON(http.StatusInternalServerError, pd)
		return
	}

	if pd != nil {
		gc.Set(sbi.IN_PB_DETAILS_CTX_STR, pd.Cause)
		gc.JSON(int(pd.Status), pd)
		return
	}

	// 2. Build Response
	ueIdInfo := &UeIdInfo{
		UeIdResults: []UeIdResult{
			{
				Gpsi: idTranslation.Gpsi,
			},
		},
	}

	gc.JSON(http.StatusOK, ueIdInfo)
}

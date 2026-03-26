package processor

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/free5gc/openapi/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestRetrieveUEId(t *testing.T) {
	// 1. Mock NRF Discovery for UDM
	gock.New("http://127.0.0.10:8000/nnrf-disc/v1").
		Get("/nf-instances").
		MatchParam("service-names", "nudm-sdm").
		MatchParam("target-nf-type", "UDM").
		MatchParam("requester-nf-type", "NEF").
		Reply(http.StatusOK).
		JSON(models.SearchResult{
			NfInstances: []models.NrfNfDiscoveryNfProfile{
				{
					NfInstanceId: "udm-1",
					NfType:       models.NrfNfManagementNfType_UDM,
					NfStatus:     models.NrfNfManagementNfStatus_REGISTERED,
					Ipv4Addresses: []string{"127.0.0.3"},
					NfServices: []models.NrfNfDiscoveryNfService{
						{
							ServiceInstanceId: "sdm-1",
							ServiceName:       models.ServiceName_NUDM_SDM,
							NfServiceStatus:   models.NfServiceStatus_REGISTERED,
							Versions: []models.NfServiceVersion{
								{ApiFullVersion: "1.0.0", ApiVersionInUri: "v2"},
							},
							Scheme: models.UriScheme_HTTP,
							IpEndPoints: []models.IpEndPoint{
								{Ipv4Address: "127.0.0.3", Port: 8000},
							},
							ApiPrefix: "http://127.0.0.3:8000",
						},
					},
				},
			},
		})

	// 2. Mock UDM SDM Identity Translation
	idTranslationRsp := models.IdTranslationResult{
		SupportedFeatures: "1",
		Gpsi:              "msisdn-12345",
		Supi:              "imsi-001010000000001",
	}

	gock.New("http://127.0.0.3:8000").
		Get("/nudm-sdm/v2/msisdn-12345/id-translation-result").
		Reply(http.StatusOK).
		JSON(idTranslationRsp)

	defer gock.Off()

	testCases := []struct {
		description      string
		ueIdReq          UeIdReq
		expectedResponse *HandlerResponse
	}{
		{
			description: "TC1: Successful UE ID retrieval",
			ueIdReq: UeIdReq{
				AfId: "af1",
				Gpsi: "msisdn-12345",
			},
			expectedResponse: &HandlerResponse{
				Status: http.StatusOK,
				Body: &UeIdInfo{
					UeIdResults: []UeIdResult{
						{
							Gpsi: "msisdn-12345",
						},
					},
				},
			},
		},
		{
			description: "TC2: Missing GPSI",
			ueIdReq: UeIdReq{
				AfId: "af1",
			},
			expectedResponse: &HandlerResponse{
				Status: http.StatusBadRequest,
				Body: &models.ProblemDetails{
					Status: http.StatusBadRequest,
					Title:  "Malformed request syntax",
					Detail: "Missing Gpsi in request",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			httpRecorder := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(httpRecorder)

			nefApp.Processor().RetrieveUEId(c, &tc.ueIdReq)
			require.Equal(t, tc.expectedResponse.Status, httpRecorder.Code)

			var actualRsp interface{}
			err := json.Unmarshal(httpRecorder.Body.Bytes(), &actualRsp)
			require.NoError(t, err)

			// Special check: ensure SUPI is NOT in the response
			var rawMap map[string]interface{}
			err = json.Unmarshal(httpRecorder.Body.Bytes(), &rawMap)
			require.NoError(t, err)
			
			// If it's UeIdInfo, check nested results
			if results, ok := rawMap["ueIdResults"].([]interface{}); ok {
				for _, res := range results {
					if resMap, ok := res.(map[string]interface{}); ok {
						require.Nil(t, resMap["supi"], "SUPI must not be exposed")
					}
				}
			}

			assertJSONBodyEqual(t, tc.expectedResponse.Body, httpRecorder.Body.Bytes())
		})
	}
}

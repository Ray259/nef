package consumer

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/free5gc/openapi"
	"github.com/free5gc/openapi/models"
	"github.com/free5gc/openapi/nrf/NFDiscovery"
	"github.com/free5gc/openapi/udm/SubscriberDataManagement"
	sbi_metrics "github.com/free5gc/util/metrics/sbi"
)

type nudmService struct {
	consumer *Consumer

	mu      sync.RWMutex
	clients map[string]*SubscriberDataManagement.APIClient
}

func (s *nudmService) getSubscriberDataManagementClient(uri string) *SubscriberDataManagement.APIClient {
	if uri == "" {
		return nil
	}

	s.mu.RLock()
	client, ok := s.clients[uri]
	if ok {
		defer s.mu.RUnlock()
		return client
	}

	configuration := SubscriberDataManagement.NewConfiguration()
	configuration.SetBasePath(uri)
	configuration.SetMetrics(sbi_metrics.SbiMetricHook)
	configuration.SetHTTPClient(http.DefaultClient)
	client = SubscriberDataManagement.NewAPIClient(configuration)

	s.mu.RUnlock()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients[uri] = client
	return client
}

func (s *nudmService) getUdmSdmUri() (string, error) {
	// Note: In a real implementation, we might want to cache this in the context
	localVarOptionals := NFDiscovery.SearchNFInstancesRequest{
		ServiceNames: []models.ServiceName{
			models.ServiceName_NUDM_SDM,
		},
	}
	_, sUri, err := s.consumer.SearchNFInstances(s.consumer.Config().NrfUri(),
		models.ServiceName_NUDM_SDM, models.NrfNfManagementNfType_UDM, models.NrfNfManagementNfType_NEF, &localVarOptionals)
	return sUri, err
}

// GetIdTranslationResult retrieves the SUPI/GPSI mapping from UDM.
// 3GPP TS 29.503 Clause 5.2.2.2.14
func (s *nudmService) GetIdTranslationResult(gpsi string) (*models.IdTranslationResult, *models.ProblemDetails, error) {
	uri, err := s.getUdmSdmUri()
	if err != nil {
		return nil, nil, err
	}

	client := s.getSubscriberDataManagementClient(uri)
	if client == nil {
		return nil, nil, fmt.Errorf("could not initialize the SubscriberDataManagement client")
	}

	ctx, _, err := s.consumer.Context().GetTokenCtx(models.ServiceName_NUDM_SDM, models.NrfNfManagementNfType_UDM)
	if err != nil {
		return nil, nil, err
	}

	req := &SubscriberDataManagement.GetSupiOrGpsiRequest{}
	req.SetUeId(gpsi)

	res, err := client.GPSIToSUPITranslationOrSUPIToGPSITranslationApi.GetSupiOrGpsi(ctx, req)
	if err != nil {
		switch apiErr := err.(type) {
		case openapi.GenericOpenAPIError:
			switch errorModel := apiErr.Model().(type) {
			case SubscriberDataManagement.GetSupiOrGpsiError:
				return nil, &errorModel.ProblemDetails, nil
			default:
				return nil, nil, openapi.ReportError("openapi error")
			}
		default:
			return nil, nil, err
		}
	}

	return &res.IdTranslationResult, nil, nil
}

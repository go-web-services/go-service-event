package service

import (
	"fmt"

	"github.com/gin-gonic/gin"
	platformUtils "github.com/Lomank123/go-web-platform/utils"

	"github.com/Lomank123/go-service-event/pkg/client/dto"
)

type EventAPIService interface {
	CreateV1(ctx *gin.Context, input dto.EventCreateInputDTO) (dto.EventCreateOutputDTO, error)
	CreateBatchV1(ctx *gin.Context, input dto.EventCreateBatchInputDTO) (dto.EventCreateBatchOutputDTO, error)
	UpdateV1(ctx *gin.Context, input dto.EventUpdateInputDTO) (dto.EventUpdateOutputDTO, error)
	DeleteV1(ctx *gin.Context, input dto.EventDeleteInputDTO) (dto.EventDeleteOutputDTO, error)
	DetailV1(ctx *gin.Context, input dto.EventDetailInputDTO) (dto.EventDetailOutputDTO, error)
	QueryV1(ctx *gin.Context, input dto.EventQueryInputDTO) (dto.EventQueryOutputDTO, error)
}

type eventAPIService struct {
	apiURL string
}

func NewEventAPIService(host string) EventAPIService {
	return &eventAPIService{
		apiURL: fmt.Sprintf("%s/api/v1", host),
	}
}

func (s *eventAPIService) CreateV1(ctx *gin.Context, input dto.EventCreateInputDTO) (dto.EventCreateOutputDTO, error) {
	baseURL := fmt.Sprintf("%s/events/create", s.apiURL)
	var output dto.EventCreateOutputDTO
	err := platformUtils.SendRequest("POST", baseURL, input, &output, ctx)
	return output, err
}

func (s *eventAPIService) CreateBatchV1(ctx *gin.Context, input dto.EventCreateBatchInputDTO) (dto.EventCreateBatchOutputDTO, error) {
	baseURL := fmt.Sprintf("%s/events/create-batch", s.apiURL)
	var output dto.EventCreateBatchOutputDTO
	err := platformUtils.SendRequest("POST", baseURL, input, &output, ctx)
	return output, err
}

func (s *eventAPIService) UpdateV1(ctx *gin.Context, input dto.EventUpdateInputDTO) (dto.EventUpdateOutputDTO, error) {
	baseURL := fmt.Sprintf("%s/events/update", s.apiURL)
	var output dto.EventUpdateOutputDTO
	err := platformUtils.SendRequest("POST", baseURL, input, &output, ctx)
	return output, err
}

func (s *eventAPIService) DeleteV1(ctx *gin.Context, input dto.EventDeleteInputDTO) (dto.EventDeleteOutputDTO, error) {
	baseURL := fmt.Sprintf("%s/events/delete", s.apiURL)
	var output dto.EventDeleteOutputDTO
	err := platformUtils.SendRequest("POST", baseURL, input, &output, ctx)
	return output, err
}

func (s *eventAPIService) DetailV1(ctx *gin.Context, input dto.EventDetailInputDTO) (dto.EventDetailOutputDTO, error) {
	baseURL := fmt.Sprintf("%s/events/detail", s.apiURL)
	var output dto.EventDetailOutputDTO
	err := platformUtils.SendRequest("POST", baseURL, input, &output, ctx)
	return output, err
}

func (s *eventAPIService) QueryV1(ctx *gin.Context, input dto.EventQueryInputDTO) (dto.EventQueryOutputDTO, error) {
	baseURL := fmt.Sprintf("%s/events/query", s.apiURL)
	var output dto.EventQueryOutputDTO
	err := platformUtils.SendRequest("POST", baseURL, input, &output, ctx)
	return output, err
}

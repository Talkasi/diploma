package dto

import "logistics/internal/domain"

// CityResponse DTO для ответа API.
type CityResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// FromDomain преобразует доменную модель в DTO.
func FromDomain(c domain.City) CityResponse {
	return CityResponse{ID: c.ID, Name: c.Name}
}

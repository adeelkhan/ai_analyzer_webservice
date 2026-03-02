package services

import (
	"github.com/adeelkhan/analytics_service/internal/models"
)

type AnalyticsService struct {
	repo *models.Repository
}

func NewAnalyticsService() *AnalyticsService {
	return &AnalyticsService{
		repo: models.DefaultRepository(),
	}
}

func (as *AnalyticsService) GetUsers() ([]models.User, error) {
	return as.repo.GetUsers()
}

func (as *AnalyticsService) CalculateSumAge() (models.SumAgeResponse, error) {
	return as.repo.CalculateSumAge()
}

func (as *AnalyticsService) FindUsersByCountry(country string) (models.UsersByCountryResponse, error) {
	return as.repo.FindUsersByCountry(country)
}

func (as *AnalyticsService) FindUsersOlderThan(age int) (models.UsersOlderThanResponse, error) {
	return as.repo.FindUsersOlderThan(age)
}

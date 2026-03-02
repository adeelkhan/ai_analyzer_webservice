package handlers

import (
	"net/http"

	"github.com/adeelkhan/code_diff/internal/logger"
	"github.com/adeelkhan/code_diff/internal/models"
	"github.com/adeelkhan/code_diff/internal/services"

	"github.com/gin-gonic/gin"
)

var csvLog = logger.GetLogger()

var analyticsService = services.NewAnalyticsService()

// UsersResponse represents response for getting all users
type UsersResponse struct {
	Users []models.User `json:"users"`
	Count int           `json:"count"`
}

type UsersByCountryRequest struct {
	Country string `json:"country" binding:"required"`
}

type UsersOlderThanRequest struct {
	Age int `json:"age" binding:"required"`
}

// SumAgeHandler returns sum of ages for all users
func SumAgeHandler(c *gin.Context) {
	csvLog.Info("SumAge analytics requested")

	response, err := analyticsService.CalculateSumAge()
	if err != nil {
		csvLog.Error("Failed to calculate sum age: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	csvLog.Info("SumAge analytics completed - Total users: %d, Sum of ages: %d, Average: %d",
		response.TotalUsers, response.SumOfAges, response.AverageAge)
	c.JSON(http.StatusOK, response)
}

// UsersByCountryHandler returns users from a specific country
func UsersByCountryHandler(c *gin.Context) {
	var req UsersByCountryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		csvLog.Warn("UsersByCountry failed: invalid request - %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	csvLog.Info("UsersByCountry analytics requested for country: %s", req.Country)

	response, err := analyticsService.FindUsersByCountry(req.Country)
	if err != nil {
		csvLog.Error("Failed to find users by country %s: %v", req.Country, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	csvLog.Info("UsersByCountry analytics completed - Found %d users from %s",
		response.Count, response.Country)
	c.JSON(http.StatusOK, response)
}

// UsersOlderThanHandler returns users older than specified age
func UsersOlderThanHandler(c *gin.Context) {
	var req UsersOlderThanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		csvLog.Warn("UsersOlderThan failed: invalid request - %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	csvLog.Info("UsersOlderThan analytics requested for age: %d", req.Age)

	response, err := analyticsService.FindUsersOlderThan(req.Age)
	if err != nil {
		csvLog.Error("Failed to find users older than %d: %v", req.Age, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	csvLog.Info("UsersOlderThan analytics completed - Found %d users older than %d",
		response.Count, response.MinAge)
	c.JSON(http.StatusOK, response)
}

// GetUsersHandler returns all users - uses service layer
func GetUsersHandler(c *gin.Context) {
	csvLog.Info("GetUsers requested")

	users, err := analyticsService.GetUsers()
	if err != nil {
		csvLog.Error("Failed to get users: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	csvLog.Info("GetUsers completed - Found %d users", len(users))
	c.JSON(http.StatusOK, UsersResponse{
		Users: users,
		Count: len(users),
	})
}

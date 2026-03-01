package handlers

import (
	"net/http"

	"github.com/adeelkhan/code_diff/logger"
	"github.com/adeelkhan/code_diff/models"

	"github.com/gin-gonic/gin"
)

var csvLog = logger.GetLogger()

type SumAgeRequest struct{}

type UsersByCountryRequest struct {
	Country string `json:"country" binding:"required"`
}

type UsersOlderThanRequest struct {
	Age int `json:"age" binding:"required"`
}

// SumAge returns sum of ages for all users
func SumAge(c *gin.Context) {
	csvLog.Info("SumAge analytics requested")

	repo := models.DefaultUserRepository()
	response, err := repo.CalculateSumAge()
	if err != nil {
		csvLog.Error("Failed to calculate sum age: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	csvLog.Info("SumAge analytics completed - Total users: %d, Sum of ages: %d, Average: %d",
		response.TotalUsers, response.SumOfAges, response.AverageAge)
	c.JSON(http.StatusOK, response)
}

// UsersByCountry returns users from a specific country
func UsersByCountry(c *gin.Context) {
	var req UsersByCountryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		csvLog.Warn("UsersByCountry failed: invalid request - %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	csvLog.Info("UsersByCountry analytics requested for country: %s", req.Country)

	repo := models.DefaultUserRepository()
	response, err := repo.FindUsersByCountry(req.Country)
	if err != nil {
		csvLog.Error("Failed to find users by country %s: %v", req.Country, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	csvLog.Info("UsersByCountry analytics completed - Found %d users from %s",
		response.Count, response.Country)
	c.JSON(http.StatusOK, response)
}

// UsersOlderThan returns users older than specified age
func UsersOlderThan(c *gin.Context) {
	var req UsersOlderThanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		csvLog.Warn("UsersOlderThan failed: invalid request - %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	csvLog.Info("UsersOlderThan analytics requested for age: %d", req.Age)

	repo := models.DefaultUserRepository()
	response, err := repo.FindUsersOlderThan(req.Age)
	if err != nil {
		csvLog.Error("Failed to find users older than %d: %v", req.Age, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	csvLog.Info("UsersOlderThan analytics completed - Found %d users older than %d",
		response.Count, response.MinAge)
	c.JSON(http.StatusOK, response)
}

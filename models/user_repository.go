package models

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

type User struct {
	Name    string `json:"name"`
	Age     int    `json:"age"`
	Country string `json:"country"`
}

// UserRepository handles CSV data operations
type UserRepository struct {
	csvPath string
}

// NewUserRepository creates a new UserRepository instance
func NewUserRepository(csvPath string) *UserRepository {
	return &UserRepository{
		csvPath: csvPath,
	}
}

// DefaultUserRepository creates a UserRepository with the default CSV location
func DefaultUserRepository() *UserRepository {
	csvPath := filepath.Join(getProjectRoot(), "data", "users.csv")
	return NewUserRepository(csvPath)
}

func getProjectRoot() string {
	_, b, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(b), "..")
}

// LoadData loads user data from CSV file
func (ur *UserRepository) LoadData() ([]User, error) {
	file, err := os.Open(ur.csvPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) == 0 {
		return []User{}, nil
	}

	var users []User
	for i, record := range records {
		if i == 0 {
			continue
		}
		if len(record) < 3 {
			continue
		}

		age, err := strconv.Atoi(strings.TrimSpace(record[1]))
		if err != nil {
			continue
		}

		users = append(users, User{
			Name:    strings.TrimSpace(record[0]),
			Age:     age,
			Country: strings.TrimSpace(record[2]),
		})
	}

	return users, nil
}

// GetUsers returns all users
func (ur *UserRepository) GetUsers() ([]User, error) {
	return ur.LoadData()
}

// CalculateSumAge calculates sum and average of all ages
type SumAgeResponse struct {
	TotalUsers int `json:"total_users"`
	SumOfAges  int `json:"sum_of_ages"`
	AverageAge int `json:"average_age"`
}

func (ur *UserRepository) CalculateSumAge() (SumAgeResponse, error) {
	users, err := ur.LoadData()
	if err != nil {
		return SumAgeResponse{}, err
	}

	sum := 0
	for _, user := range users {
		sum += user.Age
	}

	avg := 0
	if len(users) > 0 {
		avg = sum / len(users)
	}

	return SumAgeResponse{
		TotalUsers: len(users),
		SumOfAges:  sum,
		AverageAge: avg,
	}, nil
}

// FindUsersByCountry filters users by country
type UsersByCountryResponse struct {
	Country string `json:"country"`
	Count   int    `json:"count"`
	Users   []User `json:"users"`
}

func (ur *UserRepository) FindUsersByCountry(country string) (UsersByCountryResponse, error) {
	users, err := ur.LoadData()
	if err != nil {
		return UsersByCountryResponse{}, err
	}

	var filtered []User
	for _, user := range users {
		if strings.EqualFold(user.Country, country) {
			filtered = append(filtered, user)
		}
	}

	return UsersByCountryResponse{
		Country: country,
		Count:   len(filtered),
		Users:   filtered,
	}, nil
}

// FindUsersOlderThan filters users by age
type UsersOlderThanResponse struct {
	MinAge int    `json:"min_age"`
	Count  int    `json:"count"`
	Users  []User `json:"users"`
}

func (ur *UserRepository) FindUsersOlderThan(age int) (UsersOlderThanResponse, error) {
	users, err := ur.LoadData()
	if err != nil {
		return UsersOlderThanResponse{}, err
	}

	var filtered []User
	for _, user := range users {
		if user.Age > age {
			filtered = append(filtered, user)
		}
	}

	return UsersOlderThanResponse{
		MinAge: age,
		Count:  len(filtered),
		Users:  filtered,
	}, nil
}

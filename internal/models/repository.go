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

// Repository handles CSV data operations
type Repository struct {
	csvPath string
}

// NewRepository creates a new Repository instance
func NewRepository(csvPath string) *Repository {
	return &Repository{
		csvPath: csvPath,
	}
}

// DefaultRepository creates a Repository with the default CSV location
func DefaultRepository() *Repository {
	csvPath := filepath.Join(getProjectRoot(), "data", "users.csv")
	return NewRepository(csvPath)
}
func getProjectRoot() string {
	_, b, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(b), "..", "..")
}

// LoadData loads user data from CSV file
func (r *Repository) LoadData() ([]User, error) {
	file, err := os.Open(r.csvPath)
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
func (r *Repository) GetUsers() ([]User, error) {
	return r.LoadData()
}

// CalculateSumAge calculates sum and average of all ages
type SumAgeResponse struct {
	TotalUsers int `json:"total_users"`
	SumOfAges  int `json:"sum_of_ages"`
	AverageAge int `json:"average_age"`
}

func (r *Repository) CalculateSumAge() (SumAgeResponse, error) {
	users, err := r.LoadData()
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

func (r *Repository) FindUsersByCountry(country string) (UsersByCountryResponse, error) {
	users, err := r.LoadData()
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

func (r *Repository) FindUsersOlderThan(age int) (UsersOlderThanResponse, error) {
	users, err := r.LoadData()
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

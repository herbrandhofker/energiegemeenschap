package service_db

import (
	"context"
	"fmt"
	"time"

	"ws/internal/client"
	"ws/internal/model"
)

// UserService handles user-related operations
type UserService struct {
	Client *client.TibberClient
}

// GetUserData fetches user information from Tibber API
func (s *UserService) GetUserData(ctx context.Context) (*model.User, error) {
	// Use the query constant from the model package
	resp, err := s.Client.QueryAPI(ctx, model.UserQuery, nil)
	if err != nil {
		return nil, err
	}

	// Extract viewer data and build user model
	viewerData, ok := resp.Data["viewer"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("no viewer data in response")
	}

	// Create user model directly
	user := &model.User{
		Name:        client.GetString(viewerData, "name"),
		ID:          client.GetString(viewerData, "userId"),
		Email:       client.GetString(viewerData, "login"),
		AccountType: client.GetString(viewerData, "accountType"),
		LastLogin:   time.Now(),
		Homes:       []model.Home{},
	}

	// Parse homes
	if homesData, ok := viewerData["homes"].([]interface{}); ok {
		for _, homeRaw := range homesData {
			if homeData, ok := homeRaw.(map[string]interface{}); ok {
				home := model.Home{
					Id:                client.GetString(homeData, "id"),
					TimeZone:          client.GetString(homeData, "timeZone"),
					Type:              client.GetString(homeData, "type"),
					Size:              client.GetInt(homeData, "size"),
					NumberOfResidents: client.GetInt(homeData, "numberOfResidents"),
					AppNickname:       client.GetString(homeData, "appNickname"),
				}
				user.Homes = append(user.Homes, home)
			}
		}
	}

	return user, nil
}

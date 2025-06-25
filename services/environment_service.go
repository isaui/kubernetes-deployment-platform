package services

import (
	"errors"
	"fmt"

	"github.com/pendeploy-simple/models"
	"github.com/pendeploy-simple/repositories"
)

// EnvironmentService handles business logic for environments
type EnvironmentService struct {
	environmentRepo *repositories.EnvironmentRepository
	projectRepo     *repositories.ProjectRepository
}

// NewEnvironmentService creates a new environment service instance
func NewEnvironmentService() *EnvironmentService {
	return &EnvironmentService{
		environmentRepo: repositories.NewEnvironmentRepository(),
		projectRepo:     repositories.NewProjectRepository(),
	}
}

// ListEnvironments retrieves all environments for a project
func (s *EnvironmentService) ListEnvironments(projectID string, userID string, isAdmin bool) ([]models.Environment, error) {
	// Check if user can access this project
	if !isAdmin {
		ownerID, err := s.projectRepo.GetOwnerID(projectID)
		if err != nil {
			return nil, err
		}
		
		if ownerID != userID {
			return nil, errors.New("unauthorized access to project environments")
		}
	}
	
	return s.environmentRepo.FindByProjectID(projectID)
}

// GetEnvironmentDetail retrieves a specific environment
func (s *EnvironmentService) GetEnvironmentDetail(environmentID string, userID string, isAdmin bool) (models.Environment, error) {
	// Fetch the environment
	env, err := s.environmentRepo.FindByID(environmentID)
	if err != nil {
		return env, err
	}
	
	// Check if user can access this environment
	if !isAdmin {
		ownerID, err := s.projectRepo.GetOwnerID(env.ProjectID)
		if err != nil {
			return env, err
		}
		
		if ownerID != userID {
			return models.Environment{}, errors.New("unauthorized access to environment")
		}
	}
	
	return env, nil
}

// CreateEnvironment creates a new environment for a project
func (s *EnvironmentService) CreateEnvironment(env models.Environment, userID string, isAdmin bool) (models.Environment, error) {
	// Check if user can access this project
	if !isAdmin {
		ownerID, err := s.projectRepo.GetOwnerID(env.ProjectID)
		if err != nil {
			return env, err
		}
		
		if ownerID != userID {
			return models.Environment{}, errors.New("unauthorized to create environment for this project")
		}
	}
	
	// Validate environment name uniqueness within project
	exists, err := s.environmentRepo.ExistsByNameAndProject(env.Name, env.ProjectID)
	if err != nil {
		return env, err
	}
	
	if exists {
		return models.Environment{}, fmt.Errorf("environment with name '%s' already exists in this project", env.Name)
	}
	
	// Create the environment
	return s.environmentRepo.Create(env)
}

// UpdateEnvironment updates an existing environment
func (s *EnvironmentService) UpdateEnvironment(env models.Environment, userID string, isAdmin bool) (models.Environment, error) {
	// Fetch current environment
	currentEnv, err := s.environmentRepo.FindByID(env.ID)
	if err != nil {
		return env, err
	}
	
	// Check if user can access this project
	if !isAdmin {
		ownerID, err := s.projectRepo.GetOwnerID(currentEnv.ProjectID)
		if err != nil {
			return env, err
		}
		
		if ownerID != userID {
			return models.Environment{}, errors.New("unauthorized to update environment")
		}
	}
	
	// If name is changing, check uniqueness
	if env.Name != currentEnv.Name {
		exists, err := s.environmentRepo.ExistsByNameAndProject(env.Name, currentEnv.ProjectID)
		if err != nil {
			return env, err
		}
		
		if exists {
			return models.Environment{}, fmt.Errorf("environment with name '%s' already exists in this project", env.Name)
		}
	}
	
	// Update only allowed fields
	currentEnv.Name = env.Name
	currentEnv.Description = env.Description
	
	// Save changes
	err = s.environmentRepo.Update(currentEnv)
	if err != nil {
		return currentEnv, err
	}
	
	return currentEnv, nil
}

// DeleteEnvironment removes an environment if it has no associated services
func (s *EnvironmentService) DeleteEnvironment(environmentID string, userID string, isAdmin bool) error {
	// Fetch the environment
	env, err := s.environmentRepo.FindByID(environmentID)
	if err != nil {
		return err
	}
	
	// Check if user can access this project
	if !isAdmin {
		ownerID, err := s.projectRepo.GetOwnerID(env.ProjectID)
		if err != nil {
			return err
		}
		
		if ownerID != userID {
			return errors.New("unauthorized to delete environment")
		}
	}
	
	// Check if environment has services
	count, err := s.environmentRepo.CountServicesInEnvironment(environmentID)
	if err != nil {
		return err
	}
	
	if count > 0 {
		return errors.New("cannot delete environment that has services")
	}
	
	// Delete the environment
	return s.environmentRepo.Delete(environmentID)
}

package admin

import (
	"context"
	"errors"
	"fmt"
	"net/mail"

	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/pkg/nameseeds"
	"go.uber.org/zap"
)

func (s *Service) CreateOrUpdateUser(ctx context.Context, email, name, photoURL string) (*database.User, error) {
	// Validate email address
	_, err := mail.ParseAddress(email)
	if err != nil {
		return nil, fmt.Errorf("invalid user email address %q", email)
	}

	// Update user if exists
	user, err := s.DB.FindUserByEmail(ctx, email)
	if err == nil {
		return s.DB.UpdateUser(ctx, user.ID, name, photoURL)
	} else if !errors.Is(err, database.ErrNotFound) {
		return nil, err
	}

	// User does not exist. Creating a new user.
	user, err = s.DB.InsertUser(ctx, email, name, photoURL)
	if err != nil {
		return nil, err
	}

	// We create an initial org with a name derived from the user's info
	err = s.createOrgForUser(ctx, email, user.ID, name)
	if err != nil {
		s.logger.Error("failed to create organization for user", zap.String("user.id", user.ID), zap.Error(err))
		// continuing, since user was created successfully
	}

	return user, nil
}

func (s *Service) createOrgForUser(ctx context.Context, email, userID, name string) error {
	// Start a tx for creating org and adding the user
	ctx, tx, err := s.DB.NewTx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	orgNameSeeds := nameseeds.ForUser(email, name)

	org, err := s.DB.InsertOrganizationFromSeeds(ctx, orgNameSeeds, name)
	if err != nil {
		return err
	}
	_, err = s.prepareOrg(ctx, org.ID, userID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Service) CreateOrgForUser(ctx context.Context, userID, orgName, description string) (*database.Organization, error) {
	ctx, tx, err := s.DB.NewTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	org, err := s.DB.InsertOrganization(ctx, orgName, description)
	if err != nil {
		return nil, err
	}

	org, err = s.prepareOrg(ctx, org.ID, userID)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return org, nil
}

func (s *Service) prepareOrg(ctx context.Context, orgID, userID string) (*database.Organization, error) {
	// create all user group for this org
	userGroup, err := s.DB.InsertUserGroup(ctx, "all_user_group", orgID, "All users in the organization")
	if err != nil {
		return nil, err
	}
	// update org with all user group
	org, err := s.DB.UpdateOrganizationAllUserGroup(ctx, orgID, userGroup.ID)
	if err != nil {
		return nil, err
	}

	// Add user to created org with org admin role
	err = s.DB.AddOrganizationMember(ctx, orgID, userID, database.RoleIDOrgAdmin)
	if err != nil {
		return nil, err
	}
	// Add user to all user group
	err = s.DB.AddUserGroupMember(ctx, userGroup.ID, userID)
	if err != nil {
		return nil, err
	}
	return org, nil
}

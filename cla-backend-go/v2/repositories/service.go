// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT

package repositories

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/communitybridge/easycla/cla-backend-go/github/branch_protection"

	"github.com/sirupsen/logrus"

	"github.com/communitybridge/easycla/cla-backend-go/gen/v2/models"
	"github.com/go-openapi/swag"

	"github.com/communitybridge/easycla/cla-backend-go/utils"

	"github.com/communitybridge/easycla/cla-backend-go/github"

	"github.com/aws/aws-sdk-go/aws"
	v1Models "github.com/communitybridge/easycla/cla-backend-go/gen/v1/models"
	v2Models "github.com/communitybridge/easycla/cla-backend-go/gen/v2/models"
	log "github.com/communitybridge/easycla/cla-backend-go/logging"
	"github.com/communitybridge/easycla/cla-backend-go/projects_cla_groups"
	v1Repositories "github.com/communitybridge/easycla/cla-backend-go/repositories"
	v2ProjectService "github.com/communitybridge/easycla/cla-backend-go/v2/project-service"
)

// Service contains functions of GitHub V3Repositories service
type Service interface {
	AddGithubRepositories(ctx context.Context, projectSFID string, input *models.GithubRepositoryInput) ([]*v1Models.GithubRepository, error)
	EnableRepository(ctx context.Context, repositoryID string) error
	DisableRepository(ctx context.Context, repositoryID string) error
	ListProjectRepositories(ctx context.Context, projectSFID string) (*v1Models.ListGithubRepositories, error)
	GetRepository(ctx context.Context, repositoryID string) (*v1Models.GithubRepository, error)
	GetRepositoryByName(ctx context.Context, repositoryName string) (*v1Models.GithubRepository, error)
	DisableCLAGroupRepositories(ctx context.Context, claGroupID string) error
	GetProtectedBranch(ctx context.Context, projectSFID, repositoryID, branchName string) (*v2Models.GithubRepositoryBranchProtection, error)
	UpdateProtectedBranch(ctx context.Context, projectSFID, repositoryID string, input *v2Models.GithubRepositoryBranchProtectionInput) (*v2Models.GithubRepositoryBranchProtection, error)
}

// GithubOrgRepo provide method to get github organization by name
type GithubOrgRepo interface {
	GetGithubOrganizationByName(ctx context.Context, githubOrganizationName string) (*v1Models.GithubOrganizations, error)
	GetGithubOrganization(ctx context.Context, githubOrganizationName string) (*v1Models.GithubOrganization, error)
	GetGithubOrganizations(ctx context.Context, projectSFID string) (*v1Models.GithubOrganizations, error)
}

type service struct {
	repo                  v1Repositories.Repository
	projectsClaGroupsRepo projects_cla_groups.Repository
	ghOrgRepo             GithubOrgRepo
}

var (
	requiredBranchProtectionChecks = []string{"EasyCLA"}
	// ErrInvalidBranchProtectionName is returned when invalid protection option is supplied
	ErrInvalidBranchProtectionName = errors.New("invalid protection option")
)

// NewService creates a new githubOrganizations service
func NewService(repo v1Repositories.Repository, pcgRepo projects_cla_groups.Repository, ghOrgRepo GithubOrgRepo) Service {
	return &service{
		repo:                  repo,
		projectsClaGroupsRepo: pcgRepo,
		ghOrgRepo:             ghOrgRepo,
	}
}

func (s *service) AddGithubRepositories(ctx context.Context, projectSFID string, input *models.GithubRepositoryInput) ([]*v1Models.GithubRepository, error) {
	f := logrus.Fields{
		"functionName":           "v2.repositories.service.AddGithubRepositories",
		utils.XREQUESTID:         ctx.Value(utils.XREQUESTID),
		"projectSFID":            projectSFID,
		"claGroupID":             utils.StringValue(input.ClaGroupID),
		"githubOrganizationName": utils.StringValue(input.GithubOrganizationName),
		"repositoryGitHubID":     input.RepositoryGithubID,
		"repositoryGithubIds":    input.RepositoryGithubIds,
	}

	log.WithFields(f).Debugf("loading project by SFID: %s", projectSFID)
	psc := v2ProjectService.GetClient()
	project, err := psc.GetProject(projectSFID)
	if err != nil {
		log.WithFields(f).WithError(err).Warn("unable to load projectSFID from the platform project service")
		return nil, err
	}

	var parentProjectSFID string
	if utils.StringValue(project.Parent) == "" || (project.Foundation != nil &&
		(project.Foundation.Name == utils.TheLinuxFoundation || project.Foundation.Name == utils.LFProjectsLLC)) {
		parentProjectSFID = projectSFID
	} else {
		parentProjectSFID = utils.StringValue(project.Parent)
	}

	allMappings, err := s.projectsClaGroupsRepo.GetProjectsIdsForClaGroup(ctx, aws.StringValue(input.ClaGroupID))
	if err != nil {
		log.WithFields(f).WithError(err).Warn("unable to get project IDs for CLA Group")
		return nil, err
	}
	var valid bool
	for _, cgm := range allMappings {
		if cgm.ProjectSFID == projectSFID || cgm.FoundationSFID == projectSFID {
			valid = true
			break
		}
	}
	if !valid {
		return nil, fmt.Errorf("provided cla group id %s is not linked to project sfid %s", utils.StringValue(input.ClaGroupID), projectSFID)
	}

	org, err := s.ghOrgRepo.GetGithubOrganizationByName(ctx, utils.StringValue(input.GithubOrganizationName))
	if err != nil {
		log.WithFields(f).WithError(err).Warn("unable to get organization by name")
		return nil, err
	}
	if len(org.List) == 0 {
		return nil, errors.New("github app not installed on github organization")
	}

	// Updated to process a list of repository IDs - take the list (may be empty) and add the single repository GH ID if it was set
	repositoryIDList := input.RepositoryGithubIds
	if input.RepositoryGithubID != "" {
		repositoryIDList = append(repositoryIDList, input.RepositoryGithubID)
	}

	// Remove any silly duplicates that may come
	repositoryIDList = utils.RemoveDuplicates(repositoryIDList)

	var response []*v1Models.GithubRepository

	// For each repository ID provided...
	// If this is slow, may want to optimize by making separate go routines for each item in the list
	for _, repoID := range repositoryIDList {
		// Convert the string value to an integer
		repoGithubID, err := strconv.ParseInt(repoID, 10, 64)
		if err != nil {
			log.WithFields(f).WithError(err).Warnf("unable to convert repository github ID %s to an integer - invalid value", repoID)
			return nil, err
		}

		log.WithFields(f).Debugf("loading GitHub repository by external id: %d", repoGithubID)
		ghRepo, err := github.GetRepositoryByExternalID(ctx, org.List[0].OrganizationInstallationID, repoGithubID)
		if err != nil {
			log.WithFields(f).WithError(err).Warnf("unable to load repository by external ID: %d", repoGithubID)
			return nil, err
		}
		f["repositoryName"] = ghRepo.FullName
		f["repositoryURL"] = ghRepo.URL
		f["repositoryGitHubID"] = repoGithubID
		log.WithFields(f).Debugf("loaded GitHub repository by external id: %d - url: %s", repoGithubID, utils.StringValue(ghRepo.URL))

		// Check if this repository exists in our database
		log.WithFields(f).Debugf("checking if GitHub repository by name: %s exists...", utils.StringValue(ghRepo.FullName))
		existingRepositoryModel, lookupErr := s.GetRepositoryByName(ctx, utils.StringValue(ghRepo.FullName))
		if lookupErr != nil {
			// If we have the repository not found error - this is ok - we are expecting this
			if notFoundErr, ok := lookupErr.(*utils.GitHubRepositoryNotFound); ok {
				log.WithFields(f).WithError(notFoundErr).Debugf("GitHub repository lookup didn't find a match for existing repository name: %s - ok to create", utils.StringValue(ghRepo.FullName))
			} else {
				// Some other error - not good...
				log.WithFields(f).WithError(lookupErr).Warnf("GitHub repository lookup failed for repository name: %s", utils.StringValue(ghRepo.FullName))
				return nil, lookupErr
			}
		}

		// We already have an existing repository model with the same name
		if existingRepositoryModel != nil {
			if !existingRepositoryModel.Enabled {
				msg := fmt.Sprintf("Github repository: %s previously disabled - will re-enabled... ", utils.StringValue(ghRepo.FullName))
				log.WithFields(f).Debug(msg)
				enabled := true

				_, now := utils.CurrentTime()

				log.WithFields(f).Debugf("Updating GitHub repository - setting enabled: true, OrgName: %s, CLA Group ID: %s",
					utils.StringValue(input.GithubOrganizationName), utils.StringValue(input.ClaGroupID))
				v1Input := &v1Models.GithubRepositoryInput{
					Enabled:                    &enabled,
					RepositoryOrganizationName: input.GithubOrganizationName,
					RepositoryProjectID:        input.ClaGroupID,
					Note:                       fmt.Sprintf("re-enabling repository on %s.", now),
				}

				// Update Repo details in case of any changes
				updatedRepository, updateErr := s.repo.UpdateGithubRepository(ctx, existingRepositoryModel.RepositoryID, v1Input)
				if updateErr != nil {
					log.WithFields(f).WithError(updateErr).Warnf("unable to update GitHub repository with name: %s, id: %s, using input: %+v", utils.StringValue(ghRepo.FullName), existingRepositoryModel.RepositoryID, v1Input)
					return nil, updateErr
				}

				// Append the results to our response model
				response = append(response, updatedRepository)
			} else {
				log.WithFields(f).Warnf("GitHub repository already exists with repository name: %s and is already enabled - skipping update", utils.StringValue(ghRepo.FullName))
				continue
			}
		} else {
			// No record exists...
			log.WithFields(f).Debug("no existing GitHub repository configured - creating...")
			in := &v1Models.GithubRepositoryInput{
				RepositoryExternalID:       &repoID, // nolint
				RepositoryName:             ghRepo.FullName,
				RepositoryOrganizationName: input.GithubOrganizationName,
				RepositoryProjectID:        input.ClaGroupID,
				RepositoryType:             aws.String("github"),
				RepositoryURL:              ghRepo.HTMLURL,
			}

			addedModel, addErr := s.repo.AddGithubRepository(ctx, parentProjectSFID, projectSFID, in)
			if addErr != nil {
				log.WithFields(f).WithError(addErr).Warnf("unable to add github repository: %s for project: %s", *ghRepo.FullName, projectSFID)
				return nil, addErr
			}

			// Append the results to our response model
			response = append(response, addedModel)
		}
	}

	return response, nil
}

func (s *service) EnableRepository(ctx context.Context, repositoryID string) error {
	return s.repo.EnableRepository(ctx, repositoryID)
}

func (s *service) DisableRepository(ctx context.Context, repositoryID string) error {
	return s.repo.DisableRepository(ctx, repositoryID)
}

func (s *service) ListProjectRepositories(ctx context.Context, projectSFID string) (*v1Models.ListGithubRepositories, error) {
	f := logrus.Fields{
		"functionName":   "v2.repositories.service.ListProjectRepositories",
		utils.XREQUESTID: ctx.Value(utils.XREQUESTID),
		"projectSFID":    projectSFID,
	}

	log.WithFields(f).Debug("querying project service for project...")
	psc := v2ProjectService.GetClient()
	projectModel, err := psc.GetProject(projectSFID)
	if err != nil {
		log.WithFields(f).WithError(err).Warn("unable to lookup project by id in the project service")
		return nil, err
	}
	if projectModel == nil {
		log.WithFields(f).Warn("unable to lookup project by id in the project service - no record found")
		return nil, err
	}
	f["projectName"] = projectModel.Name
	if utils.StringValue(projectModel.Parent) != "" {
		f["projectParentSFID"] = projectModel.Parent
	}
	log.WithFields(f).Debug("loaded project from the project service")
	enabled := true
	return s.repo.ListProjectRepositories(ctx, projectSFID, &enabled)

	//// Lookup orgs via projectSFID
	//log.WithFields(f).Debug("querying EasyCLA for organizations by project id...")
	//var githubOrgList *v1Models.GithubOrganizations
	//githubOrgList, err = s.ghOrgRepo.GetGithubOrganizations(ctx, projectSFID)
	//if err != nil {
	//	log.WithFields(f).WithError(err).Warn("unable to lookup project by id in the github organization table")
	//	if projectModel.Parent != "" {
	//		log.WithFields(f).Debugf("querying for organizations by parent project id: %s...", projectModel.Parent)
	//		var ghOrgErr error
	//		githubOrgList, ghOrgErr = s.ghOrgRepo.GetGithubOrganizations(ctx, projectModel.Parent)
	//		if ghOrgErr != nil {
	//			log.WithFields(f).WithError(ghOrgErr).Warn("unable to lookup project by parent id in the github organization table")
	//			return nil, ghOrgErr
	//		}
	//	}
	//}
	//
	//// Our response - empty to start with
	//response := &v1Models.ListGithubRepositories{
	//	List: []*v1Models.GithubRepository{},
	//}
	//
	//if githubOrgList == nil {
	//	log.WithFields(f).Warn("unable to lookup project by id - no records found")
	//	return response, err
	//}
	//log.WithFields(f).Debugf("loaded %d EasyCLA GitHub organizations for project", len(githubOrgList.List))
	//
	//// For each of the organizations we have in our database for this project...
	//for _, gitHubOrg := range githubOrgList.List {
	//	// Query GitHub for the list of public repositories...
	//	log.WithFields(f).Debugf("querying github by organization: %s", gitHubOrg.OrganizationName)
	//	ghRepoList, getRepoErr := github.GetRepositories(ctx, gitHubOrg.OrganizationName)
	//	if getRepoErr != nil {
	//		log.WithFields(f).WithError(getRepoErr).Warn("unable to lookup github organization details")
	//		return response, getRepoErr
	//	}
	//
	//	// Add to our response model...use default values (enabled = false)
	//	log.WithFields(f).Debugf("found %d github repositories for organization: %s", len(ghRepoList), gitHubOrg.OrganizationName)
	//	for _, ghRepo := range ghRepoList {
	//		response.List = append(response.List, &v1Models.GithubRepository{
	//			Enabled:                    false,
	//			ProjectSFID:                projectSFID,
	//			RepositoryExternalID:       projectSFID,
	//			RepositoryName:             utils.StringValue(ghRepo.Name),
	//			RepositoryOrganizationName: gitHubOrg.OrganizationName,
	//			RepositoryType:             "github",
	//			RepositoryURL:              utils.StringValue(ghRepo.URL),
	//			Version:                    "v1",
	//		})
	//	}
	//}
	//
	//// Now, query our DB....
	//listOurGitHubRepos, err := s.repo.ListProjectRepositories(ctx, "", projectSFID, true)
	//if err != nil {
	//	log.WithFields(f).WithError(err).Warn("unable to lookup repository records by id in our repositories table ")
	//	return response, err
	//}
	//if listOurGitHubRepos == nil || len(listOurGitHubRepos.List) == 0 {
	//	log.WithFields(f).Warn("unable to lookup repository records by id in our repositories table ")
	//	return response, err
	//}
	//
	//// For each repo that we have...
	//for _, ourGitHubRepo := range listOurGitHubRepos.List {
	//	// Inefficient, but ok if the number of repos is relatively small
	//	for _, r := range response.List {
	//		// Copy over the additional details
	//		if ourGitHubRepo.RepositoryName == r.RepositoryName {
	//			r.RepositoryID = ourGitHubRepo.RepositoryID
	//			r.Enabled = ourGitHubRepo.Enabled
	//			r.DateCreated = ourGitHubRepo.DateCreated
	//			r.DateModified = ourGitHubRepo.DateModified
	//			r.Note = ourGitHubRepo.Note
	//			r.Version = ourGitHubRepo.Version
	//			break
	//		}
	//	}
	//}
	//
	//return response, nil
}

func (s *service) GetRepository(ctx context.Context, repositoryID string) (*v1Models.GithubRepository, error) {
	return s.repo.GetRepository(ctx, repositoryID)
}

func (s *service) GetRepositoryByName(ctx context.Context, repositoryName string) (*v1Models.GithubRepository, error) {
	return s.repo.GetRepositoryByName(ctx, repositoryName)
}

func (s *service) GetProtectedBranch(ctx context.Context, projectSFID, repositoryID, branchName string) (*v2Models.GithubRepositoryBranchProtection, error) {
	f := logrus.Fields{
		"functionName":   "v2.repositories.service.GetProtectedBranch",
		utils.XREQUESTID: ctx.Value(utils.XREQUESTID),
		"projectSFID":    projectSFID,
		"repositoryID":   repositoryID,
		"branchName":     branchName,
	}

	githubRepository, err := s.getGithubRepo(ctx, projectSFID, repositoryID)
	if err != nil {
		log.WithFields(f).WithError(err).Warnf("fetching repository %s, failed, error: %v", repositoryID, err)
		return nil, err
	}

	githubOrgName := githubRepository.RepositoryOrganizationName
	githubRepoName := githubRepository.RepositoryName
	githubRepoName = branch_protection.CleanGithubRepoName(githubRepoName)

	branchProtectionRepository, err := s.getBranchProtectionRepositoryForOrgName(ctx, githubOrgName)
	if err != nil {
		return nil, err
	}

	owner, err := s.getGithubOwner(ctx, branchProtectionRepository, githubOrgName, githubRepoName)
	if err != nil {
		return nil, err
	}

	result := &v2Models.GithubRepositoryBranchProtection{
		BranchName: &branchName,
	}

	branchProtection, err := branchProtectionRepository.GetProtectedBranch(ctx, owner, githubRepoName, branchName)
	if err != nil {
		if errors.Is(err, branch_protection.ErrBranchNotProtected) {
			return result, nil
		}
		log.WithFields(f).WithError(err).Warnf("getting the github protected branch for owner : %s, repo : %s and branch : %s failed : %v", owner, githubRepoName, branchName, err)
		return nil, err
	}

	result.ProtectionEnabled = true
	if branch_protection.IsEnforceAdminEnabled(branchProtection) {
		result.EnforceAdmin = true
	}

	requiredChecks := requiredBranchProtectionChecks
	requiredChecksResult := s.getRequiredProtectedBranchCheckStatus(branchProtection, requiredChecks)
	result.StatusChecks = requiredChecksResult

	return result, nil
}

func (s *service) UpdateProtectedBranch(ctx context.Context, projectSFID, repositoryID string, input *v2Models.GithubRepositoryBranchProtectionInput) (*v2Models.GithubRepositoryBranchProtection, error) {
	f := logrus.Fields{
		"functionName":   "v2.repositories.service.UpdateProtectedBranch",
		utils.XREQUESTID: ctx.Value(utils.XREQUESTID),
		"projectSFID":    projectSFID,
		"repositoryID":   repositoryID,
		"enforceAdmin":   aws.BoolValue(input.EnforceAdmin),
	}

	githubRepository, err := s.getGithubRepo(ctx, projectSFID, repositoryID)
	if err != nil {
		log.WithFields(f).WithError(err).Warnf("fetching repository %s, failed", repositoryID)
		return nil, err
	}

	githubOrgName := githubRepository.RepositoryOrganizationName
	githubRepoName := githubRepository.RepositoryName
	githubRepoName = branch_protection.CleanGithubRepoName(githubRepoName)

	branchProtectionRepository, err := s.getBranchProtectionRepositoryForOrgName(ctx, githubOrgName)
	if err != nil {
		log.WithFields(f).WithError(err).Warn("problem locating github client for organization name")
		return nil, err
	}

	branchName := input.BranchName
	if branchName == "" {
		branchName = branch_protection.DefaultBranchName
	}

	owner, err := s.getGithubOwner(ctx, branchProtectionRepository, githubOrgName, githubRepoName)
	if err != nil {
		log.WithFields(f).WithError(err).Warn("problem locating github owner branch name")
		return nil, err
	}
	f["owner"] = owner
	f["branchName"] = input.BranchName

	var requiredChecks []string
	var disabledChecks []string
	if input.StatusChecks != nil {
		for _, inputCheck := range input.StatusChecks {
			// we want to make sure we only mutate checks related to lf
			var found bool
			for _, rc := range requiredBranchProtectionChecks {
				if rc == *inputCheck.Name {
					found = true
					break
				}
			}

			// just ignore that check if it's something not in our options
			if !found {
				log.WithFields(f).Warnf("invalid branch protection option was found : %s", *inputCheck.Name)
				return nil, ErrInvalidBranchProtectionName
			}

			if !*inputCheck.Enabled {
				disabledChecks = append(disabledChecks, *inputCheck.Name)
				continue
			}
			requiredChecks = append(requiredChecks, *inputCheck.Name)
		}
	}

	log.WithFields(f).Debugf("enabling branch protection on repository...")
	err = branchProtectionRepository.EnableBranchProtection(ctx, owner, githubRepoName, branchName, *input.EnforceAdmin, requiredChecks, disabledChecks)
	if err != nil {
		log.WithFields(f).WithError(err).Warn("problem enabling github branch protection")
		return nil, err
	}

	return s.GetProtectedBranch(ctx, projectSFID, repositoryID, branchName)
}

func (s *service) getGithubRepo(ctx context.Context, projectSFID, repositoryID string) (*v1Models.GithubRepository, error) {
	f := logrus.Fields{
		"functionName":   "v2.repositories.service.getGitHubRepo",
		utils.XREQUESTID: ctx.Value(utils.XREQUESTID),
		"projectSFID":    projectSFID,
		"repositoryID":   repositoryID,
	}

	psc := v2ProjectService.GetClient()
	_, err := psc.GetProject(projectSFID)
	if err != nil {
		return nil, err
	}
	githubRepository, err := s.GetRepository(ctx, repositoryID)
	if err != nil {
		log.WithFields(f).Warnf("fetching repository failed : %s : %v", repositoryID, err)
		return nil, err
	}

	// check if project and repo are actually associated
	if githubRepository.ProjectSFID != projectSFID {
		msg := fmt.Sprintf("github repository %s doesn't belong to project : %s", repositoryID, projectSFID)
		log.WithFields(f).Warn(msg)
		return nil, errors.New(msg)
	}

	return githubRepository, nil
}

func (s *service) getBranchProtectionRepositoryForOrgName(ctx context.Context, githubOrgName string) (*branch_protection.BranchProtectionRepository, error) {
	f := logrus.Fields{
		"functionName":   "v2.repositories.service.getGitHubClientForOrgName",
		utils.XREQUESTID: ctx.Value(utils.XREQUESTID),
		"githubOrgName":  githubOrgName,
	}

	githubOrg, err := s.ghOrgRepo.GetGithubOrganization(ctx, githubOrgName)
	if err != nil {
		log.WithFields(f).Warnf("fetching githubOrg %s failed, error: %v", githubOrgName, err)
		return nil, err
	}

	branchProtectionRepo, err := branch_protection.NewBranchProtectionRepository(githubOrg.OrganizationInstallationID, branch_protection.EnableNonBlockingLimiter())
	if err != nil {
		return nil, err
	}
	return branchProtectionRepo, nil
}

func (s *service) getGithubOwner(ctx context.Context, branchProtectionRepository *branch_protection.BranchProtectionRepository, githubOrgName, githubRepoName string) (string, error) {
	owner, err := branchProtectionRepository.GetOwnerName(ctx, githubOrgName, githubRepoName)
	if err != nil {
		log.Warnf("getting the owner name for org : %s and repo : %s failed : %v", githubOrgName, githubRepoName, err)
		return "", err
	}

	if owner == "" {
		log.Warnf("GitHub returned empty owner name for org : %s and repo : %s", githubOrgName, githubRepoName)
		return "", fmt.Errorf("empty owner name")
	}
	return owner, nil
}

// getRequiredProtectedBranchCheckStatus
func (s *service) getRequiredProtectedBranchCheckStatus(branchProtectionRule *branch_protection.BranchProtectionRule, requiredChecks []string) []*v2Models.GithubRepositoryBranchProtectionStatusChecks {
	f := logrus.Fields{
		"functionName": "v2.repositories.service.getRequiredProtectedBranchCheckStatus",
	}

	log.WithFields(f).Debug("querying GitHub for status checks...")
	var result []*v2Models.GithubRepositoryBranchProtectionStatusChecks
	resultMap := map[string]bool{}
	for _, rc := range requiredChecks {
		result = append(result, &v2Models.GithubRepositoryBranchProtectionStatusChecks{
			Name:    swag.String(rc),
			Enabled: swag.Bool(false),
		})
		resultMap[rc] = true
	}
	if len(branchProtectionRule.RequiredStatusCheckContexts) == 0 {
		return result
	}

	for _, existingCheck := range branchProtectionRule.RequiredStatusCheckContexts {
		if !resultMap[existingCheck] {
			continue
		}

		// we mark it as enabled in this case
		for i := range result {
			if *result[i].Name == existingCheck {
				result[i].Enabled = swag.Bool(true)
			}
		}
	}
	return result
}

func (s *service) DisableCLAGroupRepositories(ctx context.Context, claGroupID string) error {
	f := logrus.Fields{
		"functionName":   "v2.repositories.service.DisableCLAGroupRepositories",
		utils.XREQUESTID: ctx.Value(utils.XREQUESTID),
		"claGroupID":     claGroupID,
	}

	var deleteErr error
	ghOrgs, err := s.repo.GetCLAGroupRepositoriesGroupByOrgs(ctx, claGroupID, true)
	if err != nil {
		return err
	}
	if len(ghOrgs) > 0 {
		log.WithFields(f).Debugf("Deleting repositories for cla-group :%s", claGroupID)
		for _, ghOrg := range ghOrgs {
			for _, item := range ghOrg.List {
				deleteErr = s.repo.DisableRepository(ctx, item.RepositoryID)
				if deleteErr != nil {
					log.WithFields(f).Warnf("Unable to remove repository: %s for project :%s error :%v", item.RepositoryID, claGroupID, deleteErr)
				}
			}
		}
	}
	return nil
}

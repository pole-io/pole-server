/**
 * Tencent is pleased to support the open source community by making Polaris available.
 *
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *
 * Licensed under the BSD 3-Clause License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * https://opensource.org/licenses/BSD-3-Clause
 *
 * Unless required by applicable law or agreed to in writing, software distributed
 * under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR
 * CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package skill

import (
	"context"
	"errors"
	"fmt"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"

	aiTypes "github.com/pole-io/pole-server/apis/pkg/types/ai"
	"github.com/pole-io/pole-server/apis/store"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/pkg/common/log"
)

var (
	server       SkillServer
	originServer = &Server{}
)

// SkillServer is the interface for skill server
type SkillServer interface {
	// Skill CRUD operations
	CreateSkill(ctx context.Context, skill *aiTypes.Skill) error
	UpdateSkill(ctx context.Context, skill *aiTypes.Skill) error
	DeleteSkill(ctx context.Context, id string) error
	GetSkill(ctx context.Context, id string) (*aiTypes.Skill, error)
	GetSkillByName(ctx context.Context, name, namespace string) (*aiTypes.Skill, error)

	// Skill batch operations
	CreateSkills(ctx context.Context, skills []*aiTypes.Skill) *apimodel.BatchWriteResponse
	UpdateSkills(ctx context.Context, skills []*aiTypes.Skill) *apimodel.BatchWriteResponse
	DeleteSkills(ctx context.Context, skills []*aiTypes.Skill) *apimodel.BatchWriteResponse
	GetSkills(ctx context.Context, query map[string]string) *apimodel.BatchQueryResponse
	GetAllSkills(ctx context.Context, query map[string]string) *apimodel.BatchQueryResponse
	GetSkillsCount(ctx context.Context) *apimodel.BatchQueryResponse

	// SkillGroup CRUD operations
	CreateSkillGroup(ctx context.Context, group *aiTypes.SkillGroup) error
	UpdateSkillGroup(ctx context.Context, group *aiTypes.SkillGroup) error
	DeleteSkillGroup(ctx context.Context, namespace, name string) error
	GetSkillGroup(ctx context.Context, namespace, name string) (*aiTypes.SkillGroup, error)

	// SkillGroup batch operations
	CreateSkillGroups(ctx context.Context, groups []*aiTypes.SkillGroup) *apimodel.BatchWriteResponse
	UpdateSkillGroups(ctx context.Context, groups []*aiTypes.SkillGroup) *apimodel.BatchWriteResponse
	DeleteSkillGroups(ctx context.Context, groups []*aiTypes.SkillGroup) *apimodel.BatchWriteResponse
	GetSkillGroups(ctx context.Context, query map[string]string) *apimodel.BatchQueryResponse

	// SkillVersion operations
	CreateSkillVersion(ctx context.Context, version *aiTypes.SkillVersion) error
	ActivateSkillVersion(ctx context.Context, versionID string) error

	// SkillVersion batch operations
	CreateSkillVersions(ctx context.Context, versions []*aiTypes.SkillVersion) *apimodel.BatchWriteResponse
	DeleteSkillVersions(ctx context.Context, ids []string) *apimodel.BatchWriteResponse
	GetSkillVersions(ctx context.Context, query map[string]string) *apimodel.BatchQueryResponse

	// SkillSubscription operations
	CreateSkillSubscription(ctx context.Context, sub *aiTypes.SkillSubscription) error
	DeleteSkillSubscription(ctx context.Context, id string) error
	GetSkillSubscriptionsBySkill(ctx context.Context, skillName, namespace string) ([]*aiTypes.SkillSubscription, error)
	GetSkillSubscriptionsByClient(ctx context.Context, clientID string) ([]*aiTypes.SkillSubscription, error)

	// SkillSubscription batch operations
	CreateSkillSubscriptions(ctx context.Context, subs []*aiTypes.SkillSubscription) *apimodel.BatchWriteResponse
	DeleteSkillSubscriptions(ctx context.Context, ids []string) *apimodel.BatchWriteResponse
	GetSkillSubscriptions(ctx context.Context, query map[string]string) *apimodel.BatchQueryResponse
}

// Server is the skill operation server
type Server struct {
	storage  store.AIStore
	initialized bool
}

// WithStorage sets the storage for skill server
func WithStorage(storage store.AIStore) ServerOption {
	return func(s *Server) {
		s.storage = storage
	}
}

// ServerOption is the option for skill server
type ServerOption func(*Server)

// Initialize initializes the skill server
func Initialize(s store.AIStore) error {
	originServer.storage = s
	originServer.initialized = true
	server = originServer
	return nil
}

// GetServer gets the skill server instance
func GetServer() (SkillServer, error) {
	if !originServer.initialized {
		return nil, errors.New("skill server has not been initialized")
	}
	return server, nil
}

// CreateSkill creates a skill
func (s *Server) CreateSkill(ctx context.Context, skill *aiTypes.Skill) error {
	if s.storage == nil {
		return errors.New("storage is not initialized")
	}

	log.Infof("[Skill] create skill, namespace: %s, name: %s", skill.Namespace, skill.Name)
	return s.storage.CreateSkill(skill)
}

// UpdateSkill updates a skill
func (s *Server) UpdateSkill(ctx context.Context, skill *aiTypes.Skill) error {
	if s.storage == nil {
		return errors.New("storage is not initialized")
	}

	log.Infof("[Skill] update skill, namespace: %s, name: %s", skill.Namespace, skill.Name)
	return s.storage.UpdateSkill(skill)
}

// DeleteSkill deletes a skill
func (s *Server) DeleteSkill(ctx context.Context, id string) error {
	if s.storage == nil {
		return errors.New("storage is not initialized")
	}

	skill, err := s.storage.GetSkill(id)
	if err != nil {
		return err
	}
	if skill == nil {
		return fmt.Errorf("skill %s not found", id)
	}

	log.Infof("[Skill] delete skill, namespace: %s, name: %s", skill.Namespace, skill.Name)
	return s.storage.DeleteSkill(id)
}

// GetSkill gets a skill by ID
func (s *Server) GetSkill(ctx context.Context, id string) (*aiTypes.Skill, error) {
	if s.storage == nil {
		return nil, errors.New("storage is not initialized")
	}

	return s.storage.GetSkill(id)
}

// GetSkillByName gets a skill by name and namespace
func (s *Server) GetSkillByName(ctx context.Context, name, namespace string) (*aiTypes.Skill, error) {
	if s.storage == nil {
		return nil, errors.New("storage is not initialized")
	}

	return s.storage.GetSkillByName(name, namespace)
}

// CreateSkillGroup creates a skill group
func (s *Server) CreateSkillGroup(ctx context.Context, group *aiTypes.SkillGroup) error {
	if s.storage == nil {
		return errors.New("storage is not initialized")
	}

	log.Infof("[SkillGroup] create skill group, namespace: %s, name: %s", group.Namespace, group.Name)
	_, err := s.storage.CreateSkillGroup(group)
	return err
}

// UpdateSkillGroup updates a skill group
func (s *Server) UpdateSkillGroup(ctx context.Context, group *aiTypes.SkillGroup) error {
	if s.storage == nil {
		return errors.New("storage is not initialized")
	}

	log.Infof("[SkillGroup] update skill group, namespace: %s, name: %s", group.Namespace, group.Name)
	return s.storage.UpdateSkillGroup(group)
}

// DeleteSkillGroup deletes a skill group
func (s *Server) DeleteSkillGroup(ctx context.Context, namespace, name string) error {
	if s.storage == nil {
		return errors.New("storage is not initialized")
	}

	log.Infof("[SkillGroup] delete skill group, namespace: %s, name: %s", namespace, name)
	return s.storage.DeleteSkillGroup(namespace, name)
}

// GetSkillGroup gets a skill group by namespace and name
func (s *Server) GetSkillGroup(ctx context.Context, namespace, name string) (*aiTypes.SkillGroup, error) {
	if s.storage == nil {
		return nil, errors.New("storage is not initialized")
	}

	return s.storage.GetSkillGroup(namespace, name)
}

// CreateSkillVersion creates a skill version
func (s *Server) CreateSkillVersion(ctx context.Context, version *aiTypes.SkillVersion) error {
	if s.storage == nil {
		return errors.New("storage is not initialized")
	}

	log.Infof("[SkillVersion] create skill version, namespace: %s, skill: %s",
		version.Namespace, version.SkillName)
	return s.storage.CreateSkillVersion(version)
}

// ActivateSkillVersion activates a skill version
func (s *Server) ActivateSkillVersion(ctx context.Context, versionID string) error {
	if s.storage == nil {
		return errors.New("storage is not initialized")
	}

	version, err := s.storage.GetSkillVersion(versionID)
	if err != nil {
		return err
	}
	if version == nil {
		return fmt.Errorf("skill version %s not found", versionID)
	}

	log.Infof("[SkillVersion] activate skill version, namespace: %s, skill: %s, version: %d",
		version.Namespace, version.SkillName, version.Version)

	// First deactivate all other versions
	versions, err := s.storage.GetSkillVersionsBySkillID(version.SkillID)
	if err != nil {
		return err
	}

	for _, v := range versions {
		if v.ID != versionID && v.Active {
			if err := s.storage.InactiveSkillVersion(v); err != nil {
				return err
			}
		}
	}

	// Then activate the new version
	return s.storage.ActiveSkillVersion(version)
}

// CreateSkillSubscription creates a skill subscription
func (s *Server) CreateSkillSubscription(ctx context.Context, sub *aiTypes.SkillSubscription) error {
	if s.storage == nil {
		return errors.New("storage is not initialized")
	}

	log.Infof("[SkillSubscription] create skill subscription, namespace: %s, skill: %s, client: %s",
		sub.Namespace, sub.SkillName, sub.ClientID)
	return s.storage.CreateSkillSubscription(sub)
}

// DeleteSkillSubscription deletes a skill subscription
func (s *Server) DeleteSkillSubscription(ctx context.Context, id string) error {
	if s.storage == nil {
		return errors.New("storage is not initialized")
	}

	log.Infof("[SkillSubscription] delete skill subscription, id: %s", id)
	return s.storage.DeleteSkillSubscription(id)
}

// GetSkillSubscriptionsBySkill gets skill subscriptions by skill
func (s *Server) GetSkillSubscriptionsBySkill(ctx context.Context, skillName, namespace string) ([]*aiTypes.SkillSubscription, error) {
	if s.storage == nil {
		return nil, errors.New("storage is not initialized")
	}

	return s.storage.GetSkillSubscriptionsBySkill(skillName, namespace)
}

// GetSkillSubscriptionsByClient gets skill subscriptions by client
func (s *Server) GetSkillSubscriptionsByClient(ctx context.Context, clientID string) ([]*aiTypes.SkillSubscription, error) {
	if s.storage == nil {
		return nil, errors.New("storage is not initialized")
	}

	return s.storage.GetSkillSubscriptionByClient(clientID)
}

// ===== Batch Operations =====

// CreateSkills creates multiple skills
func (s *Server) CreateSkills(ctx context.Context, skills []*aiTypes.Skill) *apimodel.BatchWriteResponse {
	resp := api.NewBatchWriteResponse(apimodel.Code_ExecuteSuccess)
	for _, skill := range skills {
		if err := s.CreateSkill(ctx, skill); err != nil {
			log.Errorf("[Skill] create skill failed, namespace: %s, name: %s, err: %v",
				skill.Namespace, skill.Name, err)
			api.Collect(resp, api.NewResponseWithMsg(apimodel.Code_ExecuteException, err.Error()))
		} else {
			api.Collect(resp, api.NewResponse(apimodel.Code_ExecuteSuccess))
		}
	}
	return api.FormatBatchWriteResponse(resp)
}

// UpdateSkills updates multiple skills
func (s *Server) UpdateSkills(ctx context.Context, skills []*aiTypes.Skill) *apimodel.BatchWriteResponse {
	resp := api.NewBatchWriteResponse(apimodel.Code_ExecuteSuccess)
	for _, skill := range skills {
		if err := s.UpdateSkill(ctx, skill); err != nil {
			log.Errorf("[Skill] update skill failed, namespace: %s, name: %s, err: %v",
				skill.Namespace, skill.Name, err)
			api.Collect(resp, api.NewResponseWithMsg(apimodel.Code_ExecuteException, err.Error()))
		} else {
			api.Collect(resp, api.NewResponse(apimodel.Code_ExecuteSuccess))
		}
	}
	return api.FormatBatchWriteResponse(resp)
}

// DeleteSkills deletes multiple skills
func (s *Server) DeleteSkills(ctx context.Context, skills []*aiTypes.Skill) *apimodel.BatchWriteResponse {
	resp := api.NewBatchWriteResponse(apimodel.Code_ExecuteSuccess)
	for _, skill := range skills {
		if err := s.DeleteSkill(ctx, skill.ID); err != nil {
			log.Errorf("[Skill] delete skill failed, id: %s, err: %v", skill.ID, err)
			api.Collect(resp, api.NewResponseWithMsg(apimodel.Code_ExecuteException, err.Error()))
		} else {
			api.Collect(resp, api.NewResponse(apimodel.Code_ExecuteSuccess))
		}
	}
	return api.FormatBatchWriteResponse(resp)
}

// GetSkills queries skills with filters
func (s *Server) GetSkills(ctx context.Context, query map[string]string) *apimodel.BatchQueryResponse {
	if s.storage == nil {
		return api.NewBatchQueryResponse(apimodel.Code_StoreLayerException)
	}

	offset, limit := parseOffsetLimit(query)
	count, skills, err := s.storage.QuerySkills(query, offset, limit)
	if err != nil {
		log.Errorf("[Skill] query skills failed, err: %v", err)
		return api.NewBatchQueryResponse(apimodel.Code_StoreLayerException)
	}

	resp := api.NewBatchQueryResponse(apimodel.Code_ExecuteSuccess)
	resp.Amount = count
	resp.Size = uint32(len(skills))
	return resp
}

// GetAllSkills gets all skills
func (s *Server) GetAllSkills(ctx context.Context, query map[string]string) *apimodel.BatchQueryResponse {
	if s.storage == nil {
		return api.NewBatchQueryResponse(apimodel.Code_StoreLayerException)
	}

	// Use large limit to get all
	count, skills, err := s.storage.QuerySkills(query, 0, 10000)
	if err != nil {
		log.Errorf("[Skill] get all skills failed, err: %v", err)
		return api.NewBatchQueryResponse(apimodel.Code_StoreLayerException)
	}

	resp := api.NewBatchQueryResponse(apimodel.Code_ExecuteSuccess)
	resp.Amount = count
	resp.Size = uint32(len(skills))
	return resp
}

// GetSkillsCount gets the total count of skills
func (s *Server) GetSkillsCount(ctx context.Context) *apimodel.BatchQueryResponse {
	if s.storage == nil {
		return api.NewBatchQueryResponse(apimodel.Code_StoreLayerException)
	}

	count, _, err := s.storage.QuerySkills(nil, 0, 0)
	if err != nil {
		log.Errorf("[Skill] get skills count failed, err: %v", err)
		return api.NewBatchQueryResponse(apimodel.Code_StoreLayerException)
	}

	resp := api.NewBatchQueryResponse(apimodel.Code_ExecuteSuccess)
	resp.Amount = count
	return resp
}

// CreateSkillGroups creates multiple skill groups
func (s *Server) CreateSkillGroups(ctx context.Context, groups []*aiTypes.SkillGroup) *apimodel.BatchWriteResponse {
	resp := api.NewBatchWriteResponse(apimodel.Code_ExecuteSuccess)
	for _, group := range groups {
		if err := s.CreateSkillGroup(ctx, group); err != nil {
			log.Errorf("[SkillGroup] create skill group failed, namespace: %s, name: %s, err: %v",
				group.Namespace, group.Name, err)
			api.Collect(resp, api.NewResponseWithMsg(apimodel.Code_ExecuteException, err.Error()))
		} else {
			api.Collect(resp, api.NewResponse(apimodel.Code_ExecuteSuccess))
		}
	}
	return api.FormatBatchWriteResponse(resp)
}

// UpdateSkillGroups updates multiple skill groups
func (s *Server) UpdateSkillGroups(ctx context.Context, groups []*aiTypes.SkillGroup) *apimodel.BatchWriteResponse {
	resp := api.NewBatchWriteResponse(apimodel.Code_ExecuteSuccess)
	for _, group := range groups {
		if err := s.UpdateSkillGroup(ctx, group); err != nil {
			log.Errorf("[SkillGroup] update skill group failed, namespace: %s, name: %s, err: %v",
				group.Namespace, group.Name, err)
			api.Collect(resp, api.NewResponseWithMsg(apimodel.Code_ExecuteException, err.Error()))
		} else {
			api.Collect(resp, api.NewResponse(apimodel.Code_ExecuteSuccess))
		}
	}
	return api.FormatBatchWriteResponse(resp)
}

// DeleteSkillGroups deletes multiple skill groups
func (s *Server) DeleteSkillGroups(ctx context.Context, groups []*aiTypes.SkillGroup) *apimodel.BatchWriteResponse {
	resp := api.NewBatchWriteResponse(apimodel.Code_ExecuteSuccess)
	for _, group := range groups {
		if err := s.DeleteSkillGroup(ctx, group.Namespace, group.Name); err != nil {
			log.Errorf("[SkillGroup] delete skill group failed, namespace: %s, name: %s, err: %v",
				group.Namespace, group.Name, err)
			api.Collect(resp, api.NewResponseWithMsg(apimodel.Code_ExecuteException, err.Error()))
		} else {
			api.Collect(resp, api.NewResponse(apimodel.Code_ExecuteSuccess))
		}
	}
	return api.FormatBatchWriteResponse(resp)
}

// GetSkillGroups queries skill groups with filters
func (s *Server) GetSkillGroups(ctx context.Context, query map[string]string) *apimodel.BatchQueryResponse {
	if s.storage == nil {
		return api.NewBatchQueryResponse(apimodel.Code_StoreLayerException)
	}

	offset, limit := parseOffsetLimit(query)
	count, groups, err := s.storage.QuerySkillGroups(query, offset, limit)
	if err != nil {
		log.Errorf("[SkillGroup] query skill groups failed, err: %v", err)
		return api.NewBatchQueryResponse(apimodel.Code_StoreLayerException)
	}

	resp := api.NewBatchQueryResponse(apimodel.Code_ExecuteSuccess)
	resp.Amount = count
	resp.Size = uint32(len(groups))
	return resp
}

// parseOffsetLimit parses offset and limit from query params
func parseOffsetLimit(query map[string]string) (uint32, uint32) {
	var offset, limit uint32 = 0, 100

	if o, ok := query["offset"]; ok {
		if v, err := parseUint32(o); err == nil {
			offset = v
		}
	}
	if l, ok := query["limit"]; ok {
		if v, err := parseUint32(l); err == nil {
			limit = v
		}
	}
	return offset, limit
}

func parseUint32(s string) (uint32, error) {
	var v uint32
	_, err := fmt.Sscanf(s, "%d", &v)
	return v, err
}

// ===== SkillVersion Batch Operations =====

// CreateSkillVersions creates multiple skill versions
func (s *Server) CreateSkillVersions(ctx context.Context, versions []*aiTypes.SkillVersion) *apimodel.BatchWriteResponse {
	resp := api.NewBatchWriteResponse(apimodel.Code_ExecuteSuccess)
	for _, version := range versions {
		if err := s.CreateSkillVersion(ctx, version); err != nil {
			log.Errorf("[SkillVersion] create skill version failed, skill: %s, err: %v",
				version.SkillName, err)
			api.Collect(resp, api.NewResponseWithMsg(apimodel.Code_ExecuteException, err.Error()))
		} else {
			api.Collect(resp, api.NewResponse(apimodel.Code_ExecuteSuccess))
		}
	}
	return api.FormatBatchWriteResponse(resp)
}

// DeleteSkillVersions deletes multiple skill versions
func (s *Server) DeleteSkillVersions(ctx context.Context, ids []string) *apimodel.BatchWriteResponse {
	resp := api.NewBatchWriteResponse(apimodel.Code_ExecuteSuccess)
	for _, id := range ids {
		if err := s.storage.DeleteSkillVersion(id); err != nil {
			log.Errorf("[SkillVersion] delete skill version failed, id: %s, err: %v", id, err)
			api.Collect(resp, api.NewResponseWithMsg(apimodel.Code_ExecuteException, err.Error()))
		} else {
			api.Collect(resp, api.NewResponse(apimodel.Code_ExecuteSuccess))
		}
	}
	return api.FormatBatchWriteResponse(resp)
}

// GetSkillVersions queries skill versions with filters
func (s *Server) GetSkillVersions(ctx context.Context, query map[string]string) *apimodel.BatchQueryResponse {
	if s.storage == nil {
		return api.NewBatchQueryResponse(apimodel.Code_StoreLayerException)
	}

	offset, limit := parseOffsetLimit(query)
	count, versions, err := s.storage.QuerySkillVersions(query, offset, limit)
	if err != nil {
		log.Errorf("[SkillVersion] query skill versions failed, err: %v", err)
		return api.NewBatchQueryResponse(apimodel.Code_StoreLayerException)
	}

	resp := api.NewBatchQueryResponse(apimodel.Code_ExecuteSuccess)
	resp.Amount = count
	resp.Size = uint32(len(versions))
	return resp
}

// ===== SkillSubscription Batch Operations =====

// CreateSkillSubscriptions creates multiple skill subscriptions
func (s *Server) CreateSkillSubscriptions(ctx context.Context, subs []*aiTypes.SkillSubscription) *apimodel.BatchWriteResponse {
	resp := api.NewBatchWriteResponse(apimodel.Code_ExecuteSuccess)
	for _, sub := range subs {
		if err := s.CreateSkillSubscription(ctx, sub); err != nil {
			log.Errorf("[SkillSubscription] create skill subscription failed, skill: %s, client: %s, err: %v",
				sub.SkillName, sub.ClientID, err)
			api.Collect(resp, api.NewResponseWithMsg(apimodel.Code_ExecuteException, err.Error()))
		} else {
			api.Collect(resp, api.NewResponse(apimodel.Code_ExecuteSuccess))
		}
	}
	return api.FormatBatchWriteResponse(resp)
}

// DeleteSkillSubscriptions deletes multiple skill subscriptions
func (s *Server) DeleteSkillSubscriptions(ctx context.Context, ids []string) *apimodel.BatchWriteResponse {
	resp := api.NewBatchWriteResponse(apimodel.Code_ExecuteSuccess)
	for _, id := range ids {
		if err := s.DeleteSkillSubscription(ctx, id); err != nil {
			log.Errorf("[SkillSubscription] delete skill subscription failed, id: %s, err: %v", id, err)
			api.Collect(resp, api.NewResponseWithMsg(apimodel.Code_ExecuteException, err.Error()))
		} else {
			api.Collect(resp, api.NewResponse(apimodel.Code_ExecuteSuccess))
		}
	}
	return api.FormatBatchWriteResponse(resp)
}

// GetSkillSubscriptions queries skill subscriptions with filters
func (s *Server) GetSkillSubscriptions(ctx context.Context, query map[string]string) *apimodel.BatchQueryResponse {
	if s.storage == nil {
		return api.NewBatchQueryResponse(apimodel.Code_StoreLayerException)
	}

	offset, limit := parseOffsetLimit(query)
	count, subs, err := s.storage.QuerySkillSubscriptions(query, offset, limit)
	if err != nil {
		log.Errorf("[SkillSubscription] query skill subscriptions failed, err: %v", err)
		return api.NewBatchQueryResponse(apimodel.Code_StoreLayerException)
	}

	resp := api.NewBatchQueryResponse(apimodel.Code_ExecuteSuccess)
	resp.Amount = count
	resp.Size = uint32(len(subs))
	return resp
}

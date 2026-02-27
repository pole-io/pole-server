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

	aiTypes "github.com/pole-io/pole-server/apis/pkg/types/ai"
	"github.com/pole-io/pole-server/apis/store"
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

	// SkillGroup CRUD operations
	CreateSkillGroup(ctx context.Context, group *aiTypes.SkillGroup) error
	UpdateSkillGroup(ctx context.Context, group *aiTypes.SkillGroup) error
	DeleteSkillGroup(ctx context.Context, namespace, name string) error
	GetSkillGroup(ctx context.Context, namespace, name string) (*aiTypes.SkillGroup, error)

	// SkillVersion operations
	CreateSkillVersion(ctx context.Context, version *aiTypes.SkillVersion) error
	ActivateSkillVersion(ctx context.Context, versionID string) error

	// SkillSubscription operations
	CreateSkillSubscription(ctx context.Context, sub *aiTypes.SkillSubscription) error
	DeleteSkillSubscription(ctx context.Context, id string) error
	GetSkillSubscriptionsBySkill(ctx context.Context, skillName, namespace string) ([]*aiTypes.SkillSubscription, error)
	GetSkillSubscriptionsByClient(ctx context.Context, clientID string) ([]*aiTypes.SkillSubscription, error)
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

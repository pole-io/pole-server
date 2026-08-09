/*
 * Tencent is pleased to support the open source community by making Polaris available.
 *
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *
 * Licensed under the BSD 3-Clause License (the "License");
 * you may not use this file except in compliance with the License.
 */

package config

import (
	"time"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
)

// TemplateValueReleaseType identifies the audience of a Namespace template Value release.
type TemplateValueReleaseType string

const (
	TemplateValueReleaseTypeNormal TemplateValueReleaseType = "normal"
	TemplateValueReleaseTypeGray   TemplateValueReleaseType = "gray"
)

// ConfigTemplateRelease is an internal immutable template snapshot reused by environment releases.
type ConfigTemplateRelease struct {
	ID              string
	TemplateID      uint64
	Name            string
	Content         string
	Format          string
	ParameterSchema string
	Engine          string
	EngineVersion   string
	Version         uint64
	ContentSHA256   string
	Comment         string
	CreateTime      time.Time
	CreateBy        string
}

// NamespaceTemplateValues is the mutable Value draft aggregate scoped by Namespace and template.
type NamespaceTemplateValues struct {
	ID         string
	Namespace  string
	TemplateID uint64
	Values     string
	Revision   string
	CreateTime time.Time
	CreateBy   string
	ModifyTime time.Time
	ModifyBy   string
}

// NamespaceConfigTemplateDraft is the mutable template definition scoped by
// Namespace and logical template identity. Labels and name remain global.
type NamespaceConfigTemplateDraft struct {
	Namespace       string
	TemplateID      uint64
	Name            string
	Comment         string
	Content         string
	Format          string
	ParameterSchema string
	Engine          string
	EngineVersion   string
	Revision        string
	DraftVersion    uint64
	InitializedFrom string
	CreateTime      time.Time
	CreateBy        string
	ModifyTime      time.Time
	ModifyBy        string
}

// NamespaceTemplateValueRelease is the compatibility representation of an immutable environment config release.
// It atomically binds TemplateReleaseID to the encrypted Value snapshot and rollout state.
type NamespaceTemplateValueRelease struct {
	ID                string
	ValuesID          string
	Namespace         string
	TemplateID        uint64
	TemplateReleaseID string
	Values            string
	ReleaseType       TemplateValueReleaseType
	BetaLabels        []*apimodel.ClientLabel
	Priority          int32
	Active            bool
	Version           uint64
	Revision          string
	Comment           string
	CreateTime        time.Time
	CreateBy          string
	ModifyTime        time.Time
	ModifyBy          string
}

// ConfigTemplateBinding pins a config file to an explicit immutable template release.
type ConfigTemplateBinding struct {
	BindingReleaseID  string
	Namespace         string
	Group             string
	FileName          string
	TemplateID        uint64
	TemplateReleaseID string
	Active            bool
	Version           uint64
	Comment           string
	CreateTime        time.Time
	CreateBy          string
	ModifyTime        time.Time
	ModifyBy          string
}

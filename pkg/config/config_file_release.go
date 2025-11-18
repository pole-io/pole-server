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

package config

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/golang/protobuf/jsonpb"
	"go.uber.org/zap"

	apiconfig "github.com/pole-io/specification/source/go/api/v1/config_manage"
	apimodel "github.com/pole-io/specification/source/go/api/v1/model"

	cacheapi "github.com/pole-io/pole-server/apis/cache"
	"github.com/pole-io/pole-server/apis/pkg/types"
	conftypes "github.com/pole-io/pole-server/apis/pkg/types/config"
	"github.com/pole-io/pole-server/apis/pkg/types/rules"
	"github.com/pole-io/pole-server/apis/store"
	storeapi "github.com/pole-io/pole-server/apis/store"
	api "github.com/pole-io/pole-server/pkg/common/api/v1"
	"github.com/pole-io/pole-server/pkg/common/utils"
	"github.com/pole-io/pole-server/pkg/common/utils/valid"
	"github.com/pole-io/pole-server/pkg/goverrule"
)

// PublishConfigFile 发布配置文件
func (s *Server) PublishConfigFile(ctx context.Context, req *apiconfig.ConfigFileRelease) *apimodel.Response {
	tx, err := s.storage.StartTx()
	if err != nil {
		log.Error("[Config][Release] publish config file begin tx.", utils.RequestID(ctx), zap.Error(err))
		return api.NewConfigResponse(storeapi.StoreCode2APICode(err))
	}
	defer func() {
		_ = tx.Rollback()
	}()

	data, resp := s.handlePublishConfigFile(ctx, tx, req)
	if resp.GetCode() != uint32(apimodel.Code_ExecuteSuccess) {
		_ = tx.Rollback()
		return resp
	}

	if err := tx.Commit(); err != nil {
		log.Error("[Config][Release] publish config file commit tx.", utils.RequestID(ctx), zap.Error(err))
		return api.NewConfigResponse(storeapi.StoreCode2APICode(err))
	}
	if req.ReleaseType == conftypes.ReleaseTypeGray {
		s.recordReleaseSuccess(ctx, conftypes.ReleaseTypeGray, data)
	} else {
		s.recordReleaseSuccess(ctx, conftypes.ReleaseTypeNormal, data)
	}

	return resp
}

func (s *Server) nextSequence() int64 {
	return atomic.AddInt64(&s.sequence, 1)
}

// PublishConfigFile 发布配置文件
func (s *Server) handlePublishConfigFile(ctx context.Context, tx store.Tx,
	req *apiconfig.ConfigFileRelease) (*conftypes.ConfigFileRelease, *apimodel.Response) {
	namespace := req.GetNamespace()
	group := req.GetGroup()
	fileName := req.GetFileName()

	fileRelease := &conftypes.ConfigFileRelease{
		SimpleConfigFileRelease: &conftypes.SimpleConfigFileRelease{
			ConfigFileReleaseKey: &conftypes.ConfigFileReleaseKey{
				Name:        req.GetName(),
				Namespace:   namespace,
				Group:       group,
				FileName:    fileName,
				ReleaseType: rules.ReleaseType(req.GetReleaseType()),
			},
			BetaLabels: req.GetBetaLabels(),
		},
	}

	// 确认是否存在正在灰度发布中的配置文件
	betaRelease, err := s.storage.GetConfigFileBetaReleaseTx(tx, fileRelease.ToFileKey())
	if err != nil {
		log.Error("[Config][File] get beta config file release in get target.", utils.RequestID(ctx), zap.Error(err))
		return nil, api.NewConfigResponse(storeapi.StoreCode2APICode(err))
	}
	if betaRelease != nil {
		log.Error("[Config][File] still exist beta config file release.", utils.RequestID(ctx), zap.Error(err))
		return nil, api.NewConfigResponse(apimodel.Code_DataConflict)
	}

	// 获取待发布的 configFile 信息
	toPublishFile, err := s.storage.GetConfigFileTx(tx, namespace, group, fileName)
	if err != nil {
		log.Error("[Config][Release] publish config file when get file.", utils.RequestID(ctx),
			utils.ZapNamespace(namespace), utils.ZapGroup(group), utils.ZapFileName(fileName),
			zap.Error(err))
		return nil, api.NewConfigResponse(storeapi.StoreCode2APICode(err))
	}
	if toPublishFile == nil {
		return nil, api.NewConfigResponse(apimodel.Code_NotFoundResource)
	}
	if releaseName := req.GetName(); releaseName == "" {
		// 这里要保证每一次发布都有唯一的 release_name 名称
		req.Name = fmt.Sprintf("%s-%d-%d", fileName, time.Now().Unix(), s.nextSequence())
	}

	fileRelease.Name = req.GetName()
	fileRelease.Format = toPublishFile.Format
	fileRelease.Metadata = toPublishFile.Metadata
	fileRelease.Comment = req.GetComment()
	fileRelease.Md5 = CalMd5(toPublishFile.Content)
	fileRelease.CreateBy = utils.ParseUserName(ctx)
	fileRelease.ModifyBy = utils.ParseUserName(ctx)
	fileRelease.ReleaseDescription = req.GetReleaseDescription()
	fileRelease.Content = toPublishFile.Content

	saveRelease, err := s.storage.GetConfigFileReleaseTx(tx, fileRelease.ConfigFileReleaseKey)
	if err != nil {
		log.Error("[Config][Release] publish config file when get release.",
			utils.RequestID(ctx), utils.ZapNamespace(namespace), utils.ZapGroup(group),
			utils.ZapFileName(fileName), zap.Error(err))
		return fileRelease, api.NewConfigResponse(storeapi.StoreCode2APICode(err))
	}
	// 重新激活
	if saveRelease != nil {
		// 不允许重复发布同一个版本
		return fileRelease, api.NewConfigResponse(apimodel.Code_ExistReleasedConfig)
	}

	if err = s.storage.CreateConfigFileReleaseTx(tx, fileRelease); err != nil {
		log.Error("[Config][Release] publish config file when create release.",
			utils.RequestID(ctx), utils.ZapNamespace(namespace), utils.ZapGroup(group),
			utils.ZapFileName(fileName), zap.Error(err))
		return fileRelease, api.NewConfigResponse(storeapi.StoreCode2APICode(err))
	}
	if req.GetReleaseType() == conftypes.ReleaseTypeGray {
		if errRsp := goverrule.SaveGrayRule(ctx, tx, s.storage, fileRelease); errRsp != nil {
			return fileRelease, api.NewConfigFileResponse(storeapi.StoreCode2APICode(err), nil)
		}
	}

	s.RecordHistory(ctx, configFileReleaseRecordEntry(ctx, req, fileRelease, types.OCreate))
	return fileRelease, api.NewConfigResponse(apimodel.Code_ExecuteSuccess)
}

// GetConfigFileRelease 获取配置文件发布内容
func (s *Server) GetConfigFileRelease(ctx context.Context, req *apiconfig.ConfigFileRelease) *apimodel.Response {
	namespace := req.GetNamespace()
	group := req.GetGroup()
	fileName := req.GetFileName()
	releaseName := req.GetName()
	var (
		ret *conftypes.ConfigFileRelease
		err error
	)

	// 如果没有指定专门的 releaseName，则直接查询 active 状态的配置发布, 兼容老的控制台查询逻辑
	if releaseName != "" {
		ret, err = s.storage.GetConfigFileRelease(&conftypes.ConfigFileReleaseKey{
			Namespace: namespace,
			Group:     group,
			FileName:  fileName,
			Name:      releaseName,
		})
	} else {
		ret, err = s.storage.GetConfigFileActiveRelease(&conftypes.ConfigFileKey{
			Namespace: namespace,
			Group:     group,
			Name:      fileName,
		})
	}

	if err != nil {
		log.Error("[Config][Release] get config file release.", utils.RequestID(ctx),
			utils.ZapNamespace(namespace), utils.ZapGroup(group), utils.ZapFileName(fileName), zap.Error(err))
		return api.NewConfigResponse(storeapi.StoreCode2APICode(err))
	}
	if ret == nil {
		return api.NewConfigResponse(apimodel.Code_ExecuteSuccess)
	}

	_ = s.caches.Gray().Update()
	ret, err = s.chains.AfterGetFileRelease(ctx, ret)
	if err != nil {
		log.Error("[Config][Release] get config file release run chain.", utils.RequestID(ctx),
			utils.ZapNamespace(namespace), utils.ZapGroup(group), utils.ZapFileName(fileName), zap.Error(err))
		out := api.NewConfigResponse(apimodel.Code_ExecuteException)
		return out
	}

	release := conftypes.ToConfiogFileReleaseApi(ret)
	return api.NewConfigFileReleaseResponse(apimodel.Code_ExecuteSuccess, release)
}

// DeleteConfigFileRelease 删除某个配置文件的发布 release
func (s *Server) DeleteConfigFileReleases(ctx context.Context,
	reqs []*apiconfig.ConfigFileRelease) *apimodel.BatchWriteResponse {

	responses := api.NewConfigBatchWriteResponse(apimodel.Code_ExecuteSuccess)
	chs := make([]chan *apimodel.Response, 0, len(reqs))
	for i, instance := range reqs {
		chs = append(chs, make(chan *apimodel.Response))
		go func(index int, ins *apiconfig.ConfigFileRelease) {
			chs[index] <- s.DeleteConfigFileRelease(ctx, ins)
		}(i, instance)
	}

	for _, ch := range chs {
		resp := <-ch
		api.ConfigCollect(responses, resp)
	}
	return responses
}

func (s *Server) DeleteConfigFileRelease(ctx context.Context,
	req *apiconfig.ConfigFileRelease) *apimodel.Response {
	release := &conftypes.ConfigFileRelease{
		SimpleConfigFileRelease: &conftypes.SimpleConfigFileRelease{
			ConfigFileReleaseKey: &conftypes.ConfigFileReleaseKey{
				Name:        req.GetName(),
				Namespace:   req.GetNamespace(),
				Group:       req.GetGroup(),
				FileName:    req.GetFileName(),
				ReleaseType: rules.ReleaseType(req.GetReleaseType()),
			},
		},
	}
	var (
		recordData *conftypes.ConfigFileRelease
	)

	tx, err := s.storage.StartTx()
	if err != nil {
		log.Error("[Config][File] delete config file release when begin tx.",
			utils.RequestID(ctx), zap.Error(err))
		return api.NewConfigResponse(storeapi.StoreCode2APICode(err))
	}
	defer func() {
		_ = tx.Rollback()
	}()
	if _, err := s.storage.LockConfigFile(tx, release.ToFileKey()); err != nil {
		log.Error("[Config][File] delete config file release when lock.",
			utils.RequestID(ctx), zap.Error(err))
		return api.NewConfigResponse(storeapi.StoreCode2APICode(err))
	}

	saveData, err := s.storage.GetConfigFileReleaseTx(tx, release.ConfigFileReleaseKey)
	if err != nil {
		return api.NewConfigResponse(storeapi.StoreCode2APICode(err))
	}
	recordData = saveData
	if saveData == nil {
		return api.NewConfigResponse(apimodel.Code_ExecuteSuccess)
	}
	// 如果存在处于 active 状态的配置，重新在激活一下，触发版本的更新变动
	if saveData.Active {
		if err := s.storage.ActiveConfigFileReleaseTx(tx, saveData); err != nil {
			log.Error("[Config][File] delete config file release when re-active.",
				utils.RequestID(ctx), zap.Error(err))
			return api.NewConfigResponse(storeapi.StoreCode2APICode(err))
		}
	}

	if err := s.storage.DeleteConfigFileReleaseTx(tx, saveData.ConfigFileReleaseKey); err != nil {
		log.Error("[Config][Release] delete config file release error.",
			utils.RequestID(ctx), utils.ZapNamespace(req.GetNamespace()),
			utils.ZapGroup(req.GetGroup()), utils.ZapFileName(req.GetFileName()),
			zap.Error(err))
		return api.NewConfigResponse(storeapi.StoreCode2APICode(err))
	}

	if err := tx.Commit(); err != nil {
		log.Error("[Config][Release] delete config file release when commit tx.",
			utils.RequestID(ctx), utils.ZapNamespace(req.GetNamespace()),
			utils.ZapGroup(req.GetGroup()), utils.ZapFileName(req.GetFileName()),
			zap.Error(err))
		return api.NewConfigResponse(storeapi.StoreCode2APICode(err))
	}
	s.recordReleaseSuccess(ctx, conftypes.ReleaseTypeDelete, recordData)
	s.RecordHistory(ctx, configFileReleaseRecordEntry(ctx, req, release, types.ODelete))
	return api.NewConfigResponse(apimodel.Code_ExecuteSuccess)
}

func (s *Server) GetConfigFileReleaseVersions(ctx context.Context,
	searchFilters map[string]string) *apimodel.BatchQueryResponse {

	args := cacheapi.ConfigReleaseArgs{
		BaseConfigArgs: cacheapi.BaseConfigArgs{
			Namespace: searchFilters["namespace"],
			Group:     searchFilters["group"],
		},
		FileName:    searchFilters["file_name"],
		OnlyActive:  false,
		NoPage:      true,
		IncludeGray: true,
	}
	return s.handleDescribeConfigFileReleases(ctx, args)
}

func (s *Server) GetConfigFileReleases(ctx context.Context,
	searchFilters map[string]string) *apimodel.BatchQueryResponse {

	offset, limit, _ := valid.ParseOffsetAndLimit(searchFilters)

	args := cacheapi.ConfigReleaseArgs{
		BaseConfigArgs: cacheapi.BaseConfigArgs{
			Namespace:  searchFilters["namespace"],
			Group:      searchFilters["group"],
			Offset:     offset,
			Limit:      limit,
			OrderField: searchFilters["order_field"],
			OrderType:  searchFilters["order_type"],
		},
		FileName:    searchFilters["file_name"],
		ReleaseName: searchFilters["release_name"],
		OnlyActive:  strings.Compare(searchFilters["only_active"], "true") == 0,
		IncludeGray: true,
	}
	return s.handleDescribeConfigFileReleases(ctx, args)
}

func (s *Server) handleDescribeConfigFileReleases(ctx context.Context, args cacheapi.ConfigReleaseArgs) *apimodel.BatchQueryResponse {
	total, simpleReleases, err := s.fileCache.QueryReleases(&args)
	if err != nil {
		return api.NewConfigBatchQueryResponseWithInfo(apimodel.Code_ExecuteException, err.Error())
	}
	ret := make([]*apiconfig.ConfigFileRelease, 0, len(simpleReleases))
	for i := range simpleReleases {
		item := simpleReleases[i]
		tmp, err := s.chains.AfterGetFileRelease(ctx, &conftypes.ConfigFileRelease{
			SimpleConfigFileRelease: simpleReleases[i],
		})
		if err != nil {
			log.Error("[Config][File] get config file release run chain.", utils.RequestID(ctx),
				zap.String("namespace", item.Namespace), zap.String("group", item.Group),
				zap.String("fileName", item.FileName), zap.Error(err))
			return api.NewConfigBatchQueryResponseWithInfo(apimodel.Code_ExecuteException, err.Error())
		}
		item = tmp.SimpleConfigFileRelease
		viewData := &apiconfig.ConfigFileRelease{
			Id:                 item.Id,
			Name:               item.Name,
			Namespace:          item.Namespace,
			Group:              item.Group,
			FileName:           item.FileName,
			Format:             item.Format,
			Version:            item.Version,
			Active:             item.Active,
			CreateBy:           item.CreateBy,
			ModifyBy:           item.ModifyBy,
			ReleaseDescription: item.ReleaseDescription,
			ReleaseType:        string(item.ReleaseType),
		}
		// 查询配置灰度规则标签
		if item.ReleaseType == conftypes.ReleaseTypeGray {
			viewData.BetaLabels = s.caches.Gray().GetGrayRule(GetGrayConfigReaseKey(item))
		}
		ret = append(ret, viewData)
	}

	interfaceRet := make([]interface{}, len(ret))
	for i, r := range ret {
		interfaceRet[i] = r
	}

	resp := api.NewConfigFileReleaseHistoryQueryResponse(apimodel.Code_ExecuteSuccess, total, interfaceRet)
	return resp
}

// RollbackConfigFileReleases 批量回滚配置
func (s *Server) RollbackConfigFileReleases(ctx context.Context,
	reqs []*apiconfig.ConfigFileRelease) *apimodel.BatchWriteResponse {

	responses := api.NewConfigBatchWriteResponse(apimodel.Code_ExecuteSuccess)
	chs := make([]chan *apimodel.Response, 0, len(reqs))
	for i, instance := range reqs {
		chs = append(chs, make(chan *apimodel.Response))
		go func(index int, ins *apiconfig.ConfigFileRelease) {
			chs[index] <- s.RollbackConfigFileRelease(ctx, ins)
		}(i, instance)
	}

	for _, ch := range chs {
		resp := <-ch
		api.ConfigCollect(responses, resp)
	}
	return responses
}

// RollbackConfigFileRelease 回滚配置
func (s *Server) RollbackConfigFileRelease(ctx context.Context,
	req *apiconfig.ConfigFileRelease) *apimodel.Response {
	data := &conftypes.ConfigFileRelease{
		SimpleConfigFileRelease: &conftypes.SimpleConfigFileRelease{
			ConfigFileReleaseKey: &conftypes.ConfigFileReleaseKey{
				Name:        req.GetName(),
				Namespace:   req.GetNamespace(),
				Group:       req.GetGroup(),
				FileName:    req.GetFileName(),
				ReleaseType: conftypes.ReleaseTypeNormal,
			},
		},
	}

	tx, err := s.storage.StartTx()
	if err != nil {
		log.Error("[Config][File] rollback config file releasw when begin tx.",
			utils.RequestID(ctx), zap.Error(err))
		return api.NewConfigResponse(storeapi.StoreCode2APICode(err))
	}
	defer func() {
		_ = tx.Rollback()
	}()

	targetRelease, ret := s.handleRollbackConfigFileRelease(ctx, tx, data)
	if targetRelease != nil {
		data = targetRelease
	}
	if ret != nil {
		_ = tx.Rollback()
		return ret
	}

	if err := tx.Commit(); err != nil {
		log.Error("[Config][File] rollback config file releasw when commit tx.",
			utils.RequestID(ctx), zap.Error(err))
		return api.NewConfigResponse(storeapi.StoreCode2APICode(err))
	}

	s.recordReleaseSuccess(ctx, conftypes.ReleaseTypeRollback, data)
	s.RecordHistory(ctx, configFileReleaseRecordEntry(ctx, req, data, types.ORollback))
	return api.NewConfigResponse(apimodel.Code_ExecuteSuccess)
}

// handleRollbackConfigFileRelease 回滚配置
func (s *Server) handleRollbackConfigFileRelease(ctx context.Context, tx store.Tx,
	data *conftypes.ConfigFileRelease) (*conftypes.ConfigFileRelease, *apimodel.Response) {

	targetRelease, err := s.storage.GetConfigFileReleaseTx(tx, data.ConfigFileReleaseKey)
	if err != nil {
		log.Error("[Config][Release] rollback config file get target release", zap.Error(err))
		return nil, api.NewConfigResponse(storeapi.StoreCode2APICode(err))
	}
	if targetRelease == nil {
		log.Error("[Config][Release] rollback config file to target release not found")
		return nil, api.NewConfigResponse(apimodel.Code_NotFoundResource)
	}

	if err := s.storage.ActiveConfigFileReleaseTx(tx, data); err != nil {
		log.Error("[Config][Release] rollback config file release error.",
			utils.RequestID(ctx), zap.String("namespace", data.Namespace),
			zap.String("group", data.Group), zap.String("fileName", data.FileName), zap.Error(err))
		return targetRelease, api.NewConfigResponse(storeapi.StoreCode2APICode(err))
	}
	return targetRelease, nil
}

// CasUpsertAndReleaseConfigFile 根据版本比对决定是否允许进行配置修改发布
func (s *Server) CasUpsertAndReleaseConfigFile(ctx context.Context,
	req *apiconfig.ConfigFilePublishInfo) *apimodel.Response {
	upsertFileReq := &apiconfig.ConfigFile{
		Name:        req.GetFileName(),
		Namespace:   req.GetNamespace(),
		Group:       req.GetGroup(),
		Content:     req.GetContent(),
		Format:      req.GetFormat(),
		Comment:     req.GetComment(),
	}
	if rsp := s.prepareCreateConfigFile(ctx, upsertFileReq); rsp.Code != api.ExecuteSuccess {
		return rsp
	}

	tx, err := s.storage.StartTx()
	if err != nil {
		log.Error("[Config][File] upsert config file when begin tx.", utils.RequestID(ctx),
			zap.String("namespace", req.GetNamespace()), zap.String("group", req.GetGroup()),
			zap.String("fileName", req.GetFileName()), zap.Error(err))
		return api.NewConfigResponse(storeapi.StoreCode2APICode(err))
	}

	defer func() {
		_ = tx.Rollback()
	}()
	saveFile, err := s.storage.LockConfigFile(tx, &conftypes.ConfigFileKey{
		Namespace: req.GetNamespace(),
		Group:     req.GetGroup(),
		Name:      req.GetFileName(),
	})
	if err != nil {
		log.Error("[Config][File] lock config file when begin tx.", utils.RequestID(ctx),
			zap.String("namespace", req.GetNamespace()), zap.String("group", req.GetGroup()),
			zap.String("fileName", req.GetFileName()), zap.Error(err))
		return api.NewConfigResponse(storeapi.StoreCode2APICode(err))
	}

	historyRecords := []func(){}

	var upsertResp *apimodel.Response
	if saveFile == nil {
		upsertResp = s.handleCreateConfigFile(ctx, tx, upsertFileReq)
		historyRecords = append(historyRecords, func() {
			s.RecordHistory(ctx, configFileRecordEntry(ctx, upsertFileReq, types.OCreate))
		})
	} else {
		actualMd5 := CalMd5(saveFile.Content)
		if req.Md5 != actualMd5 {
			log.Error("[Config][File] cas compare config file.", utils.RequestID(ctx),
				zap.String("namespace", req.Namespace), zap.String("group", req.Group),
				zap.String("fileName", req.FileName),
				zap.String("expect", req.Md5),zap.String("actual", actualMd5))
			return api.NewConfigResponse(apimodel.Code_DataConflict)
		}
		upsertResp = s.handleUpdateConfigFile(ctx, tx, upsertFileReq)
		historyRecords = append(historyRecords, func() {
			s.RecordHistory(ctx, configFileRecordEntry(ctx, upsertFileReq, types.OUpdate))
		})
	}
	if upsertResp.GetCode() != uint32(apimodel.Code_ExecuteSuccess) {
		return upsertResp
	}

	data, releaseResp := s.handlePublishConfigFile(ctx, tx, &apiconfig.ConfigFileRelease{
		Name:               req.GetReleaseName(),
		Namespace:          req.GetNamespace(),
		Group:              req.GetGroup(),
		FileName:           req.GetFileName(),
		CreateBy:           utils.ParseUserName(ctx),
		ModifyBy:           utils.ParseUserName(ctx),
		ReleaseDescription: req.GetReleaseDescription(),
	})
	if releaseResp.GetCode() != uint32(apimodel.Code_ExecuteSuccess) {
		_ = tx.Rollback()
		return releaseResp
	}

	if err := tx.Commit(); err != nil {
		log.Error("[Config][File] upsert config file when commit tx.", utils.RequestID(ctx), zap.Error(err))
		return api.NewConfigResponse(storeapi.StoreCode2APICode(err))
	}
	for i := range historyRecords {
		historyRecords[i]()
	}
	s.recordReleaseHistory(ctx, data, conftypes.ReleaseTypeNormal, conftypes.ReleaseStatusSuccess, "")
	return releaseResp
}

func (s *Server) UpsertAndReleaseConfigFile(ctx context.Context,
	req *apiconfig.ConfigFilePublishInfo) *apimodel.Response {
	upsertFileReq := &apiconfig.ConfigFile{
		Name:        req.GetFileName(),
		Namespace:   req.GetNamespace(),
		Group:       req.GetGroup(),
		Content:     req.GetContent(),
		Format:      req.GetFormat(),
		Comment:     req.GetComment(),
	}
	if rsp := s.prepareCreateConfigFile(ctx, upsertFileReq); rsp.Code != api.ExecuteSuccess {
		return rsp
	}

	tx, err := s.storage.StartTx()
	if err != nil {
		log.Error("[Config][File] upsert config file when begin tx.", utils.RequestID(ctx),
			zap.String("namespace", req.GetNamespace()), zap.String("group", req.GetGroup()),
			zap.String("fileName", req.GetFileName()), zap.Error(err))
		return api.NewConfigResponse(storeapi.StoreCode2APICode(err))
	}

	defer func() {
		_ = tx.Rollback()
	}()
	saveFile, err := s.storage.LockConfigFile(tx, &conftypes.ConfigFileKey{
		Namespace: req.GetNamespace(),
		Group:     req.GetGroup(),
		Name:      req.GetFileName(),
	})
	if err != nil {
		log.Error("[Config][File] lock config file when begin tx.", utils.RequestID(ctx),
			zap.String("namespace", req.GetNamespace()), zap.String("group", req.GetGroup()),
			zap.String("fileName", req.GetFileName()), zap.Error(err))
		return api.NewConfigResponse(storeapi.StoreCode2APICode(err))
	}

	historyRecords := []func(){}

	var upsertResp *apimodel.Response
	if saveFile == nil {
		upsertResp = s.handleCreateConfigFile(ctx, tx, upsertFileReq)
		historyRecords = append(historyRecords, func() {
			s.RecordHistory(ctx, configFileRecordEntry(ctx, upsertFileReq, types.OCreate))
		})
	} else {
		actualMd5 := CalMd5(saveFile.Content)
		// 只有显示设置了 md5 字段值才会进入 CAS 发布流程
		if req.GetMd5()!= "" && req.GetMd5()!= actualMd5 {
			log.Error("[Config][File] cas compare config file.", utils.RequestID(ctx),
				zap.String("namespace", req.GetNamespace()), zap.String("group", req.GetGroup()),
				zap.String("fileName", req.GetFileName()),
				zap.String("expect", req.GetMd5()), zap.String("actual", actualMd5))
			return api.NewConfigResponse(apimodel.Code_DataConflict)
		}
		upsertResp = s.handleUpdateConfigFile(ctx, tx, upsertFileReq)
		historyRecords = append(historyRecords, func() {
			s.RecordHistory(ctx, configFileRecordEntry(ctx, upsertFileReq, types.OUpdate))
		})
	}
	if upsertResp.GetCode() != uint32(apimodel.Code_ExecuteSuccess) {
		return upsertResp
	}

	data, releaseResp := s.handlePublishConfigFile(ctx, tx, &apiconfig.ConfigFileRelease{
		Name:               req.GetReleaseName(),
		Namespace:          req.GetNamespace(),
		Group:              req.GetGroup(),
		FileName:           req.GetFileName(),
		CreateBy:           utils.ParseUserName(ctx),
		ModifyBy:           utils.ParseUserName(ctx),
		ReleaseDescription: req.GetReleaseDescription(),
	})
	if releaseResp.GetCode()!= uint32(apimodel.Code_ExecuteSuccess) {
		_ = tx.Rollback()
		return releaseResp
	}

	if err := tx.Commit(); err != nil {
		log.Error("[Config][File] upsert config file when commit tx.", utils.RequestID(ctx), zap.Error(err))
		return api.NewConfigResponse(storeapi.StoreCode2APICode(err))
	}
	for i := range historyRecords {
		historyRecords[i]()
	}
	s.recordReleaseHistory(ctx, data, conftypes.ReleaseTypeNormal, conftypes.ReleaseStatusSuccess, "")
	return releaseResp
}

func (s *Server) StopGrayConfigFileReleases(ctx context.Context, reqs []*apiconfig.ConfigFileRelease) *apimodel.BatchWriteResponse {
	responses := api.NewConfigBatchWriteResponse(apimodel.Code_ExecuteSuccess)
	chs := make([]chan *apimodel.Response, 0, len(reqs))
	for i, instance := range reqs {
		chs = append(chs, make(chan *apimodel.Response))
		go func(index int, ins *apiconfig.ConfigFileRelease) {
			chs[index] <- s.StopGrayConfigFileRelease(ctx, ins)
		}(i, instance)
	}

	for _, ch := range chs {
		resp := <-ch
		api.ConfigCollect(responses, resp)
	}
	return responses
}

func (s *Server) StopGrayConfigFileRelease(ctx context.Context, req *apiconfig.ConfigFileRelease) *apimodel.Response {
	if err := valid.CheckResourceName(req.GetNamespace()); err != nil {
		return api.NewConfigResponseWithInfo(apimodel.Code_BadRequest, "invalid config namespace")
	}
	if err := valid.CheckResourceName(req.GetGroup()); err != nil {
		return api.NewConfigResponseWithInfo(apimodel.Code_BadRequest, "invalid config group")
	}
	if req.GetFileName() == "" {
		return api.NewConfigResponseWithInfo(apimodel.Code_BadRequest, "invalid config file_name")
	}
	tx, err := s.storage.StartTx()
	if err != nil {
		log.Error("[Config][File] stop beta config file when begin tx.", utils.RequestID(ctx), zap.Error(err))
		return api.NewConfigResponse(storeapi.StoreCode2APICode(err))
	}

	defer func() {
		_ = tx.Rollback()
	}()

	fileKey := &conftypes.ConfigFileKey{
		Namespace: req.GetNamespace(),
		Group:     req.GetGroup(),
		Name:      req.GetFileName(),
	}

	if _, err := s.storage.LockConfigFile(tx, fileKey); err != nil {
		log.Error("[Config][File] stop beta config file release in lock file.", utils.RequestID(ctx), zap.Error(err))
		return api.NewConfigResponse(storeapi.StoreCode2APICode(err))
	}
	betaRelease, err := s.storage.GetConfigFileBetaReleaseTx(tx, fileKey)
	if err != nil {
		log.Error("[Config][File] stop beta config file release in get target.", utils.RequestID(ctx), zap.Error(err))
		return api.NewConfigResponse(storeapi.StoreCode2APICode(err))
	}
	if betaRelease == nil {
		return api.NewConfigResponse(apimodel.Code_ExecuteSuccess)
	}
	if err := s.storage.CleanGrayResource(tx, &rules.GrayResource{
		Name: GetGrayConfigReaseKey(&conftypes.SimpleConfigFileRelease{
			ConfigFileReleaseKey: &conftypes.ConfigFileReleaseKey{
				Namespace:   req.GetNamespace(),
				Group:       req.GetGroup(),
				Name:        req.GetFileName(),
				ReleaseType: conftypes.ReleaseTypeGray,
			},
		}),
	}); err != nil {
		log.Error("[Config][File] stop beta config file release when clean beta rule.", utils.RequestID(ctx), zap.Error(err))
		return api.NewConfigResponse(storeapi.StoreCode2APICode(err))
	}

	if err = s.storage.InactiveConfigFileReleaseTx(tx, betaRelease); err != nil {
		log.Error("[Config][File] stop beta config file release.", utils.RequestID(ctx), zap.Error(err))
		return api.NewConfigResponse(storeapi.StoreCode2APICode(err))
	}
	if err := tx.Commit(); err != nil {
		log.Error("[Config][File] stop config file release when commit tx.", utils.RequestID(ctx), zap.Error(err))
		return api.NewConfigResponse(storeapi.StoreCode2APICode(err))
	}
	s.recordReleaseHistory(ctx, betaRelease, conftypes.ReleaseTypeCancelGray, conftypes.ReleaseStatusSuccess, "")
	return api.NewConfigResponse(apimodel.Code_ExecuteSuccess)
}

func (s *Server) cleanConfigFileReleases(ctx context.Context, tx store.Tx,
	file *conftypes.ConfigFile) *apimodel.Response {

	// 先重新 active 下当前正在发布的
	saveData, err := s.storage.GetConfigFileActiveReleaseTx(tx, file.Key())
	if err != nil {
		return api.NewConfigResponse(storeapi.StoreCode2APICode(err))
	}
	if saveData != nil {
		if err := s.storage.ActiveConfigFileReleaseTx(tx, saveData); err != nil {
			return api.NewConfigResponse(storeapi.StoreCode2APICode(err))
		}
	}
	if err := s.storage.CleanConfigFileReleasesTx(tx, file.Namespace, file.Group, file.Name); err != nil {
		return api.NewConfigResponse(storeapi.StoreCode2APICode(err))
	}
	return nil
}

func (s *Server) recordReleaseSuccess(ctx context.Context, rType string, release *conftypes.ConfigFileRelease) {
	s.recordReleaseHistory(ctx, release, rType, conftypes.ReleaseStatusSuccess, "")
}

// configFileReleaseRecordEntry 生成服务的记录entry
func configFileReleaseRecordEntry(ctx context.Context, req *apiconfig.ConfigFileRelease, md *conftypes.ConfigFileRelease,
	operationType types.OperationType) *types.RecordEntry {

	marshaler := jsonpb.Marshaler{}
	detail, _ := marshaler.MarshalToString(req)

	entry := &types.RecordEntry{
		ResourceType:  types.RConfigFileRelease,
		ResourceName:  req.GetName(),
		Namespace:     req.GetNamespace(),
		OperationType: operationType,
		Operator:      utils.ParseOperator(ctx),
		Detail:        detail,
		HappenTime:    time.Now(),
	}

	return entry
}

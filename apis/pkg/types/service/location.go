package service

import (
	"github.com/golang/protobuf/ptypes/wrappers"
	apimodel "github.com/polarismesh/specification/source/go/api/v1/model"
)

// LocationStore 地域信息，对应数据库字段
type LocationStore struct {
	IP         string
	Region     string
	Zone       string
	Campus     string
	RegionID   uint32
	ZoneID     uint32
	CampusID   uint32
	Flag       int
	ModifyTime int64
}

// Location cmdb信息，对应内存结构体
type Location struct {
	Proto    *apimodel.Location
	RegionID uint32
	ZoneID   uint32
	CampusID uint32
	Valid    bool
}

// LocationView cmdb信息，对应内存结构体
type LocationView struct {
	IP       string
	Region   string
	Zone     string
	Campus   string
	RegionID uint32
	ZoneID   uint32
	CampusID uint32
}

// Store2Location 转成内存数据结构
func Store2Location(s *LocationStore) *Location {
	return &Location{
		Proto: &apimodel.Location{
			Region: &wrappers.StringValue{Value: s.Region},
			Zone:   &wrappers.StringValue{Value: s.Zone},
			Campus: &wrappers.StringValue{Value: s.Campus},
		},
		RegionID: s.RegionID,
		ZoneID:   s.ZoneID,
		CampusID: s.CampusID,
		Valid:    flag2valid(s.Flag),
	}
}

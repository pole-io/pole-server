package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/pole-io/pole-server/console/bootstrap"
	httpcommon "github.com/pole-io/pole-server/console/pkg/common/http"
	"github.com/pole-io/pole-server/console/pkg/common/model"
	store "github.com/pole-io/pole-server/console/pkg/observer"
)

var (
	_searchEventLogParams = map[string]struct{}{
		"namespace":  {},
		"service":    {},
		"event_type": {},
		"resource":   {},
		"start_time": {},
		"end_time":   {},
		"limit":      {},
		"cursor":     {},
		"direction":  {},
	}
)

func DescribeEventLog(conf *bootstrap.Config) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// if !verifyAccessPermission(ctx, conf) {
		// 	return
		// }
		reader, err := store.GetStore()
		if err != nil {
			ctx.JSON(http.StatusNotFound, model.Response{
				Code: 400404,
				Info: err.Error(),
			})
			return
		}
		if reader == nil {
			ctx.JSON(http.StatusOK, model.QueryResponse{
				Code:     200000,
				Size:     0,
				Amount:   0,
				Data:     []model.EventRecord{},
				HashNext: false,
			})
			return
		}

		filters := httpcommon.ParseQueryParams(ctx.Request)
		_, limit, _ := httpcommon.ParseOffsetAndLimit(filters)
		for key := range filters {
			if _, ok := _searchEventLogParams[key]; !ok {
				ctx.JSON(http.StatusBadRequest, model.Response{
					Code: 400000,
					Info: "invalid query parameter: " + key,
				})
				return
			}
		}
		index, _ := strconv.ParseInt(filters["cursor"], 10, 64)

		filters["limit"] = strconv.Itoa(int(limit))
		filters["offset"] = strconv.Itoa(int(limit) * (int(index) - 1))

		resp := &model.EventRecordLogResponse{}
		records, err := reader.GetEvents(filters)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, model.Response{
				Code: 500000,
				Info: err.Error(),
			})
			return
		}

		for _, record := range records {
			resp.Cursor = record.Cursor
		}

		resp.Code = 200000
		resp.Info = "success"
		resp.Size = uint32(len(records))
		resp.HasNext = len(records) == int(limit)
		resp.Data = records
		ctx.JSON(http.StatusOK, resp)
	}
}

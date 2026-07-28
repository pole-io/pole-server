package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	bootstrap "github.com/pole-io/pole-server/pkg/console/config"
	httpcommon "github.com/pole-io/pole-server/pkg/console/internal/common/http"
	"github.com/pole-io/pole-server/pkg/console/internal/common/model"
	store "github.com/pole-io/pole-server/pkg/console/internal/observer"
)

func DescribeOperationHistoryLog(conf *bootstrap.Config) gin.HandlerFunc {
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
				Data:     []model.OperationRecord{},
				HashNext: false,
			})
			return
		}

		filters := httpcommon.ParseQueryParams(ctx.Request)
		_, limit, _ := httpcommon.ParseOffsetAndLimit(filters)
		index, _ := strconv.ParseInt(filters["cursor"], 10, 64)

		filters["limit"] = strconv.Itoa(int(limit))
		filters["offset"] = strconv.Itoa(int(limit) * (int(index) - 1))

		resp := &model.OperationLogResponse{}
		records, err := reader.GetHistory(filters)
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

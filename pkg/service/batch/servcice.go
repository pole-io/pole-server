package batch

import (
	"context"

	svctypes "github.com/pole-io/pole-server/apis/pkg/types/service"
	"github.com/pole-io/pole-server/apis/store"
	storeapi "github.com/pole-io/pole-server/apis/store"
	"github.com/pole-io/pole-server/pkg/common/batchctrl"
	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
)

// NewBatchServiceSubscribersCtrl 服务订阅信息批量操作对象
func NewBatchServiceSubscribersCtrl(ctx context.Context, s store.Store, config *CtrlConfig) (*batchctrl.BatchController, error) {
	ctrl, err := newBatchCtrl(ctx, "service_subscribers", config, serviceSubscriberHandler(s))
	return ctrl, err
}

func serviceSubscriberHandler(s store.Store) func(futures []batchctrl.Future) {
	return func(futures []batchctrl.Future) {
		if len(futures) == 0 {
			return
		}

		log.Infof("[Batch] Start batch service subscriber count: %d", len(futures))

		// 调用batch接口，创建实例
		subscribers := make([]*svctypes.ServiceSubscriber, 0, len(futures))
		for _, entry := range futures {
			subscribers = append(subscribers, entry.Param().(*svctypes.ServiceSubscriber))
		}
		if err := s.AddServiceSubscibes(subscribers); err != nil {
			sendReply(futures, storeapi.StoreCode2APICode(err), err)
			return
		}
		sendReply(futures, apimodel.Code_ExecuteSuccess, nil)
	}
}

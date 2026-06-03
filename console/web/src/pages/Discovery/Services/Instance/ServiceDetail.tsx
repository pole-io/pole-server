import React, { } from 'react';
import { useAppDispatch, useAppSelector } from 'modules/store';
import { listOneService, selectService } from 'modules/discovery/service';
import { Descriptions, Space, Tag } from 'tdesign-react';
import { CheckVisibilityMode, VisibilityMode_Specified, VisibilityModeMap } from 'utils/visible';
import { openErrNotification } from 'utils/notifition';

const { DescriptionsItem } = Descriptions;

interface IServiceDetailProps {
    namespace: string;
    serviceName: string;
}

const ServiceDetail: React.FC<IServiceDetailProps> = ({ }) => {
    const dispatch = useAppDispatch();

    const { editSvc, viewSvc } =  useAppSelector(selectService);

    React.useEffect(() => {
        if (editSvc) {
            dispatch(listOneService({
                id: editSvc?.id,
            })).then((res) => {
                if (res.meta.requestStatus === 'rejected') {
                    openErrNotification('请求错误', `获取服务详情失败: ${res.payload as string}`);
                }
            });
        }
    }, [editSvc])

    const svcView = () => {
        const visible = CheckVisibilityMode(viewSvc?.export_to, viewSvc?.namespace || '')

        return (
            <>
                <Descriptions
                    column={1}
                    tableLayout='auto'
                >
                    <DescriptionsItem label="ID">{viewSvc?.id}</DescriptionsItem>
                    <DescriptionsItem label="命名空间">{viewSvc?.namespace}</DescriptionsItem>
                    <DescriptionsItem label="名称">{viewSvc?.name}</DescriptionsItem>
                    <DescriptionsItem label="业务">{viewSvc?.business}</DescriptionsItem>
                    <DescriptionsItem label="部门">{viewSvc?.department}</DescriptionsItem>
                    <DescriptionsItem label="描述">{viewSvc?.comment}</DescriptionsItem>
                    <DescriptionsItem label="服务可见性">{VisibilityModeMap[visible]}</DescriptionsItem>
                    {visible === VisibilityMode_Specified && (
                        <DescriptionsItem label="可见命名空间">
                            <Space>
                                {viewSvc?.export_to ? viewSvc?.export_to.map((ns) => <Tag key={ns}>{ns}</Tag>) : '-'}
                            </Space>
                        </DescriptionsItem>
                    )}
                    <DescriptionsItem label="标签">
                        <Space>
                            {viewSvc?.metadata ? Object.entries(viewSvc?.metadata).map(([key, value]) => (
                            <Tag key={key} theme="primary" style={{ marginRight: '4px' }}>
                                {key}: {value}
                            </Tag>
                        )) : '-'}
                        </Space>
                    </DescriptionsItem>
                </Descriptions>
            </>
        )
    }

    return (
        <>
            {svcView()}
        </>
    )
}

export default React.memo(ServiceDetail);
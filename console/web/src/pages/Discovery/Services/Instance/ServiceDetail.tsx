import React from 'react';
import { Empty, Loading, Space, Tag } from 'tdesign-react';
import { ServiceIcon } from 'tdesign-icons-react';

import Text from 'components/Text';
import { useAppDispatch, useAppSelector } from 'modules/store';
import { listOneService, selectService } from 'modules/discovery/service';
import { openErrNotification } from 'utils/notifition';
import style from './index.module.less';

interface IServiceDetailProps {
    namespace: string;
    serviceName: string;
}

const emptyText = '-';

const isEmpty = (value?: string | null) => value === undefined || value === null || value === '';

const Value = ({ children }: { children?: React.ReactNode }) => (
    <div className={style.detailValue}>{children === undefined || children === null || children === '' ? emptyText : children}</div>
)

const DetailItem = ({ label, children }: { label: string, children?: React.ReactNode }) => (
    <div className={style.detailItem}>
        <div className={style.detailLabel}>{label}</div>
        <Value>{children}</Value>
    </div>
)

const ServiceDetail: React.FC<IServiceDetailProps> = ({ namespace, serviceName }) => {
    const dispatch = useAppDispatch();
    const { editSvc, viewSvc } = useAppSelector(selectService);
    const [loading, setLoading] = React.useState(false);

    React.useEffect(() => {
        if (namespace && serviceName) {
            const id = editSvc?.namespace === namespace && editSvc?.name === serviceName && editSvc?.id !== `${namespace}/${serviceName}`
                ? editSvc?.id
                : undefined;
            setLoading(true);
            dispatch(listOneService({
                id,
                namespace: namespace,
                name: serviceName,
            })).then((res) => {
                if (res.meta.requestStatus === 'rejected') {
                    openErrNotification('请求错误', `获取服务详情失败: ${res.payload as string}`);
                }
            }).finally(() => {
                setLoading(false);
            });
        }
    }, [dispatch, editSvc?.id, editSvc?.name, editSvc?.namespace, namespace, serviceName])

    const service = viewSvc?.namespace === namespace && viewSvc?.name === serviceName ? viewSvc : null;
    const metadata = Object.entries(service?.metadata || {});
    const totalInstances = service?.total_instance_count || '0';
    const healthyInstances = service?.healthy_instance_count || '0';

    if (loading && !service) {
        return (
            <div className={style.detailPanel}>
                <Loading text="加载服务详情中..." />
            </div>
        )
    }

    if (!service) {
        return (
            <div className={style.detailPanel}>
                <Empty title="暂无服务详情" />
            </div>
        )
    }

    return (
        <section className={style.detailPanel}>
            <div className={style.detailHeader}>
                <div className={style.detailIdentity}>
                    <div className={style.detailIcon}>
                        <ServiceIcon />
                    </div>
                    <div className={style.detailTitleGroup}>
                        <div className={style.detailName}>{service.name}</div>
                        <Space size={8}>
                            <Tag variant="light">{service.namespace}</Tag>
                        </Space>
                    </div>
                </div>
                <div className={style.detailStats}>
                    <div className={style.detailStat}>
                        <span>健康实例</span>
                        <strong>{healthyInstances}</strong>
                    </div>
                    <div className={style.detailStat}>
                        <span>总实例</span>
                        <strong>{totalInstances}</strong>
                    </div>
                    <div className={style.detailStat}>
                        <span>标签</span>
                        <strong>{metadata.length}</strong>
                    </div>
                </div>
            </div>

            <div className={style.detailSections}>
                <section className={style.detailSection}>
                    <div className={style.detailSectionTitle}>基础信息</div>
                    <div className={style.detailGrid}>
                        <DetailItem label="ID">{service.id}</DetailItem>
                        <DetailItem label="命名空间">{service.namespace}</DetailItem>
                        <DetailItem label="名称">{service.name}</DetailItem>
                        <DetailItem label="端口">{service.ports}</DetailItem>
                        <DetailItem label="Revision">{service.revision}</DetailItem>
                    </div>
                </section>

                <section className={style.detailSection}>
                    <div className={style.detailSectionTitle}>治理归属</div>
                    <div className={style.detailGrid}>
                        <DetailItem label="业务">{service.business}</DetailItem>
                        <DetailItem label="部门">{service.department}</DetailItem>
                    </div>
                    <div className={style.detailBlock}>
                        <div className={style.detailLabel}>描述</div>
                        <Text>{isEmpty(service.comment) ? emptyText : service.comment}</Text>
                    </div>
                </section>

                <section className={style.detailSection}>
                    <div className={style.detailSectionTitle}>标签</div>
                    {metadata.length > 0 ? (
                        <Space breakLine>
                            {metadata.map(([key, value]) => (
                                <Tag key={key} theme="primary" variant="light">
                                    {key}: {value}
                                </Tag>
                            ))}
                        </Space>
                    ) : (
                        <Text>{emptyText}</Text>
                    )}
                </section>

                <section className={style.detailSection}>
                    <div className={style.detailSectionTitle}>时间</div>
                    <div className={style.detailGrid}>
                        <DetailItem label="创建时间">{service.ctime}</DetailItem>
                        <DetailItem label="修改时间">{service.mtime}</DetailItem>
                    </div>
                </section>
            </div>
        </section>
    )
}

export default React.memo(ServiceDetail);

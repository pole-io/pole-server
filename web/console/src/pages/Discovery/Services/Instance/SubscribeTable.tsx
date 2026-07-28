import React from 'react';
import { Button, Empty, PageInfo, Pagination, Tag, Tooltip } from 'components/Fluent';
import { ArrowRightIcon, RefreshIcon, ServiceIcon } from 'components/Fluent/icons';
import { useNavigate } from 'components/Router';

import Search from 'components/Search';
import Text from 'components/Text';
import { describeServiceSubscribers, ServiceKey, ServiceSubscriberView } from 'services/service';
import { openErrNotification } from 'utils/notifition';
import style from './index.module.less';
import ResourceNameLink from 'components/ResourceNameLink';

interface IServiceSubscribeProps {
    namespace: string
    serviceName: string
}

const serviceLabel = (service?: ServiceKey) => {
    if (!service) {
        return '-'
    }
    return `${service.namespace}/${service.name}`
}

const classNames = (...names: Array<string | false | undefined>) => names.filter(Boolean).join(' ')

const ServiceNode = ({ service, role, active = false, onView }: { service?: ServiceKey, role: string, active?: boolean, onView: (service: ServiceKey) => void }) => (
    <div className={classNames(style.serviceNode, active && style.serviceNodeActive)}>
        <div className={style.serviceNodeIcon}>
            <ServiceIcon />
        </div>
        <div className={style.serviceNodeBody}>
            <div className={style.serviceNodeMeta}>
                <span>{role}</span>
                {service?.namespace && <Tag size="small" variant="light">{service.namespace}</Tag>}
            </div>
            <div className={style.serviceNodeName}>
                <ResourceNameLink name={service?.name} onClick={() => service && onView(service)} />
            </div>
        </div>
    </div>
)

const ServiceSubscribeTable: React.FC<IServiceSubscribeProps> = ({ namespace, serviceName }) => {
    const navigate = useNavigate();
    const [datas, setDatas] = React.useState<ServiceSubscriberView[]>([]);
    const [loading, setLoading] = React.useState(false);
    const [page, setPage] = React.useState(1);
    const [limit, setLimit] = React.useState(10);
    const [total, setTotal] = React.useState(0);
    const [callerName, setCallerName] = React.useState('');

    const currentService = React.useMemo(() => ({ namespace, name: serviceName }), [namespace, serviceName]);
    const viewService = React.useCallback((service: ServiceKey) => {
        navigate(`/discovery/service/instance?namespace=${encodeURIComponent(service.namespace)}&service=${encodeURIComponent(service.name)}`);
    }, [navigate]);

    const refreshTable = React.useCallback(async (nextPage = page, nextLimit = limit, nextCallerName = callerName) => {
        if (!namespace || !serviceName) {
            setDatas([]);
            setTotal(0);
            return;
        }
        setLoading(true);
        try {
            const res = await describeServiceSubscribers({
                callee_namespace: namespace,
                callee_name: serviceName,
                caller_name: nextCallerName || undefined,
                offset: (nextPage - 1) * nextLimit,
                limit: nextLimit,
            });
            setDatas(res.list);
            setTotal(res.totalCount);
            setPage(nextPage);
            setLimit(nextLimit);
            setCallerName(nextCallerName);
        } catch (err) {
            openErrNotification('获取服务订阅列表失败', String((err as Error)?.message || err));
        } finally {
            setLoading(false);
        }
    }, [callerName, limit, namespace, page, serviceName]);

    React.useEffect(() => {
        refreshTable(1, limit, callerName);
    }, [namespace, serviceName]);

    const handlePageChange = (pageInfo: PageInfo) => {
        refreshTable(pageInfo.current, pageInfo.pageSize, callerName);
    }

    return (
        <section className={`${style.subscribePanel} ${style.subscribePanelCompact}`}>
            <div className={style.subscribeHeader}>
                <div className={style.subscribeContext}>
                    <div className={style.subscribeTitle}>服务订阅</div>
                    <div className={style.subscribeScope}>
                        <Tag theme="primary" variant="light">当前服务</Tag>
                        <Text>{serviceLabel(currentService)}</Text>
                    </div>
                </div>
                <div className={style.subscribeStats}>
                    <div className={style.statItem}>
                        <span>订阅方</span>
                        <strong>{total}</strong>
                    </div>
                    <div className={style.statItem}>
                        <span>当前页</span>
                        <strong>{datas.length}</strong>
                    </div>
                </div>
            </div>

            <div className={style.subscribeToolbar}>
                <Search
                    placeholder="搜索订阅方服务"
                    onChange={(value: string) => {
                        refreshTable(1, limit, value);
                    }}
                />
                <Tooltip content="刷新">
                    <Button aria-label="刷新订阅列表" shape="square" variant="text" onClick={() => refreshTable(1, limit, callerName)}>
                        <RefreshIcon />
                    </Button>
                </Tooltip>
            </div>

            <div className={classNames(style.relationList, loading && style.relationListLoading)}>
                {datas.length === 0 && !loading && (
                    <div className={style.subscribeEmpty}>
                        <Empty title="暂无订阅关系" />
                    </div>
                )}
                {loading && datas.length === 0 && (
                    <div className={style.subscribeEmpty}>
                        <Text>加载中...</Text>
                    </div>
                )}
                {datas.map((item) => (
                    <div className={style.relationRow} key={item.id}>
                        <ServiceNode service={item.caller} role="订阅方" onView={viewService} />
                        <div className={style.relationConnector}>
                            <div className={style.connectorLine} />
                            <div className={style.connectorBadge}>
                                <ArrowRightIcon />
                            </div>
                            <span>订阅</span>
                        </div>
                        <div className={style.calleeStack}>
                            {(item.callee?.length ? item.callee : [currentService]).map((callee) => (
                                <ServiceNode
                                    key={serviceLabel(callee)}
                                    service={callee}
                                    role="被订阅"
                                    active={callee.namespace === namespace && callee.name === serviceName}
                                    onView={viewService}
                                />
                            ))}
                        </div>
                    </div>
                ))}
            </div>

            <div className={style.subscribeFooter}>
                <span>共 {total} 条关系</span>
                <Pagination
                    current={page}
                    pageSize={limit}
                    total={total}
                    showJumper
                    onChange={handlePageChange}
                />
            </div>
        </section>
    )
}

export default React.memo(ServiceSubscribeTable);

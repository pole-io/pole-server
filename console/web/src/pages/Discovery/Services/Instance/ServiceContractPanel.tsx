import React from 'react';
import { Button, Empty, Loading, Select, Tag } from 'components/Fluent';
import {
    DescribeGovernanceServiceContracts as describeGovernanceServiceContracts,
    DescribeGovernanceServiceContractVersions as describeGovernanceServiceContractVersions,
    GovernanceInterfaceDescription,
    GovernanceServiceContract,
} from 'services/service';

import style from './ServiceContractPanel.module.less';

interface ServiceContractPanelProps {
    namespace: string
    serviceName: string
}

type ContractView = 'interfaces' | 'raw'

const PROTOCOLS = [
    { key: 'http', label: 'HTTP / OpenAPI', hint: '路径、方法与请求响应模型' },
    { key: 'dubbo', label: 'Dubbo', hint: '接口、方法与类型定义' },
    { key: 'grpc', label: 'gRPC', hint: 'Service、Method 与流模式' },
    { key: 'thrift', label: 'Thrift', hint: 'Service、Function 与 IDL' },
] as const;

const normalizeProtocol = (protocol?: string) => {
    const value = String(protocol || '').trim().toLocaleLowerCase();
    if (value === 'openapi' || value.startsWith('http')) {
        return 'http';
    }
    return value;
}

const protocolLabel = (protocol?: string) => (
    PROTOCOLS.find((item) => item.key === normalizeProtocol(protocol))?.label || protocol || '未知协议'
)

const protocolOperationLabel = (protocol: string | undefined, operation: GovernanceInterfaceDescription) => {
    const normalized = normalizeProtocol(protocol);
    const path = operation.path || '未命名接口';
    const method = operation.method || '未声明方法';
    switch (normalized) {
        case 'http':
            return `${method.toLocaleUpperCase()} ${path}`;
        case 'dubbo':
            return `${path}#${operation.type && operation.type !== 'dubbo' ? operation.type : method}`;
        case 'grpc':
            return `${path} · ${method}`;
        case 'thrift':
            return `${path}.${method}`;
        default:
            return `${path} / ${method}`;
    }
}

const operationFieldLabels = (protocol?: string) => {
    switch (normalizeProtocol(protocol)) {
        case 'http':
            return { path: 'HTTP 路径', method: 'HTTP 方法' };
        case 'grpc':
            return { path: 'gRPC Service', method: 'RPC Method' };
        case 'dubbo':
            return { path: 'Dubbo 接口', method: 'RPC 方法' };
        case 'thrift':
            return { path: 'Thrift Service', method: 'Function' };
        default:
            return { path: '接口', method: '方法' };
    }
}

const contractIdentity = (contract: GovernanceServiceContract, index: number) => (
    contract.id
    || [contract.protocol, contract.name, contract.version].filter(Boolean).join('/')
    || `contract-${index}`
)

const sourceLabel = (source?: string) => {
    if (source === '2' || String(source).toLocaleLowerCase() === 'client') {
        return 'SDK 上报';
    }
    if (source === '1' || String(source).toLocaleLowerCase() === 'manual') {
        return '人工维护';
    }
    return source || '未标注来源';
}

const loadAllContracts = async (namespace: string, serviceName: string) => {
    const all: GovernanceServiceContract[] = [];
    const limit = 100;
    for (let offset = 0; ; offset += limit) {
        const response = await describeGovernanceServiceContracts({
            namespace,
            service: serviceName,
            offset,
            limit,
        });
        const page = response.data ?? [];
        all.push(...page);
        if (page.length < limit || all.length >= (response.amount ?? all.length)) {
            return all;
        }
    }
}

const ServiceContractPanel: React.FC<ServiceContractPanelProps> = ({ namespace, serviceName }) => {
    const [contracts, setContracts] = React.useState<GovernanceServiceContract[]>([]);
    const [versions, setVersions] = React.useState<string[]>([]);
    const [selectedVersion, setSelectedVersion] = React.useState('');
    const [selectedContractId, setSelectedContractId] = React.useState('');
    const [activeView, setActiveView] = React.useState<ContractView>('interfaces');
    const [loading, setLoading] = React.useState(true);
    const [error, setError] = React.useState('');
    const [reloadRevision, setReloadRevision] = React.useState(0);

    React.useEffect(() => {
        let disposed = false;

        const load = async () => {
            if (!namespace || !serviceName) {
                setContracts([]);
                setVersions([]);
                setSelectedVersion('');
                setSelectedContractId('');
                setLoading(false);
                return;
            }

            setLoading(true);
            setError('');
            try {
                const [versionResponse, nextContracts] = await Promise.all([
                    describeGovernanceServiceContractVersions({ namespace, service: serviceName }),
                    loadAllContracts(namespace, serviceName),
                ]);
                if (disposed) {
                    return;
                }

                const nextVersions = Array.from(new Set([
                    ...(versionResponse.data ?? []).map((item) => item.version || ''),
                    ...nextContracts.map((item) => item.version || ''),
                ].filter(Boolean)));
                const nextVersion = nextVersions[0] || '';
                const firstContractIndex = nextContracts.findIndex((item) => !nextVersion || item.version === nextVersion);

                setContracts(nextContracts);
                setVersions(nextVersions);
                setSelectedVersion(nextVersion);
                setSelectedContractId(firstContractIndex >= 0 ? contractIdentity(nextContracts[firstContractIndex], firstContractIndex) : '');
                setActiveView('interfaces');
            } catch (reason) {
                if (!disposed) {
                    setContracts([]);
                    setVersions([]);
                    setSelectedVersion('');
                    setSelectedContractId('');
                    setError(String((reason as Error)?.message || reason || '未知错误'));
                }
            } finally {
                if (!disposed) {
                    setLoading(false);
                }
            }
        }

        load();
        return () => {
            disposed = true;
        }
    }, [namespace, reloadRevision, serviceName]);

    const visibleContracts = React.useMemo(
        () => contracts.filter((item) => !selectedVersion || item.version === selectedVersion),
        [contracts, selectedVersion],
    );

    const selectedContract = React.useMemo(
        () => {
            const exactIndex = contracts.findIndex((item, index) => contractIdentity(item, index) === selectedContractId);
            if (exactIndex >= 0) {
                return contracts[exactIndex];
            }
            return visibleContracts[0];
        },
        [contracts, selectedContractId, visibleContracts],
    );

    const protocolCounts = React.useMemo(() => contracts.reduce<Record<string, number>>((result, contract) => {
        const protocol = normalizeProtocol(contract.protocol);
        result[protocol] = (result[protocol] || 0) + 1;
        return result;
    }, {}), [contracts]);

    const selectVersion = (value: string) => {
        const nextVisible = contracts
            .map((contract, index) => ({ contract, id: contractIdentity(contract, index) }))
            .filter((item) => !value || item.contract.version === value);
        setSelectedVersion(value);
        setSelectedContractId(nextVisible[0]?.id || '');
        setActiveView('interfaces');
    }

    if (loading) {
        return (
            <section className={style.state} aria-live="polite">
                <Loading text="加载服务契约中..." />
            </section>
        );
    }

    if (error) {
        return (
            <section className={style.state} role="alert">
                <strong>服务契约加载失败</strong>
                <span>{error}</span>
                <Button variant="outline" onClick={() => setReloadRevision((value) => value + 1)}>重新加载</Button>
            </section>
        );
    }

    if (contracts.length === 0) {
        return (
            <section className={style.state}>
                <Empty
                    title="暂无服务契约"
                    description="可通过 CI 上传或 SDK 上报 OpenAPI、Dubbo、gRPC、Thrift 契约。"
                />
            </section>
        );
    }

    const fields = operationFieldLabels(selectedContract?.protocol);
    const interfaces = selectedContract?.interfaces ?? [];

    return (
        <section className={style.panel} aria-label={`${namespace}/${serviceName} 服务契约`}>
            <header className={style.header}>
                <div>
                    <span className={style.eyebrow}>SERVICE CONTRACT</span>
                    <h2>服务契约</h2>
                    <p>统一查看 HTTP、Dubbo、gRPC 与 Thrift 的接口定义和原始描述。</p>
                </div>
                <div className={style.headerMeta}>
                    <span>契约</span>
                    <strong>{contracts.length}</strong>
                </div>
            </header>

            <div className={style.protocolGrid} aria-label="服务契约协议能力">
                {PROTOCOLS.map((protocol) => {
                    const count = protocolCounts[protocol.key] || 0;
                    return (
                        <div className={style.protocolCard} key={protocol.key}>
                            <div className={style.protocolCardHeader}>
                                <Tag variant={count ? 'outline' : 'light'}>{protocol.label}</Tag>
                                <strong>{count}</strong>
                            </div>
                            <span>{protocol.hint}</span>
                        </div>
                    )
                })}
            </div>

            <div className={style.selectorBar}>
                <label>
                    <span>契约版本</span>
                    <Select
                        label="契约版本"
                        value={selectedVersion}
                        options={versions.map((version) => ({ label: version, value: version }))}
                        onChange={(value: string) => selectVersion(value)}
                        placeholder="选择版本"
                    />
                </label>
                <label>
                    <span>契约定义</span>
                    <Select
                        label="契约定义"
                        value={selectedContract ? contracts
                            .map((contract, index) => ({ contract, id: contractIdentity(contract, index) }))
                            .find((item) => item.contract === selectedContract)?.id : ''}
                        options={visibleContracts.map((contract) => {
                            const originalIndex = contracts.indexOf(contract);
                            return {
                                label: `${protocolLabel(contract.protocol)} · ${contract.type || contract.name || '未命名契约'}`,
                                value: contractIdentity(contract, originalIndex),
                            }
                        })}
                        onChange={(value: string) => {
                            setSelectedContractId(value);
                            setActiveView('interfaces');
                        }}
                        placeholder="选择契约"
                    />
                </label>
                <dl className={style.contractMeta}>
                    <div>
                        <dt>协议</dt>
                        <dd>{protocolLabel(selectedContract?.protocol)}</dd>
                    </div>
                    <div>
                        <dt>修订</dt>
                        <dd title={selectedContract?.revision}>{selectedContract?.revision || '-'}</dd>
                    </div>
                    <div>
                        <dt>状态</dt>
                        <dd>{selectedContract?.status === 'Online' ? '在线' : '离线'}</dd>
                    </div>
                </dl>
            </div>

            <div className={style.viewTabs} role="tablist" aria-label="契约展示方式">
                <Button
                    role="tab"
                    aria-selected={activeView === 'interfaces'}
                    theme={activeView === 'interfaces' ? 'primary' : 'default'}
                    variant={activeView === 'interfaces' ? 'base' : 'text'}
                    onClick={() => setActiveView('interfaces')}
                >
                    接口清单
                </Button>
                <Button
                    role="tab"
                    aria-selected={activeView === 'raw'}
                    theme={activeView === 'raw' ? 'primary' : 'default'}
                    variant={activeView === 'raw' ? 'base' : 'text'}
                    onClick={() => setActiveView('raw')}
                >
                    原始契约
                </Button>
            </div>

            {activeView === 'interfaces' ? (
                <div className={style.interfaceList} role="tabpanel" aria-label="接口清单">
                    {interfaces.length === 0 ? (
                        <Empty title="暂无接口定义" description="当前契约没有可展示的归一化接口。" />
                    ) : interfaces.map((operation, index) => (
                        <article className={style.interfaceRow} key={operation.id || `${operation.path}-${operation.method}-${index}`}>
                            <div className={style.operationIdentity}>
                                <strong>{protocolOperationLabel(selectedContract?.protocol, operation)}</strong>
                                <span>{operation.content || '暂无接口说明'}</span>
                            </div>
                            <dl className={style.operationFields}>
                                <div>
                                    <dt>{fields.path}</dt>
                                    <dd>{operation.path || '-'}</dd>
                                </div>
                                <div>
                                    <dt>{fields.method}</dt>
                                    <dd>{operation.method || '-'}</dd>
                                </div>
                                <div>
                                    <dt>来源</dt>
                                    <dd>{sourceLabel(operation.source)}</dd>
                                </div>
                            </dl>
                        </article>
                    ))}
                </div>
            ) : (
                <div className={style.rawPanel} role="tabpanel" aria-label="原始契约">
                    <div className={style.rawHeader}>
                        <span>原始内容</span>
                        <Tag variant="outline">{protocolLabel(selectedContract?.protocol)}</Tag>
                    </div>
                    <pre className={style.rawContent}>{selectedContract?.content || '当前契约未保存原始内容。'}</pre>
                </div>
            )}
        </section>
    )
}

export default React.memo(ServiceContractPanel);

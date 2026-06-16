import React from 'react';
import { Button, Form, Input, InputNumber, Select, Space, StickyTool, Switch, Tag, Textarea } from 'tdesign-react';
import type { FormProps } from 'tdesign-react';
import { AddIcon, CheckIcon, ChevronRightIcon, CloseIcon, Edit1Icon, RocketIcon, RollbackIcon, SaveIcon } from 'tdesign-icons-react';

import { cleanNamespacePage, listAllNamespaces, selectNamespace } from 'modules/namespace';
import { cleanServicePage, listAllServices, selectService } from 'modules/discovery/service';
import { useAppDispatch, useAppSelector } from 'modules/store';
import { NamespaceView } from 'services/namespace';
import { ServiceView } from 'services/service';
import { PolicySourceType } from 'services/auth_policy';
import {
    InterfaceProtocol,
    MatchLogic,
    MatchType,
    MatchTypeOption,
    MatchValueType,
    MatchValueTypeOption,
    Op,
} from 'services/types';
import {
    createTrafficGovernanceRules,
    defaultTrafficGovernanceRule,
    defaultTrafficMirrorRule,
    defaultTrafficMockRule,
    defaultTrafficSecurityRule,
    describeOneTrafficGovernanceRule,
    MirrorRule,
    MockRule,
    modifyTrafficGovernanceRules,
    TrafficApiScope,
    TrafficGovernanceAuthResource,
    TrafficGovernanceKind,
    TrafficGovernanceKindLabel,
    TrafficGovernanceReleaseResource,
    TrafficMatchRule,
    TrafficMirror,
    TrafficMock,
    TrafficGovernanceRule,
    TrafficSecurityAction,
    TrafficSecurityActionMap,
    TrafficSecurityPolicy,
    TrafficSecurityRule,
} from 'services/traffic_governance';
import { openErrNotification, openInfoNotification } from 'utils/notifition';
import PublishForm from '../RuleRelease/PublishForm';
import RuleStickyAction from '../RuleRelease/RuleStickyAction';
import RuleLabelField from '../shared/RuleLabelField';
import shared from '../shared/governance.module.less';
import style from './index.module.less';

const { FormItem } = Form;
const { StickyItem } = StickyTool;

const matchSourceOptions = ['HEADER', 'QUERY', 'COOKIE', 'METHOD', 'PATH', 'CALLER_SERVICE', 'CALLER_IP'].map((item) => ({ label: item, value: item }));
const methodOptions = ['GET', 'POST', 'PUT', 'DELETE', 'PATCH', 'HEAD', 'OPTIONS', 'ANY'].map((item) => ({ label: item, value: item }));
const protocolOptions = [
    { label: 'HTTP', value: InterfaceProtocol.HTTP },
    { label: 'GRPC', value: InterfaceProtocol.GRPC },
];
const matchModeOptions = [
    { label: 'AND', value: MatchLogic.AND },
    { label: 'OR', value: MatchLogic.OR },
];

interface TrafficGovernanceEditorProps {
    kind: TrafficGovernanceKind;
    op: Op;
    data?: TrafficGovernanceRule;
    visible: boolean;
    refresh: (close: boolean) => void;
}

const readAction = (value?: string | number) => {
    if (value === 1 || value === '1' || value === 'TRAFFIC_SECURITY_DENY' || value === 'DENY') {
        return TrafficSecurityAction.DENY;
    }
    return TrafficSecurityAction.ALLOW;
};

const readonlyItem = (label: string, value: React.ReactNode, full = false) => (
    <div className={`${shared.field}${full ? ` ${shared.full}` : ''}`}>
        <div className={shared.fieldLabel}>{label}</div>
        <div className={shared.fieldValue}>{value || '-'}</div>
    </div>
);

const text = (value?: string | number | boolean) => {
    if (value === undefined || value === null || value === '') return '-';
    return String(value);
};

const matchText = (value?: { type?: string; value?: string; value_type?: string }) => {
    if (!value) return '-';
    return [value.type, value.value, value.value_type].filter(Boolean).join(' / ') || '-';
};

const apiText = (api?: TrafficApiScope) => {
    if (!api) return '-';
    return `${text(api.protocol)} ${text(api.method)} ${matchText(api.path)}`;
};

const durationText = (value?: string | { seconds?: number | string; nanos?: number }) => {
    if (!value) return '-';
    if (typeof value === 'string') return value;
    return `${value.seconds ?? 0}s`;
};

const renderMatchRule = (match?: TrafficMatchRule) => {
    const args = match?.arguments || [];
    if (!args.length && match?.randomPercent === undefined && !match?.matchMode) {
        return <span>-</span>;
    }
    return (
        <div className={style.conditionStack}>
            <div className={style.conditionMeta}>
                <Tag variant="light-outline">{text(match?.matchMode || 'AND')}</Tag>
                <Tag variant="light-outline">命中比例 {match?.randomPercent ?? 100}%</Tag>
            </div>
            {args.length ? args.map((arg, index) => (
                <div className={shared.tagRow} key={`${arg.type}-${arg.key}-${index}`}>
                    <span className={`${shared.tagPlain} ${shared.tagMethod}`}>{text(arg.type)}</span>
                    <span className={shared.chip}>
                        <span className={shared.chipKey}>{text(arg.key)}</span>
                        <span className={shared.chipValue}>{matchText(arg.value)}</span>
                    </span>
                </div>
            )) : <div className={shared.emptyLine}>全部流量</div>}
        </div>
    );
};

const renderRecord = (record?: Record<string, string | { type?: string; value?: string; value_type?: string }>) => {
    const entries = Object.entries(record || {});
    if (!entries.length) return <span>-</span>;
    return (
        <div className={shared.labelDisplay}>
            {entries.map(([key, value]) => (
                <span className={shared.chip} key={key}>
                    <span className={shared.chipKey}>{key}</span>
                    <span className={shared.chipValue}>{typeof value === 'string' ? value : matchText(value)}</span>
                </span>
            ))}
        </div>
    );
};

const rejectEffectText = (policy: TrafficSecurityPolicy) => {
    const effect = policy.reject_effect;
    if (!effect) return '-';
    return [effect.status_code, effect.code, effect.message].filter(Boolean).join(' / ') || '-';
};

const securityPolicySummary = (policy: TrafficSecurityPolicy) => {
    const matchCount = policy.traffic_match_rule?.arguments?.length || 0;
    const action = TrafficSecurityActionMap[String(policy.action)] || '-';
    const api = apiText(policy.api);
    const rejectEffect = rejectEffectText(policy);
    return `${matchCount} 个匹配条件 / ${api} / ${action}${readAction(policy.action) === TrafficSecurityAction.DENY ? ` / ${rejectEffect}` : ''}`;
};

const renderApiScope = (api?: TrafficApiScope) => {
    if (!api) return <span className={shared.emptyLine}>全部接口</span>;
    return (
        <div className={shared.tagRow}>
            {api.protocol && <span className={shared.tagPlain}>{api.protocol}</span>}
            {api.method && <span className={`${shared.tagPlain} ${shared.tagMethod}`}>{api.method}</span>}
            {api.path?.type && <span className={shared.tagPlain}>{api.path.type}</span>}
            <span className={shared.pathTag}>{api.path?.value || '/'}</span>
            {api.path?.value_type && <span className={shared.tagPlain}>{api.path.value_type}</span>}
        </div>
    );
};

const renderVerdictBlock = (policy: TrafficSecurityPolicy) => {
    const isDeny = readAction(policy.action) === TrafficSecurityAction.DENY;
    const effect = policy.reject_effect;
    return (
        <div className={`${shared.verdict} ${isDeny ? shared.verdictDeny : shared.verdictAllow}`}>
            <div className={shared.verdictIcon}>{isDeny ? <CloseIcon /> : <CheckIcon />}</div>
            <div>
                <div className={shared.verdictHead}>{TrafficSecurityActionMap[String(policy.action)] || (isDeny ? '拒绝' : '放通')}</div>
                <div className={shared.verdictDesc}>
                    {isDeny ? '命中该策略的请求将被拦截，并返回下方响应信息' : '命中该策略的请求允许继续访问目标接口'}
                </div>
            </div>
            {isDeny && (
                <div className={shared.verdictMeta}>
                    <div>
                        <div className={shared.verdictMetaLabel}>状态码</div>
                        <div className={`${shared.verdictMetaValue} ${shared.mono}`}>{text(effect?.status_code ?? 403)}</div>
                    </div>
                    <div>
                        <div className={shared.verdictMetaLabel}>Code</div>
                        <div className={`${shared.verdictMetaValue} ${shared.mono}`}>{text(effect?.code)}</div>
                    </div>
                    <div>
                        <div className={shared.verdictMetaLabel}>文案</div>
                        <div className={shared.verdictMetaValue}>{text(effect?.message)}</div>
                    </div>
                </div>
            )}
        </div>
    );
};

const defaultMatchValue = () => ({
    type: MatchType.EXACT,
    value: '',
    value_type: MatchValueType.TEXT,
});

const defaultArgument = () => ({
    type: 'HEADER',
    key: '',
    value: defaultMatchValue(),
});

const normalizeMatchRule = (match?: TrafficMatchRule): TrafficMatchRule => ({
    matchMode: match?.matchMode || MatchLogic.AND,
    randomPercent: match?.randomPercent ?? 100,
    arguments: match?.arguments?.length ? match.arguments : [defaultArgument()],
});

export const trafficRuleCount = (kind: TrafficGovernanceKind, rule?: TrafficGovernanceRule) => {
    if (!rule) return 0;
    if (kind === 'security') return ((rule as TrafficSecurityRule).policies || []).length;
    return ((rule as any).rules || []).length;
};

export const trafficRuleSummary = (kind: TrafficGovernanceKind, rule?: TrafficGovernanceRule) => {
    if (!rule) return '-';
    if (kind === 'security') {
        const security = rule as TrafficSecurityRule;
        const first = security.policies?.[0];
        const action = TrafficSecurityActionMap[String(first?.action || '')] || '-';
        return `${action} / 默认${TrafficSecurityActionMap[String(security.default_action)] || '-'}`;
    }
    if (kind === 'mirror') {
        const first = (rule as any).rules?.[0];
        return first ? `${first.source?.namespace || '-'}/${first.source?.service || '-'} -> ${first.destination?.namespace || '-'}/${first.destination?.service || '-'} / ${first.mirror_percent ?? 0}%` : '-';
    }
    const first = (rule as any).rules?.[0];
    return first ? `${first.source?.namespace || '-'}/${first.source?.service || '-'} / ${first.response?.status_code || 200} / ${first.mock_percent ?? 0}%` : '-';
};

const TrafficGovernanceEditor: React.FC<TrafficGovernanceEditorProps> = ({ kind, op, data, visible, refresh }) => {
    const [form] = Form.useForm();
    const dispatch = useAppDispatch();
    const namespaceState = useAppSelector(selectNamespace);
    const serviceState = useAppSelector(selectService);
    const [rule, setRule] = React.useState<TrafficGovernanceRule>(() => defaultTrafficGovernanceRule(kind));
    const [editable, setEditable] = React.useState(op === 'create');
    const [publishVisible, setPublishVisible] = React.useState(false);
    const [collapsedSecurityPolicyIndexes, setCollapsedSecurityPolicyIndexes] = React.useState<Set<number>>(() => new Set());

    React.useEffect(() => {
        dispatch(listAllNamespaces());
        dispatch(listAllServices());
        return () => {
            dispatch(cleanNamespacePage());
            dispatch(cleanServicePage());
        };
    }, [dispatch]);

    React.useEffect(() => {
        if (!visible) return;
        setEditable(op === 'create');
        setCollapsedSecurityPolicyIndexes(new Set());
        const load = async () => {
            const next = data?.id && op !== 'create'
                ? await describeOneTrafficGovernanceRule(kind, data.id)
                : (data || defaultTrafficGovernanceRule(kind));
            setRule(next);
            form.setFieldsValue({
                name: next.name,
                namespace: next.namespace,
                service: next.service,
                description: next.description,
                priority: next.priority ?? 0,
                enable: next.enable ?? true,
                default_action: kind === 'security' ? readAction((next as TrafficSecurityRule).default_action) : undefined,
            });
        };
        load().catch((error) => {
            openErrNotification('请求失败', `获取${TrafficGovernanceKindLabel[kind]}详情失败: ${(error as Error).message}`);
        });
    }, [data, form, kind, op, visible]);

    const namespaceOptions = (namespaceState.datas || []).map((ns: NamespaceView) => ({ label: ns.name, value: ns.name }));
    const serviceOptions = (serviceState.datas || [])
        .filter((svc: ServiceView) => !rule.namespace || rule.namespace === '*' || svc.namespace === rule.namespace)
        .map((svc: ServiceView) => ({ label: svc.name, value: svc.name }));

    const setSecurityPolicy = (index: number, updater: (policy: TrafficSecurityPolicy) => TrafficSecurityPolicy) => {
        setRule((prev) => {
            const security = prev as TrafficSecurityRule;
            const policies = [...(security.policies || [])];
            policies[index] = updater(policies[index] || defaultTrafficSecurityRule().policies[0]);
            return { ...security, policies };
        });
    };

    const toggleSecurityPolicyCollapsed = (index: number) => {
        setCollapsedSecurityPolicyIndexes((prev) => {
            const next = new Set(prev);
            if (next.has(index)) {
                next.delete(index);
            } else {
                next.add(index);
            }
            return next;
        });
    };

    const setMirrorRule = (index: number, updater: (item: MirrorRule) => MirrorRule) => {
        setRule((prev) => {
            const mirror = prev as TrafficMirror;
            const rules = [...(mirror.rules || [])];
            rules[index] = updater(rules[index] || defaultTrafficMirrorRule().rules[0]);
            return { ...mirror, rules };
        });
    };

    const setMockRule = (index: number, updater: (item: MockRule) => MockRule) => {
        setRule((prev) => {
            const mock = prev as TrafficMock;
            const rules = [...(mock.rules || [])];
            rules[index] = updater(rules[index] || defaultTrafficMockRule().rules[0]);
            return { ...mock, rules };
        });
    };

    // 匹配值的三段控件（匹配方式 / 值 / 值类型），以 fragment 返回，由父行栅格统一排成一行
    const renderMatchValueEditor = (
        value: { type?: string; value?: string; value_type?: string } | undefined,
        onChange: (next: { type?: string; value?: string; value_type?: string }) => void,
    ) => {
        const current = value || defaultMatchValue();
        return (
            <>
                <Select
                    options={MatchTypeOption}
                    value={current.type || MatchType.EXACT}
                    onChange={(next) => onChange({ ...current, type: next as string })}
                />
                <Input
                    value={current.value || ''}
                    placeholder="匹配值"
                    onChange={(next) => onChange({ ...current, value: next })}
                />
                <Select
                    options={MatchValueTypeOption}
                    value={current.value_type || MatchValueType.TEXT}
                    onChange={(next) => onChange({ ...current, value_type: next as string })}
                />
            </>
        );
    };

    const renderMatchRuleEditor = (match: TrafficMatchRule | undefined, onChange: (next: TrafficMatchRule) => void) => {
        const current = normalizeMatchRule(match);
        const args = current.arguments || [];
        const updateArg = (index: number, updater: (arg: NonNullable<TrafficMatchRule['arguments']>[number]) => NonNullable<TrafficMatchRule['arguments']>[number]) => {
            const nextArgs = [...args];
            nextArgs[index] = updater(nextArgs[index] || defaultArgument());
            onChange({ ...current, arguments: nextArgs });
        };
        return (
            <div className={style.repeatEditor}>
                <div className={style.repeatMeta}>
                    <span className={style.repeatMetaHint}>匹配关系</span>
                    <Select
                        className={style.metaSelect}
                        options={matchModeOptions}
                        value={current.matchMode || MatchLogic.AND}
                        onChange={(next) => onChange({ ...current, matchMode: next as string })}
                    />
                    <span className={style.repeatMetaHint}>命中比例</span>
                    <InputNumber
                        className={style.metaPercent}
                        min={0}
                        max={100}
                        suffix="%"
                        value={current.randomPercent ?? 100}
                        onChange={(next) => onChange({ ...current, randomPercent: Number(next ?? 0) })}
                    />
                </div>
                {args.map((arg, index) => (
                    <div className={style.conditionRow} key={`${arg.type}-${arg.key}-${index}`}>
                        <Select
                            filterable
                            creatable
                            options={matchSourceOptions}
                            value={arg.type || 'HEADER'}
                            onChange={(next) => updateArg(index, (item) => ({ ...item, type: next as string }))}
                        />
                        <Input
                            value={arg.key || ''}
                            placeholder="匹配 Key"
                            onChange={(next) => updateArg(index, (item) => ({ ...item, key: next }))}
                        />
                        {renderMatchValueEditor(arg.value, (next) => updateArg(index, (item) => ({ ...item, value: next })))}
                        <Button
                            shape="circle"
                            variant="text"
                            disabled={args.length <= 1}
                            onClick={() => onChange({ ...current, arguments: args.filter((_, idx) => idx !== index) })}
                        >
                            <CloseIcon />
                        </Button>
                    </div>
                ))}
                <Button
                    className={shared.addRuleButton}
                    variant="dashed"
                    icon={<AddIcon />}
                    onClick={() => onChange({ ...current, arguments: [...args, defaultArgument()] })}
                >
                    添加匹配条件
                </Button>
            </div>
        );
    };

    const renderApiEditor = (api: TrafficApiScope | undefined, onChange: (next: TrafficApiScope) => void) => {
        const current = api || {};
        return (
            <div className={style.apiEditor}>
                <Select
                    options={protocolOptions}
                    value={current.protocol || InterfaceProtocol.HTTP}
                    onChange={(next) => onChange({ ...current, protocol: next as string })}
                />
                <Select
                    filterable
                    creatable
                    options={methodOptions}
                    value={current.method || 'GET'}
                    onChange={(next) => onChange({ ...current, method: next as string })}
                />
                {renderMatchValueEditor(current.path, (next) => onChange({ ...current, path: next }))}
            </div>
        );
    };

    const renderStringRecordEditor = (record: Record<string, string> | undefined, onChange: (next: Record<string, string>) => void, addLabel: string) => {
        const rows = Object.entries(record || {}).map(([key, value]) => ({ key, value }));
        const commitRows = (nextRows: Array<{ key: string; value: string }>) => {
            onChange(nextRows.reduce<Record<string, string>>((acc, row) => {
                if (row.key.trim()) acc[row.key.trim()] = row.value;
                return acc;
            }, {}));
        };
        return (
            <div className={style.repeatEditor}>
                {rows.map((row, index) => (
                    <div className={style.kvRow} key={`${row.key}-${index}`}>
                        <Input value={row.key} placeholder="Key" onChange={(next) => commitRows(rows.map((item, idx) => idx === index ? { ...item, key: next } : item))} />
                        <Input value={row.value} placeholder="Value" onChange={(next) => commitRows(rows.map((item, idx) => idx === index ? { ...item, value: next } : item))} />
                        <Button shape="circle" variant="text" onClick={() => commitRows(rows.filter((_, idx) => idx !== index))}><CloseIcon /></Button>
                    </div>
                ))}
                <Button className={shared.addRuleButton} variant="dashed" icon={<AddIcon />} onClick={() => commitRows([...rows, { key: '', value: '' }])}>{addLabel}</Button>
            </div>
        );
    };

    const renderMatchRecordEditor = (
        record: Record<string, { type?: string; value?: string; value_type?: string }> | undefined,
        onChange: (next: Record<string, { type?: string; value?: string; value_type?: string }>) => void,
    ) => {
        const rows = Object.entries(record || {}).map(([key, value]) => ({ key, value }));
        const commitRows = (nextRows: Array<{ key: string; value: { type?: string; value?: string; value_type?: string } }>) => {
            onChange(nextRows.reduce<Record<string, { type?: string; value?: string; value_type?: string }>>((acc, row) => {
                if (row.key.trim()) acc[row.key.trim()] = row.value;
                return acc;
            }, {}));
        };
        return (
            <div className={style.repeatEditor}>
                {rows.map((row, index) => (
                    <div className={style.labelRow} key={`${row.key}-${index}`}>
                        <Input
                            value={row.key}
                            placeholder="标签 Key"
                            onChange={(next) => commitRows(rows.map((item, idx) => idx === index ? { ...item, key: next } : item))}
                        />
                        {renderMatchValueEditor(row.value, (next) => commitRows(rows.map((item, idx) => idx === index ? { ...item, value: next } : item)))}
                        <Button shape="circle" variant="text" onClick={() => commitRows(rows.filter((_, idx) => idx !== index))}><CloseIcon /></Button>
                    </div>
                ))}
                <Button className={shared.addRuleButton} variant="dashed" icon={<AddIcon />} onClick={() => commitRows([...rows, { key: '', value: defaultMatchValue() }])}>添加标签</Button>
            </div>
        );
    };

    const buildRequest = (): TrafficGovernanceRule => {
        const values = form.getFieldsValue(true) as Partial<TrafficGovernanceRule & TrafficSecurityRule>;
        const base = {
            ...rule,
            ...values,
            metadata: rule.metadata || {},
            enable: values.enable ?? true,
            priority: Number(values.priority ?? 0),
        } as TrafficGovernanceRule;
        if (kind === 'security') {
            return {
                ...base,
                default_action: readAction(values.default_action),
                policies: (rule as TrafficSecurityRule).policies || [],
            } as TrafficSecurityRule;
        }
        return {
            ...base,
            rules: ((rule as TrafficMirror | TrafficMock).rules || []),
        } as TrafficGovernanceRule;
    };

    const onSubmit: FormProps['onSubmit'] = async (e) => {
        if (e.validateResult !== true) return;
        const req = buildRequest();
        try {
            if (op === 'create') {
                await createTrafficGovernanceRules(kind, [req]);
                openInfoNotification('请求成功', `创建${TrafficGovernanceKindLabel[kind]}规则成功`);
                refresh(true);
            } else {
                await modifyTrafficGovernanceRules(kind, [req]);
                openInfoNotification('请求成功', `更新${TrafficGovernanceKindLabel[kind]}规则成功`);
                setEditable(false);
                refresh(false);
            }
        } catch (error) {
            openErrNotification('请求失败', `${op === 'create' ? '创建' : '更新'}${TrafficGovernanceKindLabel[kind]}规则失败: ${(error as Error).message}`);
        }
    };

    const renderCommonFields = () => {
        if (!editable) {
            return (
                <div className={shared.infoGrid}>
                    {readonlyItem('规则名称', rule.name)}
                    {readonlyItem('启用状态', <span className={`${shared.pill} ${rule.enable ? shared.pillOk : shared.pillOff} ${shared.pillDot}`}>{rule.enable ? '启用' : '禁用'}</span>)}
                    {readonlyItem('优先级', rule.priority ?? 0)}
                    {readonlyItem('命名空间', rule.namespace)}
                    {readonlyItem('服务名称', rule.service)}
                    {readonlyItem('Revision', rule.revision)}
                    {readonlyItem('描述', rule.description, true)}
                    {readonlyItem('规则标签', <RuleLabelField metadata={rule.metadata} />, true)}
                </div>
            );
        }
        return (
            <div className={shared.infoGrid}>
                <FormItem className={shared.field} label="规则名称" name="name" rules={[{ required: true, message: '规则名称不能为空' }]}>
                    <Input maxlength={128} />
                </FormItem>
                <FormItem className={shared.field} label="启用状态" name="enable">
                    <Switch label={['启用', '禁用']} />
                </FormItem>
                <FormItem className={shared.field} label="优先级" name="priority">
                    <InputNumber min={0} max={1000000} />
                </FormItem>
                <FormItem className={shared.field} label="命名空间" name="namespace" rules={[{ required: true, message: '命名空间不能为空' }]}>
                    <Select filterable creatable options={namespaceOptions} onChange={(value) => setRule((prev) => ({ ...prev, namespace: value as string, service: '' }))} />
                </FormItem>
                <FormItem className={shared.field} label="服务名称" name="service" rules={[{ required: true, message: '服务名称不能为空' }]}>
                    <Select filterable creatable options={serviceOptions} />
                </FormItem>
                <div className={`${shared.field} ${shared.full}`}>
                    <div className={shared.fieldLabel}>规则标签</div>
                    <RuleLabelField
                        metadata={rule.metadata}
                        editable
                        onChange={(next) => setRule((prev) => ({ ...prev, metadata: next } as TrafficGovernanceRule))}
                    />
                </div>
                <FormItem className={`${shared.field} ${shared.full}`} label="描述" name="description">
                    <Textarea autosize={{ minRows: 2, maxRows: 4 }} maxlength={512} />
                </FormItem>
            </div>
        );
    };

    const renderSecurityRules = () => {
        const security = rule as TrafficSecurityRule;
        const policies = security.policies || [];
        return (
            <>
                <div className={style.securityDefaultBar}>
                    <div>
                        <div className={style.ruleSectionTitle}>未命中默认动作</div>
                        <div className={style.ruleHelp}>所有策略都未命中时使用该动作，决定请求是否继续放通。</div>
                    </div>
                    <Tag theme={readAction(security.default_action) === TrafficSecurityAction.ALLOW ? 'success' : 'danger'} variant="light-outline">
                        {TrafficSecurityActionMap[String(security.default_action)] || '-'}
                    </Tag>
                </div>
                <div className={shared.ruleList}>
                    {policies.length ? policies.map((policy, index) => (
                        <div className={`${shared.policy} ${collapsedSecurityPolicyIndexes.has(index) ? shared.policyCollapsed : ''}`} key={`${policy.action}-${index}`}>
                            <div className={shared.policyHead} onClick={() => toggleSecurityPolicyCollapsed(index)}>
                                <div className={shared.policyHeadMain}>
                                    <span className={shared.caret}><ChevronRightIcon /></span>
                                    <div>
                                        <div className={shared.policyIndex}>规则 [{index + 1}]</div>
                                        {collapsedSecurityPolicyIndexes.has(index) && (
                                            <div className={shared.policySummary}>{securityPolicySummary(policy)}</div>
                                        )}
                                    </div>
                                </div>
                                <div className={shared.policyHeadActions} onClick={(e) => e.stopPropagation()}>
                                    <Tag theme={readAction(policy.action) === TrafficSecurityAction.ALLOW ? 'success' : 'danger'} variant="light-outline">
                                        {TrafficSecurityActionMap[String(policy.action)] || '-'}
                                    </Tag>
                                </div>
                            </div>
                            <div className={shared.policyBody}>
                                <div className={shared.step} data-step="1">
                                    <div className={shared.stepTitle}>接口范围</div>
                                    <div className={shared.stepContent}>{renderApiScope(policy.api)}</div>
                                </div>
                                <div className={shared.step} data-step="2">
                                    <div className={shared.stepTitle}>匹配条件</div>
                                    <div className={shared.stepContent}>{renderMatchRule(policy.traffic_match_rule)}</div>
                                </div>
                                <div className={shared.step} data-step="3">
                                    <div className={shared.stepTitle}>鉴权结果</div>
                                    <div className={shared.stepContent}>{renderVerdictBlock(policy)}</div>
                                </div>
                            </div>
                        </div>
                    )) : <div className={shared.emptyLine}>暂无策略</div>}
                </div>
            </>
        );
    };

    const renderMirrorRules = () => {
        const mirror = rule as TrafficMirror;
        const rules = mirror.rules || [];
        return (
            <div className={shared.ruleList}>
                {rules.length ? rules.map((item: MirrorRule, index) => (
                    <div className={`${shared.group} ${item.disable ? shared.groupDisabled : ''}`} key={`${item.source?.service}-${item.destination?.service}-${index}`}>
                        <div className={shared.groupHead}>
                            <span className={shared.groupTitle}><span className={shared.groupIndex}>{index + 1}</span>镜像规则</span>
                            <Tag theme={item.disable ? 'default' : 'success'} variant="light-outline">{item.disable ? '禁用' : '启用'}</Tag>
                        </div>
                        <div className={shared.groupBody}>
                            <div className={shared.infoGrid}>
                                {readonlyItem('镜像接口', renderApiScope(item.source?.api), true)}
                                {readonlyItem('来源服务', `${text(item.source?.namespace)}/${text(item.source?.service)}`)}
                                {readonlyItem('请求标签匹配', renderMatchRule(item.source?.traffic_match_rule), true)}
                                {readonlyItem('镜像比例', `${item.mirror_percent ?? 0}%`)}
                                {readonlyItem('生效时长', durationText(item.duration))}
                                {readonlyItem('镜像目标', `${text(item.destination?.namespace)}/${text(item.destination?.service)}`)}
                                {readonlyItem('目标实例标签', renderRecord(item.destination?.labels), true)}
                            </div>
                        </div>
                    </div>
                )) : <div className={shared.emptyLine}>暂无规则</div>}
            </div>
        );
    };

    const renderMockRules = () => {
        const mock = rule as TrafficMock;
        const rules = mock.rules || [];
        return (
            <div className={shared.ruleList}>
                {rules.length ? rules.map((item: MockRule, index) => (
                    <div className={`${shared.group} ${item.disable ? shared.groupDisabled : ''}`} key={`${item.source?.service}-${item.source?.api?.method}-${index}`}>
                        <div className={shared.groupHead}>
                            <span className={shared.groupTitle}><span className={shared.groupIndex}>{index + 1}</span>Mock 规则</span>
                            <Tag theme={item.disable ? 'default' : 'success'} variant="light-outline">{item.disable ? '禁用' : '启用'}</Tag>
                        </div>
                        <div className={shared.groupBody}>
                            <div className={shared.infoGrid}>
                                {readonlyItem('来源服务', `${text(item.source?.namespace)}/${text(item.source?.service)}`)}
                                {readonlyItem('接口范围', renderApiScope(item.source?.api), true)}
                                {readonlyItem('Mock 比例', `${item.mock_percent ?? 0}%`)}
                                {readonlyItem('延迟', durationText(item.delay))}
                                {readonlyItem('响应状态', `${item.response?.status_code ?? 200} / ${text(item.response?.code)} / ${text(item.response?.message)}`)}
                                {readonlyItem('响应头', renderRecord(item.response?.headers), true)}
                                {readonlyItem('匹配条件', renderMatchRule(item.source?.traffic_match_rule), true)}
                                {readonlyItem('响应体', <pre className={style.bodyView}>{text(item.response?.body)}</pre>, true)}
                            </div>
                        </div>
                    </div>
                )) : <div className={shared.emptyLine}>暂无规则</div>}
            </div>
        );
    };

    const renderEditSecurityRules = () => {
        const security = rule as TrafficSecurityRule;
        const policies = security.policies || [];
        return (
            <>
                <div className={style.securityDefaultBar}>
                    <div>
                        <div className={style.ruleSectionTitle}>未命中默认动作</div>
                        <div className={style.ruleHelp}>没有策略命中时的兜底动作，通常用于确定默认放通或默认拒绝。</div>
                    </div>
                    <FormItem className={style.defaultActionField} name="default_action">
                        <Select options={[
                            { label: '放通', value: TrafficSecurityAction.ALLOW },
                            { label: '拒绝', value: TrafficSecurityAction.DENY },
                        ]} />
                    </FormItem>
                </div>
                <div className={shared.ruleList}>
                    {policies.map((policy, index) => (
                        <div className={`${shared.policy} ${collapsedSecurityPolicyIndexes.has(index) ? shared.policyCollapsed : ''}`} key={`${policy.action}-${index}`}>
                            <div className={shared.policyHead} onClick={() => toggleSecurityPolicyCollapsed(index)}>
                                <div className={shared.policyHeadMain}>
                                    <span className={shared.caret}><ChevronRightIcon /></span>
                                    <div>
                                        <div className={shared.policyIndex}>规则 [{index + 1}]</div>
                                        {collapsedSecurityPolicyIndexes.has(index) && (
                                            <div className={shared.policySummary}>{securityPolicySummary(policy)}</div>
                                        )}
                                    </div>
                                </div>
                                <div className={shared.policyHeadActions} onClick={(e) => e.stopPropagation()}>
                                    <Select
                                        className={style.actionSelect}
                                        options={[
                                            { label: '放通', value: TrafficSecurityAction.ALLOW },
                                            { label: '拒绝', value: TrafficSecurityAction.DENY },
                                        ]}
                                        value={readAction(policy.action)}
                                        onChange={(value) => setSecurityPolicy(index, (item) => ({ ...item, action: value as string }))}
                                    />
                                    <Button
                                        shape="circle"
                                        variant="text"
                                        disabled={policies.length <= 1}
                                        onClick={() => setRule((prev) => ({ ...(prev as TrafficSecurityRule), policies: policies.filter((_, idx) => idx !== index) }))}
                                    >
                                        <CloseIcon />
                                    </Button>
                                </div>
                            </div>
                            <div className={shared.policyBody}>
                                <div className={shared.step} data-step="1">
                                    <div className={shared.stepTitle}>接口范围<span className={shared.stepHint}>满足该范围的请求才会进入鉴权条件判断</span></div>
                                    <div className={shared.stepContent}>
                                        {renderApiEditor(policy.api, (next) => setSecurityPolicy(index, (item) => ({ ...item, api: next })))}
                                    </div>
                                </div>
                                <div className={shared.step} data-step="2">
                                    <div className={shared.stepTitle}>匹配条件<span className={shared.stepHint}>满足以下条件的请求会命中该策略</span></div>
                                    <div className={shared.stepContent}>
                                        {renderMatchRuleEditor(policy.traffic_match_rule, (next) => setSecurityPolicy(index, (item) => ({ ...item, traffic_match_rule: next })))}
                                    </div>
                                </div>
                                <div className={shared.step} data-step="3">
                                    <div className={shared.stepTitle}>鉴权结果<span className={shared.stepHint}>命中后的处理动作；拒绝会返回下方响应信息</span></div>
                                    <div className={shared.stepContent}>
                                        {readAction(policy.action) === TrafficSecurityAction.DENY ? (
                                            <div className={`${shared.verdictEdit} ${shared.verdictDeny}`}>
                                                <div className={shared.verdictEditHead}>
                                                    <div className={shared.verdictIcon}><CloseIcon /></div>
                                                    <div>
                                                        <div className={shared.verdictHead}>拒绝</div>
                                                        <div className={shared.verdictDesc}>命中请求将被拦截，按下方配置返回响应</div>
                                                    </div>
                                                </div>
                                                <div className={shared.verdictEditDetail}>
                                                    <div className={shared.verdictEditDetailInner}>
                                                        <div>
                                                            <div className={shared.editLabel}>拒绝状态码</div>
                                                            <InputNumber
                                                                min={100}
                                                                max={599}
                                                                value={policy.reject_effect?.status_code ?? 403}
                                                                onChange={(value) => setSecurityPolicy(index, (item) => ({
                                                                    ...item,
                                                                    reject_effect: { ...(item.reject_effect || {}), status_code: Number(value ?? 403) },
                                                                }))}
                                                            />
                                                        </div>
                                                        <div>
                                                            <div className={shared.editLabel}>拒绝 Code</div>
                                                            <Input
                                                                value={policy.reject_effect?.code || ''}
                                                                onChange={(value) => setSecurityPolicy(index, (item) => ({
                                                                    ...item,
                                                                    reject_effect: { ...(item.reject_effect || {}), code: value },
                                                                }))}
                                                            />
                                                        </div>
                                                        <div className={shared.full}>
                                                            <div className={shared.editLabel}>拒绝文案</div>
                                                            <Input
                                                                value={policy.reject_effect?.message || ''}
                                                                onChange={(value) => setSecurityPolicy(index, (item) => ({
                                                                    ...item,
                                                                    reject_effect: { ...(item.reject_effect || {}), message: value },
                                                                }))}
                                                            />
                                                        </div>
                                                    </div>
                                                </div>
                                            </div>
                                        ) : (
                                            <div className={`${shared.verdictEdit} ${shared.verdictAllow}`}>
                                                <div className={shared.verdictEditHead}>
                                                    <div className={shared.verdictIcon}><CheckIcon /></div>
                                                    <div>
                                                        <div className={shared.verdictHead}>放通</div>
                                                        <div className={shared.verdictDesc}>命中请求允许继续访问，无需配置拒绝响应</div>
                                                    </div>
                                                </div>
                                            </div>
                                        )}
                                    </div>
                                </div>
                            </div>
                        </div>
                    ))}
                    <Button
                        className={shared.addRuleButton}
                        variant="dashed"
                        icon={<AddIcon />}
                        onClick={() => setRule((prev) => ({ ...(prev as TrafficSecurityRule), policies: [...policies, defaultTrafficSecurityRule().policies[0]] }))}
                    >
                        添加策略
                    </Button>
                </div>
            </>
        );
    };

    const renderEditMirrorRules = () => {
        const mirror = rule as TrafficMirror;
        const rules = mirror.rules || [];
        return (
            <div className={shared.ruleList}>
                {rules.map((item, index) => (
                    <div className={shared.group} key={`${item.source?.service}-${item.destination?.service}-${index}`}>
                        <div className={shared.groupHead}>
                            <span className={shared.groupTitle}><span className={shared.groupIndex}>{index + 1}</span>镜像规则</span>
                            <Space size={8}>
                                <Switch
                                    label={['启用', '禁用']}
                                    value={!item.disable}
                                    onChange={(value) => setMirrorRule(index, (current) => ({ ...current, disable: !value }))}
                                />
                                <Button
                                    shape="circle"
                                    variant="text"
                                    disabled={rules.length <= 1}
                                    onClick={() => setRule((prev) => ({ ...(prev as TrafficMirror), rules: rules.filter((_, idx) => idx !== index) }))}
                                >
                                    <CloseIcon />
                                </Button>
                            </Space>
                        </div>
                        <div className={shared.policyBody}>
                            <div className={shared.step} data-step="1">
                                <div className={shared.stepTitle}>镜像接口<span className={shared.stepHint}>满足该接口范围的请求才会被镜像</span></div>
                                <div className={shared.stepContent}>
                                    {renderApiEditor(item.source?.api, (next) => setMirrorRule(index, (current) => ({ ...current, source: { ...(current.source || {}), api: next } })))}
                                </div>
                            </div>
                            <div className={shared.step} data-step="2">
                                <div className={shared.stepTitle}>流量匹配<span className={shared.stepHint}>来源服务与请求标签共同界定需要镜像的流量</span></div>
                                <div className={shared.stepContent}>
                                    <div className={shared.kv2}>
                                        <div>
                                            <div className={shared.editLabel}>来源命名空间</div>
                                            <Select
                                                filterable
                                                creatable
                                                options={namespaceOptions}
                                                value={item.source?.namespace || ''}
                                                onChange={(value) => setMirrorRule(index, (current) => ({ ...current, source: { ...(current.source || {}), namespace: value as string } }))}
                                            />
                                        </div>
                                        <div>
                                            <div className={shared.editLabel}>来源服务</div>
                                            <Input
                                                value={item.source?.service || ''}
                                                onChange={(value) => setMirrorRule(index, (current) => ({ ...current, source: { ...(current.source || {}), service: value } }))}
                                            />
                                        </div>
                                        <div className={shared.full}>
                                            <div className={shared.editLabel}>请求标签匹配</div>
                                            {renderMatchRuleEditor(item.source?.traffic_match_rule, (next) => setMirrorRule(index, (current) => ({ ...current, source: { ...(current.source || {}), traffic_match_rule: next } })))}
                                        </div>
                                    </div>
                                </div>
                            </div>
                            <div className={shared.step} data-step="3">
                                <div className={shared.stepTitle}>采样<span className={shared.stepHint}>命中流量中按比例采样镜像，可设置生效时长</span></div>
                                <div className={shared.stepContent}>
                                    <div className={shared.kv2}>
                                        <div>
                                            <div className={shared.editLabel}>镜像比例</div>
                                            <InputNumber
                                                min={0}
                                                max={100}
                                                suffix="%"
                                                value={item.mirror_percent ?? 0}
                                                onChange={(value) => setMirrorRule(index, (current) => ({ ...current, mirror_percent: Number(value ?? 0) }))}
                                            />
                                        </div>
                                        <div>
                                            <div className={shared.editLabel}>生效时长</div>
                                            <Input
                                                value={durationText(item.duration)}
                                                onChange={(value) => setMirrorRule(index, (current) => ({ ...current, duration: value }))}
                                            />
                                        </div>
                                    </div>
                                </div>
                            </div>
                            <div className={shared.step} data-step="4">
                                <div className={shared.stepTitle}>镜像目标<span className={shared.stepHint}>采样流量被复制到的目标服务与实例标签</span></div>
                                <div className={shared.stepContent}>
                                    <div className={shared.kv2}>
                                        <div>
                                            <div className={shared.editLabel}>目标命名空间</div>
                                            <Select
                                                filterable
                                                creatable
                                                options={namespaceOptions}
                                                value={item.destination?.namespace || ''}
                                                onChange={(value) => setMirrorRule(index, (current) => ({ ...current, destination: { ...(current.destination || {}), namespace: value as string } }))}
                                            />
                                        </div>
                                        <div>
                                            <div className={shared.editLabel}>目标服务</div>
                                            <Input
                                                value={item.destination?.service || ''}
                                                onChange={(value) => setMirrorRule(index, (current) => ({ ...current, destination: { ...(current.destination || {}), service: value } }))}
                                            />
                                        </div>
                                        <div className={shared.full}>
                                            <div className={shared.editLabel}>目标实例标签</div>
                                            {renderMatchRecordEditor(item.destination?.labels, (next) => setMirrorRule(index, (current) => ({ ...current, destination: { ...(current.destination || {}), labels: next } })))}
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                ))}
                <Button
                    className={shared.addRuleButton}
                    variant="dashed"
                    icon={<AddIcon />}
                    onClick={() => setRule((prev) => ({ ...(prev as TrafficMirror), rules: [...rules, defaultTrafficMirrorRule().rules[0]] }))}
                >
                    添加镜像规则
                </Button>
            </div>
        );
    };

    const renderEditMockRules = () => {
        const mock = rule as TrafficMock;
        const rules = mock.rules || [];
        return (
            <div className={shared.ruleList}>
                {rules.map((item, index) => (
                    <div className={shared.group} key={`${item.source?.service}-${item.source?.api?.method}-${index}`}>
                        <div className={shared.groupHead}>
                            <span className={shared.groupTitle}><span className={shared.groupIndex}>{index + 1}</span>Mock 规则</span>
                            <Space size={8}>
                                <Switch
                                    label={['启用', '禁用']}
                                    value={!item.disable}
                                    onChange={(value) => setMockRule(index, (current) => ({ ...current, disable: !value }))}
                                />
                                <Button
                                    shape="circle"
                                    variant="text"
                                    disabled={rules.length <= 1}
                                    onClick={() => setRule((prev) => ({ ...(prev as TrafficMock), rules: rules.filter((_, idx) => idx !== index) }))}
                                >
                                    <CloseIcon />
                                </Button>
                            </Space>
                        </div>
                        <div className={shared.kv2}>
                            <div>
                                <div className={shared.editLabel}>来源命名空间</div>
                                <Select
                                    filterable
                                    creatable
                                    options={namespaceOptions}
                                    value={item.source?.namespace || ''}
                                    onChange={(value) => setMockRule(index, (current) => ({ ...current, source: { ...(current.source || {}), namespace: value as string } }))}
                                />
                            </div>
                            <div>
                                <div className={shared.editLabel}>来源服务</div>
                                <Input
                                    value={item.source?.service || ''}
                                    onChange={(value) => setMockRule(index, (current) => ({ ...current, source: { ...(current.source || {}), service: value } }))}
                                />
                            </div>
                            <div className={shared.full}>
                                <div className={shared.editLabel}>接口范围</div>
                                {renderApiEditor(item.source?.api, (next) => setMockRule(index, (current) => ({ ...current, source: { ...(current.source || {}), api: next } })))}
                            </div>
                            <div>
                                <div className={shared.editLabel}>Mock 比例</div>
                                <InputNumber
                                    min={0}
                                    max={100}
                                    suffix="%"
                                    value={item.mock_percent ?? 0}
                                    onChange={(value) => setMockRule(index, (current) => ({ ...current, mock_percent: Number(value ?? 0) }))}
                                />
                            </div>
                            <div>
                                <div className={shared.editLabel}>延迟</div>
                                <Input
                                    value={durationText(item.delay)}
                                    onChange={(value) => setMockRule(index, (current) => ({ ...current, delay: value }))}
                                />
                            </div>
                            <div>
                                <div className={shared.editLabel}>响应状态码</div>
                                <InputNumber
                                    min={100}
                                    max={599}
                                    value={item.response?.status_code ?? 200}
                                    onChange={(value) => setMockRule(index, (current) => ({ ...current, response: { ...(current.response || {}), status_code: Number(value ?? 200) } }))}
                                />
                            </div>
                            <div>
                                <div className={shared.editLabel}>响应 Code</div>
                                <Input
                                    value={item.response?.code || ''}
                                    onChange={(value) => setMockRule(index, (current) => ({ ...current, response: { ...(current.response || {}), code: value } }))}
                                />
                            </div>
                            <div className={shared.full}>
                                <div className={shared.editLabel}>响应文案</div>
                                <Input
                                    value={item.response?.message || ''}
                                    onChange={(value) => setMockRule(index, (current) => ({ ...current, response: { ...(current.response || {}), message: value } }))}
                                />
                            </div>
                            <div className={shared.full}>
                                <div className={shared.editLabel}>响应头</div>
                                {renderStringRecordEditor(item.response?.headers, (next) => setMockRule(index, (current) => ({ ...current, response: { ...(current.response || {}), headers: next } })), '添加响应头')}
                            </div>
                            <div className={shared.full}>
                                <div className={shared.editLabel}>匹配条件</div>
                                {renderMatchRuleEditor(item.source?.traffic_match_rule, (next) => setMockRule(index, (current) => ({ ...current, source: { ...(current.source || {}), traffic_match_rule: next } })))}
                            </div>
                            <div className={shared.full}>
                                <div className={shared.editLabel}>响应体</div>
                                <Textarea
                                    autosize={{ minRows: 3, maxRows: 8 }}
                                    value={item.response?.body || ''}
                                    onChange={(value) => setMockRule(index, (current) => ({ ...current, response: { ...(current.response || {}), body: value } }))}
                                />
                            </div>
                        </div>
                    </div>
                ))}
                <Button
                    className={shared.addRuleButton}
                    variant="dashed"
                    icon={<AddIcon />}
                    onClick={() => setRule((prev) => ({ ...(prev as TrafficMock), rules: [...rules, defaultTrafficMockRule().rules[0]] }))}
                >
                    添加 Mock 规则
                </Button>
            </div>
        );
    };

    const renderReadonlyPayload = () => {
        if (kind === 'security') return renderSecurityRules();
        if (kind === 'mirror') return renderMirrorRules();
        return renderMockRules();
    };

    const renderEditablePayload = () => {
        if (kind === 'security') return renderEditSecurityRules();
        if (kind === 'mirror') return renderEditMirrorRules();
        return renderEditMockRules();
    };

    const renderPayload = () => (
        <div className={shared.section}>
            <div className={shared.sectionHeader}>
                <span>{kind === 'security' ? '鉴权策略' : kind === 'mirror' ? '镜像规则' : 'Mock 规则'}</span>
                <span className={shared.countTag}>{trafficRuleCount(kind, rule)} 条</span>
            </div>
            <div className={shared.sectionBody}>
                {editable ? renderEditablePayload() : renderReadonlyPayload()}
            </div>
        </div>
    );

    const renderStickyTool = (
        <StickyTool
            style={{ zIndex: 1000 }}
            placement="right-bottom"
            offset={[-10, 200]}
        >
            <StickyItem
                label=""
                icon={editable ? (
                    <RuleStickyAction label="保存" icon={<SaveIcon />} onClick={() => form.submit()} />
                ) : (
                    <span style={rule.editable === false ? { opacity: 0.45, cursor: 'not-allowed' } : undefined}>
                        <RuleStickyAction
                            label="编辑"
                            icon={<Edit1Icon />}
                            onClick={() => {
                                if (rule.editable === false) return;
                                setEditable(true);
                            }}
                        />
                    </span>
                )}
            />
            {editable && (
                <StickyItem
                    label=""
                    icon={(
                        <RuleStickyAction
                            label="撤销"
                            icon={<RollbackIcon />}
                            onClick={() => {
                                if (op === 'create') {
                                    refresh(true);
                                    return;
                                }
                                setEditable(false);
                            }}
                        />
                    )}
                />
            )}
            {!editable && op !== 'create' && (
                <StickyItem
                    label=""
                    icon={<RuleStickyAction label="发布" icon={<RocketIcon />} onClick={() => setPublishVisible(true)} />}
                />
            )}
        </StickyTool>
    );

    return (
        <div className={style.editor}>
            <Form form={form} layout="vertical" onSubmit={onSubmit}>
                <div className={shared.section}>
                    <div className={shared.sectionHeader}>基础信息</div>
                    <div className={shared.sectionBody}>{renderCommonFields()}</div>
                </div>
                {renderPayload()}
                {renderStickyTool}
            </Form>
            <PublishForm
                ruleId={rule.id || ''}
                ruleName={rule.name || ''}
                resource={TrafficGovernanceAuthResource[kind] as PolicySourceType}
                releaseResource={TrafficGovernanceReleaseResource[kind]}
                visible={publishVisible}
                close={() => setPublishVisible(false)}
            />
        </div>
    );
};

export default React.memo(TrafficGovernanceEditor);

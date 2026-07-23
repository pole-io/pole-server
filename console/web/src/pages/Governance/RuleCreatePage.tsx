import React from 'react';
import { Breadcrumb, Empty, Loading } from 'components/Fluent';
import { useNavigate, useSearchParams } from 'react-router-dom';

import { useAppDispatch } from 'modules/store';
import { resetCustomRoute } from 'modules/governance/route';
import { resetRateLimitRule } from 'modules/governance/ratelimit';
import { resetCircuitBreaker } from 'modules/governance/circuitbreaker';
import { resetFaultDetect } from 'modules/governance/faultdetect';
import { resetLosslessRule } from 'modules/governance/lossless';
import { resetLaneGroup } from 'modules/governance/lane_group';
import CustomRouteEditor from './Router/CustomRouteEditor';
import RateLimitEditor from './RateLimit/RateLimitEditor';
import CircuitBreakerEditor from './CircuitBreaker/CircuitBreakerEditor';
import FaultDetectEditor from './CircuitBreaker/FaultDetectEditor';
import LossLessEditor from './LossLess/LossLessEditor';
import LaneGroupEdtor from './Router/LaneGroupEdtor';
import TrafficGovernanceEditor from './Security/TrafficGovernanceEditor';
import RuleDetailFrame from './RuleRelease/RuleDetailFrame';
import { LimitType } from 'services/ratelimit';
import { TrafficGovernanceKind, TrafficGovernanceKindLabel } from 'services/traffic_governance';
import { GovernanceServiceContext, GovernanceServiceRole } from './shared/serviceContext';
import { RuleNamespaceProvider } from './shared/ruleNamespace';
import style from './RuleRelease/RuleDetailDrawer.module.less';

const { BreadcrumbItem } = Breadcrumb;

type CreateRuleKind = 'route' | 'ratelimit-local' | 'ratelimit-global' | 'circuitbreaker' | 'faultdetect' | 'lossless' | 'lane' | 'traffic-security' | 'traffic-mirror' | 'traffic-mock';

interface CreateTarget {
    label: string;
    trafficKind?: TrafficGovernanceKind;
    limitType?: LimitType;
}

const createTargets: Record<CreateRuleKind, CreateTarget> = {
    route: { label: '路由' },
    'ratelimit-local': { label: '限流', limitType: LimitType.LOCAL },
    'ratelimit-global': { label: '限流', limitType: LimitType.GLOBAL },
    circuitbreaker: { label: '熔断' },
    faultdetect: { label: '探测' },
    lossless: { label: '无损' },
    lane: { label: '泳道' },
    'traffic-security': { label: TrafficGovernanceKindLabel.security, trafficKind: 'security' },
    'traffic-mirror': { label: TrafficGovernanceKindLabel.mirror, trafficKind: 'mirror' },
    'traffic-mock': { label: TrafficGovernanceKindLabel.mock, trafficKind: 'mock' },
};

const isCreateRuleKind = (value?: string | null): value is CreateRuleKind => !!value && value in createTargets;

const RuleCreatePage: React.FC = () => {
    const dispatch = useAppDispatch();
    const navigate = useNavigate();
    const [searchParams] = useSearchParams();
    const kindParam = searchParams.get('kind');
    const kind = isCreateRuleKind(kindParam) ? kindParam : undefined;
    const ruleNamespace = searchParams.get('ruleNamespace') || '';
    const [ready, setReady] = React.useState(false);

    const serviceContext = React.useMemo<GovernanceServiceContext | undefined>(() => {
        const namespace = searchParams.get('namespace') || '';
        const service = searchParams.get('service') || '';
        const role = searchParams.get('role');
        if (!namespace || !service || (role !== 'caller' && role !== 'callee')) return undefined;
        return { namespace, service, role: role as GovernanceServiceRole };
    }, [searchParams]);

    React.useEffect(() => {
        setReady(false);
        if (!kind) return;
        if (kind === 'route') dispatch(resetCustomRoute());
        if (kind === 'ratelimit-local' || kind === 'ratelimit-global') dispatch(resetRateLimitRule());
        if (kind === 'circuitbreaker') dispatch(resetCircuitBreaker());
        if (kind === 'faultdetect') dispatch(resetFaultDetect());
        if (kind === 'lossless') dispatch(resetLosslessRule());
        if (kind === 'lane') dispatch(resetLaneGroup());
        setReady(true);
    }, [dispatch, kind]);

    const leaveCreatePage = React.useCallback((_close: boolean) => {
        navigate('/governance/workbench');
    }, [navigate]);

    const renderEditor = () => {
        if (!kind) return null;
        const target = createTargets[kind];
        if (kind === 'route') return <CustomRouteEditor op="create" editable serviceContext={serviceContext} refresh={leaveCreatePage} />;
        if (kind === 'ratelimit-local' || kind === 'ratelimit-global') return <RateLimitEditor limitType={target.limitType || LimitType.LOCAL} visible op="create" serviceContext={serviceContext} refresh={leaveCreatePage} />;
        if (kind === 'circuitbreaker') return <CircuitBreakerEditor op="create" serviceContext={serviceContext} refresh={leaveCreatePage} />;
        if (kind === 'faultdetect') return <FaultDetectEditor op="create" serviceContext={serviceContext} refresh={leaveCreatePage} />;
        if (kind === 'lossless') return <LossLessEditor visible op="create" serviceContext={serviceContext} refresh={leaveCreatePage} />;
        if (kind === 'lane') return <LaneGroupEdtor op="create" serviceContext={serviceContext} refresh={leaveCreatePage} />;
        return <TrafficGovernanceEditor kind={target.trafficKind!} op="create" visible serviceContext={serviceContext} refresh={leaveCreatePage} />;
    };

    if (!kind || !ruleNamespace) {
        return (
            <div className={style.standalonePage}>
                <Breadcrumb className={style.detailBreadcrumb} maxItemWidth="220px">
                    <BreadcrumbItem onClick={() => navigate('/governance/workbench')}>治理工作台</BreadcrumbItem>
                    <BreadcrumbItem>新建规则</BreadcrumbItem>
                </Breadcrumb>
                <div className={style.emptyPane}><Empty description={!kind ? '缺少或不支持的规则类型' : '缺少规则归属环境'} /></div>
            </div>
        );
    }

    const target = createTargets[kind];
    return (
        <div className={`${style.standalonePage} ${style.createPage}`}>
            <Breadcrumb className={style.detailBreadcrumb} maxItemWidth="220px">
                <BreadcrumbItem onClick={() => navigate('/governance/workbench')}>治理工作台</BreadcrumbItem>
                <BreadcrumbItem>新建{target.label}规则</BreadcrumbItem>
            </Breadcrumb>
            <RuleDetailFrame title={`新建${target.label}规则`} subtitle={`${target.label} · 归属环境 ${ruleNamespace}`}>
                {ready ? (
                    <RuleNamespaceProvider value={ruleNamespace}>
                        <div className={style.createPane}>{renderEditor()}</div>
                    </RuleNamespaceProvider>
                ) : <div className={style.loadingPane}><Loading text="正在初始化规则表单..." /></div>}
            </RuleDetailFrame>
        </div>
    );
};

export default React.memo(RuleCreatePage);

import React from 'react';
import classNames from 'classnames';
import { Button, Tag } from 'components/Fluent';

import style from './index.module.less';

export interface EnvironmentResourceItem {
    namespace: string;
    summary?: string;
}

interface EnvironmentResourceSwitcherProps {
    currentNamespace: string;
    items: EnvironmentResourceItem[];
    resourceLabel: string;
    onSelect: (namespace: string) => void;
    density?: 'default' | 'compact';
    presentation?: 'cards' | 'tabs';
}

const EnvironmentResourceSwitcher: React.FC<EnvironmentResourceSwitcherProps> = ({
    currentNamespace,
    items,
    resourceLabel,
    onSelect,
    density = 'default',
    presentation = 'cards',
}) => {
    const environments = React.useMemo(() => [...items].sort((left, right) => {
        if (left.namespace === currentNamespace) return -1;
        if (right.namespace === currentNamespace) return 1;
        return left.namespace.localeCompare(right.namespace);
    }), [currentNamespace, items]);

    if (environments.length === 0) {
        return null;
    }

    if (presentation === 'tabs') {
        return (
            <section className={style.tabContainer} aria-label={`${resourceLabel}跨环境视图`}>
                <span className={style.tabLabel}>环境</span>
                <div className={style.environmentTabs} role="tablist" aria-label={`${resourceLabel}环境版本`}>
                    {environments.map((item) => {
                        const active = item.namespace === currentNamespace;
                        return (
                            <Button
                                key={item.namespace}
                                role="tab"
                                aria-selected={active}
                                theme="default"
                                variant="text"
                                className={style.environmentTab}
                                onClick={() => {
                                    if (!active) onSelect(item.namespace);
                                }}
                            >
                                <span className={style.namespace}>{item.namespace}</span>
                                {item.summary && <span className={style.tabSummary}>{item.summary}</span>}
                            </Button>
                        );
                    })}
                </div>
            </section>
        );
    }

    return (
        <section
            className={classNames(style.container, { [style.compact]: density === 'compact' })}
            aria-label={`${resourceLabel}跨环境视图`}
        >
            <div className={style.heading}>
                <div>
                    <strong>环境版本</strong>
                    <span>同一逻辑{resourceLabel}在不同 namespace 环境中的独立实例</span>
                </div>
                <Tag variant="light">可访问 {environments.length} 个环境</Tag>
            </div>
            <div className={style.items}>
                {environments.map((item) => {
                    const active = item.namespace === currentNamespace;
                    return (
                        <Button
                            key={item.namespace}
                            theme="default"
                            variant="outline"
                            aria-current={active ? 'page' : undefined}
                            className={style.item}
                            onClick={() => {
                                if (!active) onSelect(item.namespace);
                            }}
                        >
                            <span className={style.namespace}>{item.namespace}</span>
                            <span className={style.summary}>{item.summary || '可查看'}</span>
                            {active && <span className={style.current}>当前环境</span>}
                        </Button>
                    );
                })}
            </div>
        </section>
    );
};

export default React.memo(EnvironmentResourceSwitcher);

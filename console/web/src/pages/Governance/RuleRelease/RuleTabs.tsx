import React from 'react';
import { Tabs } from "tdesign-react";
import type { PageInfo, PaginationProps, TableRowData } from 'tdesign-react';

import { Op } from 'services/types';
import ReleaseTable from './ReleaseTable';
import SubscribeTable from 'components/SubscribeTable';

const { TabPanel } = Tabs;

export interface VersionProps {
    datas: TableRowData[];
    loading: boolean;
    pagination: PaginationProps;
    onPageChange?: (pageInfo: PageInfo) => void;
    action: (op: Op, row: TableRowData) => void;
    editable: boolean;
    deleteable: boolean;
}

interface IRuleTabsProps {
    view: React.ReactNode;
    subscribe: React.ReactNode;
    versions: VersionProps;
    op: Op
    onRulesView?: () => void;
    onVersionView?: () => void;
    onSubscribeView?: () => void;
}

const RuleTabs: React.FC<IRuleTabsProps> = (props) => {

    const [activeTab, setActiveTab] = React.useState('rule_view');

    return (
        <Tabs value={activeTab} onChange={(v) => {
            switch (v) {
                case 'rule_view':
                    props.onRulesView?.();
                    break;
                case 'listener':
                    props.onSubscribeView?.();
                    break;
                case 'versions':
                    props.onVersionView?.();
                    break;
            }
            setActiveTab(v as string)
         }}>
            <TabPanel label={"规则"} value={"rule_view"}>
                <>{props.view}</>
            </TabPanel>
            <TabPanel label='版本' value={'versions'} disabled={props.op === 'create'}>
                {activeTab === 'versions' && (
                    <ReleaseTable
                        datas={props.versions.datas}
                        action={props.versions.action}
                        editable={props.versions.editable}
                        deleteable={props.versions.deleteable}
                        loading={props.versions.loading}
                        pagination={props.versions.pagination}
                        onPageChange={props.versions.onPageChange}
                    />
                )}
            </TabPanel>
            <TabPanel label='监听' value={'listener'} disabled={props.op === 'create'}>
                {activeTab === 'listener' && (
                    <>{props.subscribe}</>
                )}
            </TabPanel>
        </Tabs>
    )
}

export default React.memo(RuleTabs)

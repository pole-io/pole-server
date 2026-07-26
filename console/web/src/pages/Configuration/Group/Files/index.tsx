import React, { useState, useEffect } from 'react';
import { Button, Tooltip, Breadcrumb, Tree, Input, TreeInstanceFunctions, Popconfirm, Tag } from 'components/Fluent';
import { Delete1Icon, FileAddIcon, Icon, RefreshIcon } from 'components/Fluent/icons';
import { useNavigate, BrowserRouterProps } from 'components/Router';
import type { TreeProps, TreeNodeModel } from 'components/Fluent';

import AuthorizeInput from 'components/Authorize';
import { useAppDispatch, useAppSelector } from 'modules/store';
import { openErrNotification, openInfoNotification } from 'utils/notifition';
import style from './index.module.less';
import { ConfigFileView, FileStatusMap } from 'services/config_files';
import FileCreator from './FileCreator';
import { editorConfigGroup, listConfigGroups, selectConfigGroup } from 'modules/configuration/group';
import FileView from './FileView';
import { Op } from 'services/types';
import { editorConfigFile, listAllConfigFiles, removeConfigFeils, selectConfigFile } from 'modules/configuration/file';
import { PolicySourceType } from 'services/auth_policy';
import { describeConfigGroupEnvironments, type ConfigFileGroupView } from 'services/config_group';
import EnvironmentResourceSwitcher from 'components/EnvironmentResourceSwitcher';

const { BreadcrumbItem } = Breadcrumb;
interface IFileListProps {
}

const renderIcon: TreeProps['icon'] = (node) => {
    let name = 'file';
    if (node.getChildren(true)) {
        if (node.expanded) {
            name = 'folder-open';
            if (node.loading) {
                name = 'loading';
            }
        } else {
            name = 'folder';
        }
    }
    return <Icon name={name} />;
};

const statusTheme = (status?: string) => {
    const theme = FileStatusMap?.[status as keyof typeof FileStatusMap]?.theme;
    return (theme || 'default') as 'default' | 'primary' | 'success' | 'warning' | 'danger';
};

const statusText = (status?: string) => FileStatusMap?.[status as keyof typeof FileStatusMap]?.text || '未发布';

const renderNodeLabel = (file?: ConfigFileView, label?: string) => {
    if (!file) {
        return <span className={style.treeNodeName}>{label}</span>;
    }
    return (
        <span className={style.treeNodeLabel}>
            <span className={style.treeNodeName}>{label}</span>
            <Tag size="small" variant="light" theme={statusTheme(file.status)} className={style.treeNodeStatus}>
                {statusText(file.status)}
            </Tag>
        </span>
    );
};

const renderTree = (files: ConfigFileView[]) => {
    const root: any = [];

    files.forEach(file => {
        const parts = file.name.split('/').filter(Boolean);
        let currentLevel = root;

        parts.forEach((part, idx) => {
            let node = currentLevel.find((item: any) => item.label === part);
            if (!node) {
                node = {
                    label: renderNodeLabel(idx === parts.length - 1 ? file : undefined, part),
                    rawLabel: part,
                    value: idx === parts.length - 1 ? file.name : part,
                    disabled: idx === parts.length - 1 ? false : true,
                    children: idx === parts.length - 1 ? false : [],
                    file: idx === parts.length - 1 ? file : undefined,
                };
                currentLevel.push(node);
            }
            if (node.children !== false) {
                currentLevel = node.children;
            }
        });
    });
    return root;
}

export default React.memo((props: IFileListProps & BrowserRouterProps) => {
    const dispatch = useAppDispatch();
    const navigate = useNavigate()
    const urlParams = new URLSearchParams(window.location.search);
    const namespace = urlParams.get('namespace');
    const group = urlParams.get('group');
    const requestedFile = urlParams.get('file');

    const treeRef = React.useRef<TreeInstanceFunctions>(null);

    const groupState = useAppSelector(selectConfigGroup);
    const { editGroup: ownerGroup } = groupState;
    const activeGroup = ownerGroup?.namespace === namespace && ownerGroup?.name === group ? ownerGroup : undefined;

    const fileState = useAppSelector(selectConfigFile);
    const { datas } = fileState;

    // 合并编辑相关状态
    const [editState, setEditState] = useState<{
        activeNode?: TreeNodeModel;
        visible: boolean;
        mode: Op;
        editable?: boolean;
        deleteable?: boolean;
        nodeFilter: string;
    }>({
        activeNode: undefined,
        visible: false,
        mode: 'view',
        editable: activeGroup?.editable,
        deleteable: activeGroup?.deleteable,
        nodeFilter: '',
    });
    const [authorizeVisible, setAuthorizeVisible] = useState(false);
    const [environmentGroups, setEnvironmentGroups] = useState<ConfigFileGroupView[]>([]);
    const selectedFileName = editState.activeNode?.value as string | undefined;

    // 模拟远程请求
    const fetchData = () => {
        dispatch(listAllConfigFiles({
            param: {
                namespace: namespace ? namespace : '',
                group: group ? group : '',
            }
        })).then(res => {
            if (res.meta.requestStatus === 'rejected') {
                openErrNotification('获取配置文件列表失败', res?.payload as string);
            }
        })
    }

    React.useEffect(() => {
        fetchData();
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [namespace, group]);

    React.useEffect(() => {
        let active = true;
        if (!group) return () => { active = false; };
        describeConfigGroupEnvironments(group)
            .then((items) => {
                if (active) setEnvironmentGroups(items);
            })
            .catch(() => {
                if (active) setEnvironmentGroups([]);
            });
        return () => { active = false; };
    }, [group]);

    React.useEffect(() => {
        if (!requestedFile || !namespace || !group || editState.visible) return;
        const requested = datas.find((item) => item.name === requestedFile);
        if (!requested) return;
        dispatch(editorConfigFile({
            ...requested,
            namespace,
            group,
            editable: activeGroup?.editable,
            deleteable: activeGroup?.deleteable,
        }));
        setEditState((prev) => ({
            ...prev,
            activeNode: { value: requested.name, data: { value: requested.name } } as TreeNodeModel,
            visible: true,
            mode: 'edit',
        }));
    }, [activeGroup?.deleteable, activeGroup?.editable, datas, dispatch, editState.visible, group, namespace, requestedFile]);

    React.useEffect(() => {
        if (!namespace || !group) {
            return;
        }
        if (activeGroup) {
            setEditState((prev) => ({
                ...prev,
                editable: activeGroup.editable,
                deleteable: activeGroup.deleteable,
            }));
            return;
        }
        dispatch(listConfigGroups({
            param: {
                namespace,
                group,
                offset: 0,
                limit: 1,
            },
        })).then((res) => {
            if (listConfigGroups.fulfilled.match(res)) {
                const nextGroup = res.payload.datas[0];
                if (nextGroup) {
                    dispatch(editorConfigGroup(nextGroup));
                    setEditState((prev) => ({
                        ...prev,
                        editable: nextGroup.editable,
                        deleteable: nextGroup.deleteable,
                    }));
                }
            } else {
                openErrNotification('获取配置分组失败', res.payload as string);
            }
        });
    }, [activeGroup, dispatch, group, namespace]);

    const treeSearch: TreeProps['filter'] = (node) => {
        const rawLabel = (node.data.rawLabel || node.data.value || '') as string;
        const rs = rawLabel.indexOf(editState.nodeFilter) >= 0;
        return rs;
    };

    const renderOperations = (node: TreeNodeModel) => (
        <>
            {/* 只有是激活状态的 node 才可以展示删除操作 */}
            {editState.activeNode && editState.activeNode.data.value === node.value && (
                <Tooltip content={editState.deleteable ? '删除' : '无权限操作'}>
                    <Popconfirm
                        content="确认删除配置文件吗？删除后当前文件的配置内容和发布记录将不可继续使用。"
                        destroyOnClose
                        placement="top"
                        showArrow
                        theme="default"
                        onConfirm={() => {
                            handleOperateFile(node, 'delete');
                        }}
                    >
                        <Button style={{ marginLeft: '10px' }} disabled={!editState.deleteable} size="small" variant='text' icon={<Delete1Icon />} />
                    </Popconfirm>
                </Tooltip>

            )}
        </>
    );

    const handleOperateFile = (node: TreeNodeModel, op: Op) => {
        switch (op) {
            case 'view':
                dispatch(editorConfigFile({
                    namespace: namespace ? namespace : '',
                    group: group ? group : '',
                    name: node.value as string,
                    editable: activeGroup?.editable,
                    deleteable: activeGroup?.deleteable,
                }));
                setEditState(prev => ({ ...prev, visible: true, mode: 'view', activeNode: { ...node } }));
                break;
            case 'create':
                dispatch(editorConfigFile({
                    namespace: namespace ? namespace : '',
                    group: group ? group : '',
                    name: node.value as string,
                    editable: activeGroup?.editable,
                    deleteable: activeGroup?.deleteable,
                }));
                setEditState(prev => ({ ...prev, visible: true, mode: 'create', activeNode: undefined }));
                break;
            case 'delete':
                dispatch(removeConfigFeils({
                    param: [{
                        namespace: namespace ? namespace : '',
                        group: group ? group : '',
                        name: node.value as string,
                    }]
                })).then((res) => {
                    if (res.meta.requestStatus === 'fulfilled') {
                        openInfoNotification('删除成功', '配置文件已删除');
                        setEditState(prev => ({ ...prev, visible: false, mode: 'view', activeNode: undefined }));
                        fetchData();
                    } else {
                        openErrNotification('删除配置文件失败', res.payload as string);
                    }
                });
                break;
        }
    }

    const mainView = (
        <div className={style.groupDetail}>
            <Breadcrumb maxItemWidth="200px" className={style.breadcrumb}>
                <BreadcrumbItem>配置中心</BreadcrumbItem>
                <BreadcrumbItem onClick={() => {
                    navigate(-1);
                }}>{namespace}</BreadcrumbItem>
                <BreadcrumbItem>
                    {group}
                </BreadcrumbItem>
            </Breadcrumb>
            {!selectedFileName && (
                <EnvironmentResourceSwitcher
                    currentNamespace={namespace || ''}
                    resourceLabel="配置分组"
                    presentation="tabs"
                    items={environmentGroups.map((item) => ({
                        namespace: item.namespace,
                        summary: `${item.fileCount || 0} 个配置文件`,
                    }))}
                    onSelect={(nextNamespace) => {
                        const params = new URLSearchParams(window.location.search);
                        params.set('namespace', nextNamespace);
                        params.set('group', group || '');
                        navigate(`${window.location.pathname}?${params.toString()}`);
                    }}
                />
            )}
            <div className={style.workbench}>
                <aside className={style.fileExplorer}>
                    <div className={style.fileExplorerHeader}>
                        <div className={style.fileExplorerTitle}>
                            <span>配置文件</span>
                            <div className={style.fileExplorerActions}>
                            <Tooltip content={activeGroup?.editable ? '新建配置文件' : '没有权限'}>
                                <Button
                                    shape="square"
                                    variant='text'
                                    size='small'
                                    icon={<FileAddIcon />}
                                    onClick={() => handleOperateFile({} as TreeNodeModel, 'create')}
                                    disabled={!activeGroup?.editable}
                                />
                            </Tooltip>
                            <Tooltip content="刷新">
                                <Button
                                    shape="square"
                                    variant='text'
                                    size='small'
                                    icon={<RefreshIcon />}
                                    onClick={() => fetchData()}
                                />
                            </Tooltip>
                            </div>
                        </div>
                        <Input
                            className={style.treeSearch}
                            placeholder="搜索文件路径"
                            value={editState.nodeFilter}
                            onChange={value => setEditState(s => ({ ...s, nodeFilter: value }))}
                        />
                    </div>
                    <div className={style.treeScroll}>
                    <Tree
                        ref={treeRef}
                        className={style.fileTree}
                        activable={true}
                        checkStrictly={true}
                        valueMode='onlyLeaf'
                        allowFoldNodeOnFilter={true}
                        data={renderTree(datas)}
                        hover
                        expandAll={true}
                        icon={renderIcon}
                        scroll={{ type: 'virtual' }}
                        filter={treeSearch}
                        onActive={(node, ctx) => {
                            setEditState(s => ({ ...s, activeNode: { ...ctx.node }, visible: true, mode: 'edit' }));
                            dispatch(editorConfigFile({
                                namespace: namespace ? namespace : '',
                                group: group ? group : '',
                                name: ctx.node.value as string,
                            }));
                            const params = new URLSearchParams(window.location.search);
                            params.set('file', ctx.node.value as string);
                            navigate(`${window.location.pathname}?${params.toString()}`, { replace: true });
                        }}
                        operations={renderOperations}
                    />
                    </div>
                </aside>
                <main className={style.canvas}>
                    {(editState.visible && editState.mode === 'edit') && (
                        <FileView
                            editable={activeGroup?.editable}
                            deleteable={activeGroup?.deleteable}
                            key={editState.activeNode?.value}
                            onAuthorize={() => setAuthorizeVisible(true)}
                        />
                    )}
                    {(!editState.visible || editState.mode !== 'edit') && (
                        <div className={style.emptyCanvas}>选择左侧配置文件后查看配置正文、发布记录和订阅查询</div>
                    )}
                </main>
            </div>
            {(editState.visible && editState.mode === 'create') && (
                <FileCreator
                    namespace={namespace ? namespace : ''}
                    group={group ? group : ''}
                    op={editState.mode}
                    visible={editState.visible && editState.mode === 'create'}
                    closeDrawer={() => {
                        setEditState(s => ({ ...s, visible: false, mode: 'view' }));
                        fetchData();
                    }}
                />
            )}
            {authorizeVisible && selectedFileName && activeGroup?.id && (
                <AuthorizeInput
                    resource_type={PolicySourceType.ConfigGroups}
                    resource_id={activeGroup?.id}
                    resource_name={`${namespace || '-'}/${group || '-'}/${selectedFileName}`}
                    visible={authorizeVisible}
                    onClose={() => setAuthorizeVisible(false)}
                />
            )}
        </div>
    );

    return (
        <>{mainView}</>
    );
});

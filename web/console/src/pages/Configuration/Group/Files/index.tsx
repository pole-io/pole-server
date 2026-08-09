import React, { useState } from 'react';
import { Button, Tooltip, Tree, Input, TreeInstanceFunctions, Popconfirm, Tag } from 'components/Fluent';
import { AddIcon, Delete1Icon, Icon, RefreshIcon } from 'components/Fluent/icons';
import { useNavigate } from 'components/Router';
import type { TreeProps, TreeNodeModel } from 'components/Fluent';

import AuthorizeInput from 'components/Authorize';
import { ResourceToolbar } from 'components/ResourceLayout';
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
import { describeConfigTemplates, type ConfigFileTemplate } from 'services/config_templates';
import EnvironmentResourceSwitcher from 'components/EnvironmentResourceSwitcher';
import GroupWorkspaceNav from '../GroupWorkspaceNav';
import TemplateWorkspace from '../../Template';

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
            <span className={style.treeNodeMeta}>
                <Tag size="small" variant="light" theme={file.configType === 'CONFIG_TEMPLATE' ? 'primary' : 'default'}>
                    {file.configType === 'CONFIG_TEMPLATE' ? '模板' : '文本'}
                </Tag>
                <Tag size="small" variant="light" theme={statusTheme(file.status)} className={style.treeNodeStatus}>
                    {statusText(file.status)}
                </Tag>
            </span>
        </span>
    );
};

const renderTemplateNodeLabel = (template: ConfigFileTemplate) => (
    <span className={style.treeNodeLabel}>
        <span className={style.treeNodeName}>{template.name}</span>
        <span className={style.treeNodeMeta}>
            <Tag size="small" variant="light" theme="primary">全局模板</Tag>
        </span>
    </span>
);

const renderTree = (files: ConfigFileView[], templates: ConfigFileTemplate[]) => {
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

    templates.forEach((template) => {
        root.push({
            label: renderTemplateNodeLabel(template),
            rawLabel: template.name,
            value: `template:${template.id}`,
            disabled: false,
            children: false,
            resourceKind: 'template',
            templateId: String(template.id),
        });
    });
    return root;
}

export default React.memo(() => {
    const dispatch = useAppDispatch();
    const navigate = useNavigate()
    const urlParams = new URLSearchParams(window.location.search);
    const namespace = urlParams.get('namespace');
    const group = urlParams.get('group');
    const requestedFile = urlParams.get('file');
    const requestedTemplateId = urlParams.get('templateId');

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
    const [templates, setTemplates] = useState<ConfigFileTemplate[]>([]);
    const selectedTemplateId = editState.activeNode?.data.resourceKind === 'template'
        ? String(editState.activeNode.data.templateId || '')
        : '';
    const selectedFileName = editState.activeNode?.data.resourceKind === 'template'
        ? undefined
        : editState.activeNode?.value as string | undefined;

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
        });
        describeConfigTemplates()
            .then((result) => setTemplates(result.templates))
            .catch((error) => {
                setTemplates([]);
                openErrNotification('获取配置模板列表失败', error instanceof Error ? error.message : String(error));
            });
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
        if (!requestedTemplateId || editState.visible) return;
        const requested = templates.find((item) => String(item.id) === requestedTemplateId);
        if (!requested) return;
        setEditState((prev) => ({
            ...prev,
            activeNode: {
                value: `template:${requested.id}`,
                data: {
                    value: `template:${requested.id}`,
                    resourceKind: 'template',
                    templateId: String(requested.id),
                },
            } as TreeNodeModel,
            visible: true,
            mode: 'view',
        }));
    }, [editState.visible, requestedTemplateId, templates]);

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
            {node.data.resourceKind !== 'template' && editState.activeNode && editState.activeNode.data.value === node.value && (
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
            <GroupWorkspaceNav
                group={group || ''}
                onBack={() => navigate('/configuration/group')}
            />
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
            <ResourceToolbar
                density="compact"
                title="配置清单"
                count={`共 ${datas.length + templates.length} 条 · 文件 ${datas.length} · 模板 ${templates.length}`}
                filters={(
                    <>
                        <Input
                            className={style.treeSearch}
                            placeholder="搜索文件路径"
                            value={editState.nodeFilter}
                            onChange={value => setEditState(s => ({ ...s, nodeFilter: value }))}
                        />
                        <Tooltip content="刷新配置文件">
                            <Button
                                aria-label="刷新配置文件"
                                shape="square"
                                variant="outline"
                                icon={<RefreshIcon />}
                                onClick={() => fetchData()}
                            />
                        </Tooltip>
                        <Tooltip content={activeGroup?.editable ? '新建配置' : '没有权限'}>
                            <Button
                                theme="primary"
                                icon={<AddIcon />}
                                onClick={() => handleOperateFile({} as TreeNodeModel, 'create')}
                                disabled={!activeGroup?.editable}
                            >
                                新建配置
                            </Button>
                        </Tooltip>
                    </>
                )}
            />
            <div className={style.workbench}>
                <aside className={style.fileExplorer}>
                    <div className={style.treeScroll}>
                    <Tree
                        ref={treeRef}
                        className={style.fileTree}
                        activable={true}
                        checkStrictly={true}
                        valueMode='onlyLeaf'
                        allowFoldNodeOnFilter={true}
                        data={renderTree(datas, templates)}
                        hover
                        expandAll={true}
                        icon={renderIcon}
                        scroll={{ type: 'virtual' }}
                        filter={treeSearch}
                        onActive={(node, ctx) => {
                            if (ctx.node.data.resourceKind === 'template') {
                                setEditState(s => ({
                                    ...s,
                                    activeNode: { ...ctx.node },
                                    visible: true,
                                    mode: 'view',
                                }));
                                const params = new URLSearchParams(window.location.search);
                                params.delete('file');
                                params.set('templateId', String(ctx.node.data.templateId || ''));
                                navigate(`${window.location.pathname}?${params.toString()}`, { replace: true });
                                return;
                            }
                            setEditState(s => ({ ...s, activeNode: { ...ctx.node }, visible: true, mode: 'edit' }));
                            dispatch(editorConfigFile({
                                namespace: namespace ? namespace : '',
                                group: group ? group : '',
                                name: ctx.node.value as string,
                            }));
                            const params = new URLSearchParams(window.location.search);
                            params.delete('templateId');
                            params.set('file', ctx.node.value as string);
                            navigate(`${window.location.pathname}?${params.toString()}`, { replace: true });
                        }}
                        operations={renderOperations}
                    />
                    </div>
                </aside>
                <main className={style.canvas}>
                    {selectedTemplateId && (
                        <TemplateWorkspace
                            key={selectedTemplateId}
                            embedded
                            templateId={selectedTemplateId}
                            namespace={namespace || ''}
                            group={group || ''}
                        />
                    )}
                    {!selectedTemplateId && editState.visible && editState.mode === 'edit' && (
                        <FileView
                            editable={activeGroup?.editable}
                            deleteable={activeGroup?.deleteable}
                            key={editState.activeNode?.value}
                            onAuthorize={() => setAuthorizeVisible(true)}
                        />
                    )}
                    {!selectedTemplateId && (!editState.visible || editState.mode !== 'edit') && (
                        <div className={style.emptyCanvas}>选择左侧配置文件或模板后查看详情</div>
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

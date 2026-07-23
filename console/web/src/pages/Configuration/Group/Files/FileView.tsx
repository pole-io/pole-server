import React from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { Button, Form, Input, Select, Space, Switch, Tag, Tabs, FormProps } from 'components/Fluent';

import { OperationButton } from 'components/OperationButton';
import { useAppDispatch, useAppSelector } from 'modules/store';
import { openErrNotification, openInfoNotification } from 'utils/notifition';
import CodeEditor from 'components/CodeEditor';
import { listConfigFileCryptoAlgos, listOneConfigFile, selectConfigFile, updateConfigFiles } from 'modules/configuration/file';
import { describeConfigFileEnvironments, FileStatusMap, type ConfigFileView } from 'services/config_files';
import LabelInput from 'components/LabelInput';
import { BotIcon, Edit1Icon, RocketIcon, RollbackIcon, SaveIcon } from 'components/Fluent/icons';
import PublishForm from '../Releases/PublishForm';
import ReleaseTable from '../Releases/ReleaseTable';
import SubscribeTable from './SubscribeTable';
import { Label, Op } from 'services/types';
import { resolveFileFormat } from 'utils/path';
import style from './index.module.less';
import EnvironmentResourceSwitcher from 'components/EnvironmentResourceSwitcher';

const { FormItem } = Form;
const { TabPanel } = Tabs;

interface IFileViewProps {
    editable?: boolean;
    deleteable?: boolean;
    onAuthorize?: () => void;
}

type ResourceTab = 'content' | 'basic' | 'release' | 'subscribe';

const FileView: React.FC<IFileViewProps> = (props) => {
    const [form] = Form.useForm();
    const dispatch = useAppDispatch();
    const navigate = useNavigate();
    const location = useLocation();

    const fileState = useAppSelector(selectConfigFile);
    const { editFile, viewFile, cryptoAlgos } = fileState;

    const [editorState, setEditorState] = React.useState<{
        model: Op;
        publishView: boolean;
    }>({ model: 'view', publishView: false });
    const [activeTab, setActiveTab] = React.useState<ResourceTab>('content');
    const [environmentFiles, setEnvironmentFiles] = React.useState<ConfigFileView[]>([]);

    const fetchOneFile = React.useCallback(() => {
        dispatch(listOneConfigFile({
            param: {
                namespace: editFile?.namespace,
                group: editFile?.group,
                name: editFile?.name,
            },
        }));
    }, [dispatch, editFile?.group, editFile?.name, editFile?.namespace]);

    React.useEffect(() => {
        dispatch(listConfigFileCryptoAlgos()).then((res) => {
            if (res.meta.requestStatus === 'rejected') {
                openErrNotification('获取加密算法列表失败', res?.payload as string);
            }
        });
        fetchOneFile();
    }, [dispatch, fetchOneFile]);

    React.useEffect(() => {
        setActiveTab('content');
        setEditorState({ model: 'view', publishView: false });
    }, [editFile?.group, editFile?.name, editFile?.namespace]);

    React.useEffect(() => {
        let active = true;
        const currentGroup = editFile?.group || '';
        const currentName = editFile?.name || '';
        if (!currentGroup || !currentName) return () => { active = false; };
        describeConfigFileEnvironments(currentGroup, currentName)
            .then((items) => {
                if (active) setEnvironmentFiles(items);
            })
            .catch(() => {
                if (active) setEnvironmentFiles([]);
            });
        return () => { active = false; };
    }, [editFile?.group, editFile?.name]);

    React.useEffect(() => {
        if (!viewFile) return;
        form.setFieldsValue({
            comment: viewFile.comment || '',
            encrypted: Boolean(viewFile.encrypted),
            encryptAlgo: viewFile.encryptAlgo || '',
            content: viewFile.content || '',
            file_tags: viewFile.tags || [],
        });
    }, [form, viewFile]);

    const onSubmit: FormProps['onSubmit'] = async (e) => {
        if (e.validateResult !== true) return;
        const updateData = {
            id: editFile?.id || 0,
            namespace: editFile?.namespace || '',
            group: editFile?.group || '',
            name: editFile?.name || '',
            format: resolveFileFormat(editFile?.name || ''),
            comment: form.getFieldValue('comment') as string,
            encrypted: form.getFieldValue('encrypted') as boolean,
            encryptAlgo: form.getFieldValue('encryptAlgo') as string,
            content: form.getFieldValue('content') as string,
            tags: form.getFieldValue('file_tags') as Label[],
        };

        const result = await dispatch(updateConfigFiles({ param: updateData }));
        if (result.meta.requestStatus !== 'fulfilled') {
            openErrNotification('请求错误', result?.payload as string);
            return;
        }
        openInfoNotification('请求成功', '配置文件已成功保存');
        fetchOneFile();
        setEditorState(prev => ({ ...prev, model: 'view' }));
    };

    const fileStatus = FileStatusMap?.[viewFile?.status as keyof typeof FileStatusMap];
    const fileStatusTheme = (fileStatus?.theme || 'default') as 'success' | 'danger' | 'default' | 'primary' | 'warning';
    const currentFileName = viewFile?.name || editFile?.name || '-';
    const currentNamespace = viewFile?.namespace || editFile?.namespace || '';
    const currentGroup = viewFile?.group || editFile?.group || '';
    const currentFormat = viewFile?.format || resolveFileFormat(editFile?.name || '') || 'text';
    const currentTags = viewFile?.tags || [];

    const handoffToAgent = () => {
        if (!currentNamespace || !currentGroup || !currentFileName) {
            openErrNotification('无法交给 Agent', '当前配置资源上下文不完整');
            return;
        }
        const params = new URLSearchParams({
            kind: 'config.file',
            namespace: currentNamespace,
            group: currentGroup,
            name: currentFileName,
            returnTo: `${location.pathname}${location.search}${location.hash}`,
        });
        navigate(`/agent?${params.toString()}`);
    };

    const cancelEdit = () => {
        if (viewFile) {
            form.setFieldsValue({
                comment: viewFile.comment || '',
                encrypted: Boolean(viewFile.encrypted),
                encryptAlgo: viewFile.encryptAlgo || '',
                content: viewFile.content || '',
                file_tags: viewFile.tags || [],
            });
        }
        setEditorState(prev => ({ ...prev, model: 'view' }));
    };

    const renderHeader = (
        <>
            <EnvironmentResourceSwitcher
                currentNamespace={currentNamespace}
                resourceLabel="配置文件"
                presentation="tabs"
                items={environmentFiles.map((item) => ({
                    namespace: item.namespace,
                    summary: FileStatusMap?.[item.status as keyof typeof FileStatusMap]?.text || '未发布',
                }))}
                onSelect={(nextNamespace) => {
                    const params = new URLSearchParams(window.location.search);
                    params.set('namespace', nextNamespace);
                    params.set('group', currentGroup);
                    params.set('file', currentFileName);
                    navigate(`${window.location.pathname}?${params.toString()}`);
                }}
            />
            <header className={style.fileSummary}>
                <div className={style.fileIdentity}>
                    <div className={style.fileTitleRow}>
                        <div className={`${style.currentFileName} ${style.mono}`}>{currentFileName}</div>
                        <Space size={6}>
                            <Tag variant="light" theme="primary">{currentFormat}</Tag>
                            <Tag theme={viewFile?.encrypted ? 'warning' : 'default'} variant="light">
                                {viewFile?.encrypted ? '已加密' : '未加密'}
                            </Tag>
                            <Tag theme={fileStatusTheme} variant="light">{fileStatus?.text || '未发布'}</Tag>
                        </Space>
                    </div>
                    <div className={style.resourcePath} aria-label="配置文件资源路径">
                        <span>{currentGroup || '-'}</span>
                        <i>/</i>
                        <strong className={style.mono}>{currentFileName}</strong>
                    </div>
                </div>
                <div className={style.fileActions}>
                    {editorState.model === 'view' && (
                        <Button variant="text" icon={<BotIcon />} onClick={handoffToAgent}>交给 Agent</Button>
                    )}
                    {editorState.model === 'view' && props.editable && (
                        <OperationButton action="authorize" label="授权" variant="outline" onClick={props.onAuthorize} />
                    )}
                    {editorState.model === 'view' && props.editable && (
                        <Button
                            variant="outline"
                            icon={<Edit1Icon />}
                            onClick={() => {
                                setEditorState(prev => ({ ...prev, model: 'edit' }));
                                setActiveTab('content');
                            }}
                        >
                            编辑
                        </Button>
                    )}
                    {editorState.model === 'edit' && props.editable && (
                        <>
                            <Button variant="outline" icon={<RollbackIcon />} onClick={cancelEdit}>撤销</Button>
                            <Button theme="primary" icon={<SaveIcon />} onClick={() => form.submit()}>保存草稿</Button>
                        </>
                    )}
                    {editorState.model === 'view' && props.editable && (
                        <Button theme="primary" icon={<RocketIcon />} onClick={() => setEditorState(prev => ({ ...prev, publishView: true }))}>
                            发布配置
                        </Button>
                    )}
                </div>
            </header>
        </>
    );

    const renderContent = (
        <div className={style.contentPane}>
            <div className={style.contentToolbar}>
                <div>
                    <strong>{editorState.model === 'edit' ? '编辑配置内容' : '当前配置内容'}</strong>
                    <span>{editorState.model === 'edit' ? '修改后保存为正式草稿' : '只读预览，支持全屏查看'}</span>
                </div>
                <Tag variant="light">{currentFormat} · {editorState.model === 'edit' ? '编辑态' : '只读'}</Tag>
            </div>
            <div className={style.editorShell}>
                <div className={style.editorBody}>
                    {editorState.model === 'edit' ? (
                        <FormItem name="content">
                            <CodeEditor allowFullScreen language={viewFile?.format} height="100%" />
                        </FormItem>
                    ) : (
                        <CodeEditor allowFullScreen language={viewFile?.format} value={viewFile?.content} readonly height="100%" />
                    )}
                </div>
                <div className={style.editorStatusBar}>
                    <span>UTF-8</span>
                    <span>LF / spaces: 2</span>
                </div>
            </div>
        </div>
    );

    const renderBasicInfo = (
        <div className={style.basicInfoPane}>
            {editorState.model === 'edit' ? (
                <div className={style.basicEditGrid}>
                    <FormItem
                        label="文件描述"
                        name="comment"
                        rules={[{ max: 1024, message: '长度不超过1024个字符' }]}
                    >
                        <Input />
                    </FormItem>
                    <FormItem label="配置加密" name="encrypted">
                        <Switch />
                    </FormItem>
                    <FormItem shouldUpdate={(prev, next) => prev.encrypted !== next.encrypted}>
                        {({ getFieldValue }) => getFieldValue('encrypted') === true ? (
                            <FormItem label="加密算法类型" name="encryptAlgo">
                                <Select options={cryptoAlgos.map(item => ({ label: item, value: item }))} />
                            </FormItem>
                        ) : <></>}
                    </FormItem>
                    <div className={style.basicTagsEditor}>
                        <LabelInput label="文件标签" name="file_tags" editable />
                    </div>
                </div>
            ) : (
                <div className={style.basicInfoGrid}>
                    <div className={`${style.basicField} ${style.basicDescription}`}>
                        <span>文件描述</span>
                        <strong>{viewFile?.comment || '暂无描述'}</strong>
                    </div>
                    <div className={style.basicField}>
                        <span>创建时间</span>
                        <strong className={style.mono}>{viewFile?.createTime || '-'}</strong>
                    </div>
                    <div className={style.basicField}>
                        <span>修改时间</span>
                        <strong className={style.mono}>{viewFile?.modifyTime || '-'}</strong>
                    </div>
                    <div className={style.basicField}>
                        <span>加密方式</span>
                        <strong>{viewFile?.encrypted ? (viewFile?.encryptAlgo || '已启用') : '未启用'}</strong>
                    </div>
                    <div className={`${style.basicField} ${style.basicTags}`}>
                        <span>文件标签 · {currentTags.length}</span>
                        <div className={style.tagList}>
                            {currentTags.map((item) => (
                                <Tag key={`${item.key}-${item.value}`} variant="outline">{item.key} : {item.value}</Tag>
                            ))}
                            {currentTags.length === 0 && <span className={style.fieldValue}>暂无标签</span>}
                        </div>
                    </div>
                </div>
            )}
        </div>
    );

    return (
        <Form form={form} layout="vertical" className={style.fileDetail} onSubmit={onSubmit}>
            {renderHeader}
            <Tabs className={style.resourceTabs} value={activeTab} onChange={(value) => setActiveTab(value as ResourceTab)}>
                <TabPanel value="content" label="文件内容">{renderContent}</TabPanel>
                <TabPanel value="basic" label="基本信息">{renderBasicInfo}</TabPanel>
                <TabPanel value="release" label="发布记录">
                    <div className={style.tablePane}>
                        <ReleaseTable
                            namespace={currentNamespace}
                            group={currentGroup}
                            filename={currentFileName}
                            editable={Boolean(props.editable)}
                            deleteable={Boolean(props.deleteable)}
                        />
                    </div>
                </TabPanel>
                <TabPanel value="subscribe" label="订阅查询">
                    <div className={style.tablePane}>
                        <SubscribeTable
                            namespace={currentNamespace}
                            group={currentGroup}
                            filename={currentFileName}
                            editable={Boolean(props.editable)}
                            deleteable={Boolean(props.deleteable)}
                        />
                    </div>
                </TabPanel>
            </Tabs>
            {editorState.publishView && (
                <PublishForm
                    namespace={currentNamespace}
                    group={currentGroup}
                    filename={currentFileName}
                    visible={editorState.publishView}
                    close={() => {
                        setEditorState(prev => ({ ...prev, publishView: false }));
                        fetchOneFile();
                    }}
                />
            )}
        </Form>
    );
};

export default React.memo(FileView);

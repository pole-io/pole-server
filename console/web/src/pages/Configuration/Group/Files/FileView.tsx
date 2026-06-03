import React, { } from 'react';
import { Form, Space, Descriptions, Input, Switch, Select, StickyTool, Tag, FormProps } from "tdesign-react";

import { useAppDispatch, useAppSelector } from 'modules/store';;
import { openErrNotification, openInfoNotification } from 'utils/notifition';
import CodeEditor from 'components/CodeEditor';
import DescriptionsItem from 'tdesign-react/es/descriptions/DescriptionsItem';
import { listConfigFileCryptoAlgos, listOneConfigFile, selectConfigFile, updateConfigFiles } from 'modules/configuration/file';
import { FileStatusMap } from 'services/config_files';
import LabelInput from 'components/LabelInput';
import { Edit1Icon, SaveIcon, RocketIcon, RollbackIcon } from 'tdesign-icons-react';
import StickyItem from 'tdesign-react/es/sticky-tool/StickyItem';
import PublishForm from '../Releases/PublishForm';
import { Label, Op } from 'services/types';
import { resolveFileFormat } from 'utils/path';

const { FormItem } = Form;

interface IFileViewProps {
    editable?: boolean;
    deleteable?: boolean;
}

const FileView: React.FC<IFileViewProps> = (props) => {
    const [form] = Form.useForm();
    const dispatch = useAppDispatch();

    const fileState = useAppSelector(selectConfigFile)
    const { editFile, viewFile, cryptoAlgos } = fileState;

    const [editorState, setEditorState] = React.useState<{
        model: Op;
        publishView: boolean;
    }>({ model: 'view', publishView: false });


    React.useEffect(() => {
        dispatch(listConfigFileCryptoAlgos())
            .then(res => {
                if (res.meta.requestStatus === 'rejected') {
                    openErrNotification('获取加密算法列表失败', res?.payload as string);
                }
            })
        fetchOneFile();
        return () => {

        }
    }, [editFile]);

    const fetchOneFile = () => {
        dispatch(listOneConfigFile({
            param: {
                namespace: editFile?.namespace,
                group: editFile?.group,
                name: editFile?.name,
            }
        }))
    }

    const onSubmit: FormProps['onSubmit'] = async (e) => {
        if (!e.validateResult) {
            return;
        }
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
        }

        const result = await dispatch(updateConfigFiles({ param: updateData }));
        if (result.meta.requestStatus !== 'fulfilled') {
            openErrNotification('请求错误', result?.payload as string);
        } else {
            openInfoNotification('请求成功', '配置文件已成功保存');
            // 重新获取文件信息
            dispatch(listOneConfigFile({
                param: {
                    namespace: editFile?.namespace,
                    group: editFile?.group,
                    name: editFile?.name,
                }
            }))
            setEditorState(prev => ({ ...prev, model: 'view' }));
        }
    }

    const renderView = (
        <>
            {editorState.model === 'view' && (
                <Space direction="vertical" style={{ width: '100%' }}>
                    <Descriptions
                        itemLayout="horizontal"
                        layout="horizontal"
                        size="small"
                        title={viewFile?.name}
                        column={3}
                    >
                        <DescriptionsItem label="发布状态">
                            <Tag theme={FileStatusMap?.[viewFile?.status as keyof typeof FileStatusMap]?.theme as "success" | "danger" | "default" | "primary" | "warning"} variant="outline">{FileStatusMap?.[viewFile?.status as keyof typeof FileStatusMap]?.text ?? '-'}</Tag>
                        </DescriptionsItem>
                        <DescriptionsItem label="文件格式">
                            {viewFile?.format}
                        </DescriptionsItem>
                        <DescriptionsItem label="修改时间">
                            {viewFile?.modifyTime}
                        </DescriptionsItem>
                        <DescriptionsItem label="创建时间">
                            {viewFile?.createTime}
                        </DescriptionsItem>
                        <DescriptionsItem label="加密状态">
                            {viewFile?.encrypted ? '已加密' : '未加密'}
                        </DescriptionsItem>
                        <DescriptionsItem label="加密算法">
                            {viewFile?.encryptAlgo}
                        </DescriptionsItem>
                        <DescriptionsItem label="文件标签">
                            <Space>
                                {viewFile?.tags?.map((item) => (
                                    <Tag variant='outline'>
                                        {item.key} : {item.value}
                                    </Tag>
                                ))}
                            </Space>
                        </DescriptionsItem>
                    </Descriptions>
                    <Space style={{ marginTop: 20, width: '100%' }}>
                        <CodeEditor allowFullScreen={true} language={viewFile?.format} value={viewFile?.content} readonly={true} />
                    </Space>
                </Space>
            )}
        </>
    )

    const renderPublish = (
        <>
            {editorState.publishView && (
                <PublishForm
                    namespace={editFile?.namespace || ''}
                    group={editFile?.group || ''}
                    filename={editFile?.name || ''}
                    visible={editorState.publishView}
                    close={() => {
                        setEditorState(prev => ({ ...prev, publishView: false }));
                        // 重新获取文件信息
                        fetchOneFile();
                    }}
                />
            )}
        </>
    )

    const renderEditor = (
        <>
            {editorState.model === 'edit' && (
                <Form
                    form={form}
                    layout="vertical"
                >
                    <FormItem
                        label="文件描述"
                        name="comment"
                        rules={[{ max: 1024, message: '长度不超过1024个字符' }]}>
                        <Input />
                    </FormItem>
                    <FormItem label="配置加密" name={'encrypted'}>
                        <Switch />
                    </FormItem>
                    <FormItem shouldUpdate={(prev, next) => {
                        const enableChange = prev.encrypted !== next.encrypted;
                        return enableChange;
                    }}>
                        {({ getFieldValue }) => {
                            if (getFieldValue('encrypted') === true) {
                                return (
                                    <FormItem label="加密算法类型" key="ice" name={'encryptAlgo'}>
                                        <Select options={cryptoAlgos.map(item => ({ label: item, value: item }))} />
                                    </FormItem>
                                );
                            }
                            return <></>;
                        }}
                    </FormItem>
                    <LabelInput label="文件标签" name="file_tags" editable={true} />
                    <Space style={{ marginTop: 20, width: '100%' }}>
                        <FormItem>
                            <CodeEditor allowFullScreen={true} language={viewFile?.format} />
                        </FormItem>
                    </Space>
                </Form>
            )}
        </>
    )

    const renderStickyTool = (
        <>
            {(props.editable || props.deleteable) && (
                <StickyTool
                    style={{ zIndex: 1000 }}
                    placement='right-bottom'
                    offset={[-10, 200]}
                >
                    {props.editable && (
                        <StickyItem
                            label={editorState.model === 'view' ? '编辑' : '保存'}
                            icon={editorState.model === 'view' ?
                                <Edit1Icon onClick={() => {
                                    if (editorState.model === 'view') {
                                        setEditorState(prev => ({ ...prev, model: 'edit' }));
                                    }
                                }} />
                                :
                                <SaveIcon onClick={() => {
                                    if (editorState.model === 'edit') {
                                        form.submit();
                                    }
                                }} />}
                        />
                    )}
                    {(editorState.model === 'edit' && props.editable) && (
                        <StickyItem label="撤销" icon={
                            <RollbackIcon onClick={() => {
                                setEditorState(prev => ({ ...prev, model: 'view' }));
                            }} />}
                        />
                    )}
                    {(editorState.model === 'view' && props.editable) && (
                        <StickyItem label="发布" icon={
                            <RocketIcon onClick={() => {
                                setEditorState(prev => ({ ...prev, publishView: true }));
                            }} />}
                        />
                    )}
                </StickyTool>
            )}
        </>
    )

    return (
        <>
            {renderView}
            {renderEditor}
            {renderPublish}
            {renderStickyTool}
        </>
    )
}

export default React.memo(FileView)
import React from 'react';
import { Drawer, Form, Input, Space, Button } from "tdesign-react";
import type { FormProps } from 'tdesign-react';
import { openErrNotification, openInfoNotification } from 'utils/notifition';
import { useAppDispatch, useAppSelector } from 'modules/store';
import { saveNamespace, updateNamespace, selectNamespace } from 'modules/namespace';
import LabelInput from 'components/LabelInput';
import { Label, Op } from 'services/types';
import style from './index.module.less';

const { FormItem } = Form;

interface NamespaceEditorProps {
    op: Op;
    closeDrawer: () => void;
    onAuthorize?: () => void;
    visible: boolean;
}

const NamespaceEditor: React.FC<NamespaceEditorProps> = ({ visible, op, closeDrawer, onAuthorize }) => {
    const [form] = Form.useForm();
    const [editable, setEditable] = React.useState(op !== 'view');

    const dispatch = useAppDispatch();
    const namespaceState = useAppSelector(selectNamespace);
    const { editNs } = namespaceState;

    React.useEffect(() => {
        if (visible) {
            setEditable(op !== 'view');
        }
    }, [visible, op]);

    React.useEffect(() => {
        if (visible && editNs) {
            form.setFieldsValue({
                name: editNs.name,
                comment: editNs.comment,
                namespace_labels: editNs?.metadata ? Object.entries(editNs.metadata).map(([key, value]) => ({ key, value } as Label)) : [],
            });
        }
    }, [visible, editNs]);

    const onSubmit: FormProps['onSubmit'] = async (e) => {
        console.log('onSubmit', e);
        if (e.validateResult !== true) {
            return;
        }

        const labels = form.getFieldValue('namespace_labels') as { key: string, value: string }[]
        const data = {
            name: form.getFieldValue('name') as string,
            comment: form.getFieldValue('comment') as string,
            metadata: labels.reduce((acc: { [key: string]: string }, { key, value }) => {
                acc[key] = value;
                return acc;
            }, {}),
        }
        let result;
        if (op === 'create') {
            result = await dispatch(saveNamespace({ param: data }))
        } else {
            result = await dispatch(updateNamespace({ param: data }))
        }
        if (result.meta.requestStatus !== 'fulfilled') {
            openErrNotification('请求错误', result?.payload as string);
        } else {
            openInfoNotification('请求成功', op === 'create' ? '创建命名空间成功' : '修改命名空间成功');
            closeDrawer();
        }
    };

    const labels = editNs?.metadata ? Object.entries(editNs.metadata) : [];

    const namespaceView = (
        <div className={style.namespaceDetail}>
            <div className={style.detailField}>
                <span>名称</span>
                <strong>{editNs?.name || '-'}</strong>
            </div>
            <div className={style.detailField}>
                <span>描述</span>
                <p>{editNs?.comment || '-'}</p>
            </div>
            <div className={style.detailField}>
                <span>标签</span>
                <div className={style.detailLabels}>
                    {labels.length > 0 ? labels.map(([key, value]) => (
                        <em key={key}>{key}: {value}</em>
                    )) : '-'}
                </div>
            </div>
        </div>
    );

    const namespaceForm = (
        <Form
            form={form}
            layout="vertical"
            // labelWidth={120}
            // labelAlign={'left'}
            onSubmit={onSubmit}
        >
            <FormItem
                label={'名称'}
                name={"name"}
                showErrorMessage={op === 'create'}
                rules={[
                    { required: true, message: '命名空间不能为空' },
                    { pattern: /^[a-zA-Z0-9._-]+$/, message: '只允许数字、英文字母、.、-、_' },
                    { max: 128, message: '长度不超过128个字符' }
                ]}>
                <Input readonly={op !== 'create'} />
            </FormItem>
            <FormItem
                label={'描述'}
                name={"comment"}
                rules={[{ max: 1024, message: '长度不超过1024个字符' }]}>
                <Input />
            </FormItem>
            <LabelInput form={form} label='标签' name='namespace_labels' editable={true} />
            <FormItem style={{ marginTop: 100 }}>
                <Space>
                    <Button type="submit" theme="primary">
                        提交
                    </Button>
                    <Button type="reset" theme="default" onClick={() => closeDrawer()}>
                        取消
                    </Button>
                </Space>
            </FormItem>
        </Form>
    )

    return (
        <div>
            <Drawer
                size='large'
                header={op === 'create' ? "创建命名空间" : editable ? "编辑命名空间" : "命名空间详情"}
                footer={op === 'view' && !editable ? (
                    <Space>
                        <Button theme="primary" onClick={() => setEditable(true)}>
                            编辑
                        </Button>
                        <Button theme="default" onClick={onAuthorize}>
                            授权
                        </Button>
                        <Button theme="default" onClick={closeDrawer}>
                            关闭
                        </Button>
                    </Space>
                ) : false}
                visible={visible}
                showOverlay={false}
                onClose={closeDrawer}
            >
                {editable ? namespaceForm : namespaceView}
            </Drawer>
        </div>
    );
}

export default React.memo(NamespaceEditor);

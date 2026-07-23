import React from 'react';
import { Drawer, Form, Input, Space, Button } from 'components/Fluent';
import type { FormProps } from 'components/Fluent';
import { openErrNotification, openInfoNotification } from 'utils/notifition';
import { useAppDispatch, useAppSelector } from 'modules/store';
import { saveNamespace, updateNamespace, selectNamespace } from 'modules/namespace';
import LabelInput from 'components/LabelInput';
import { Op } from 'services/types';
import style from './index.module.less';

const { FormItem } = Form;

interface NamespaceEditorProps {
    op: Op;
    closeDrawer: () => void;
    visible: boolean;
}

interface NamespaceLabelRow {
    key: string;
    value: string;
}

const NAMESPACE_NAME_REG = /^[a-zA-Z0-9._-]+$/;

const metadataToRows = (metadata?: Record<string, string>): NamespaceLabelRow[] => (
    Object.entries(metadata || {}).map(([key, value]) => ({ key, value }))
);

const ReadonlyField = ({ value }: { value?: string }) => (
    <div className={style.readonlyField}>{value || '-'}</div>
);

const NamespaceEditor: React.FC<NamespaceEditorProps> = ({ visible, op, closeDrawer }) => {
    const [form] = Form.useForm();
    const [editable, setEditable] = React.useState(op !== 'view');
    const [nameError, setNameError] = React.useState('');

    const dispatch = useAppDispatch();
    const namespaceState = useAppSelector(selectNamespace);
    const { editNs } = namespaceState;

    const watchedName = (Form.useWatch('name', form) as string | undefined) || '';
    const watchedComment = (Form.useWatch('comment', form) as string | undefined) || '';

    const initialValues = React.useCallback(() => ({
        name: editNs?.name || undefined,
        comment: editNs?.comment || '',
        namespace_labels: metadataToRows(editNs?.metadata),
    }), [editNs?.comment, editNs?.metadata, editNs?.name]);

    const validateNamespaceName = React.useCallback((name: string) => {
        const trimmedName = name.trim();
        if (!trimmedName) {
            return '命名空间不能为空';
        }
        if (!NAMESPACE_NAME_REG.test(trimmedName)) {
            return '只允许数字、英文字母、.、-、_';
        }
        if (trimmedName.length > 128) {
            return '长度不超过128个字符';
        }
        return '';
    }, []);

    const resetForm = React.useCallback(() => {
        form.setFieldsValue(initialValues());
        setNameError('');
    }, [form, initialValues]);

    React.useEffect(() => {
        if (visible) {
            setEditable(op !== 'view');
            resetForm();
        }
    }, [visible, op, resetForm]);

    const validateLabels = () => {
        const seenKeys = new Set<string>();
        const metadata: Record<string, string> = {};
        const labelRows = (form.getFieldValue('namespace_labels') as NamespaceLabelRow[] | undefined) || [];
        let firstError = '';

        labelRows.forEach((row) => {
            const key = row.key.trim();
            const value = row.value.trim();
            if (!key && !value) {
                return;
            }
            if (!key && value) {
                firstError = firstError || '标签键不能为空';
                return;
            }
            if (seenKeys.has(key)) {
                firstError = firstError || `标签键 ${key} 重复`;
                return;
            }
            seenKeys.add(key);
            metadata[key] = value;
        });

        return {
            metadata,
            firstError,
        };
    };

    const onSubmit: FormProps['onSubmit'] = async (e) => {
        if (e.validateResult !== true) {
            if (op === 'create') {
                setNameError(validateNamespaceName(watchedName));
            }
            return;
        }

        if (op === 'create') {
            const nextNameError = validateNamespaceName(watchedName);
            if (nextNameError) {
                setNameError(nextNameError);
                openErrNotification('校验失败', nextNameError);
                return;
            }
        }

        const labels = validateLabels();
        if (labels.firstError) {
            openErrNotification('校验失败', labels.firstError);
            return;
        }

        const data = {
            name: op === 'create' ? watchedName.trim() : editNs?.name || watchedName,
            comment: form.getFieldValue('comment') as string,
            metadata: labels.metadata,
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

    const namespaceForm = (
        <Form
            form={form}
            className={style.namespaceForm}
            layout="inline"
            labelWidth={104}
            labelAlign="right"
            onSubmit={onSubmit}
        >
            <div className={style.namespaceDrawerContent}>
                <section className={style.namespaceFormSection}>
                    <div className={style.namespaceFormSectionTitle}>基础信息</div>
                    <div id="inNsName">
                        <FormItem
                            label="名称"
                            name="name"
                            showErrorMessage={editable && op === 'create'}
                            rules={editable && op === 'create' ? [
                                { required: true, message: '命名空间不能为空' },
                                { pattern: NAMESPACE_NAME_REG, message: '只允许数字、英文字母、.、-、_' },
                                { max: 128, message: '长度不超过128个字符' }
                            ] : []}
                        >
                            {editable && op === 'create' ? (
                                <Input
                                    allowInputOverMax
                                    placeholder="允许数字、英文字母、.、-、_，限制128个字符"
                                    onChange={(value) => setNameError(validateNamespaceName(String(value)))}
                                />
                            ) : <ReadonlyField value={watchedName} />}
                        </FormItem>
                        {editable && op === 'create' && (
                            <>
                                <div className={style.nameMeta}>
                                    <span>命名后不可修改，请谨慎填写</span>
                                    <span className={watchedName.length > 128 ? style.nameCountError : undefined}>
                                        {watchedName.length}/128
                                    </span>
                                </div>
                                <div className={style.inlineError}>{nameError}</div>
                            </>
                        )}
                    </div>
                    <div id="inNsDesc">
                        <FormItem
                            label="描述"
                            name="comment"
                            rules={editable ? [{ max: 1024, message: '长度不超过1024个字符' }] : []}
                        >
                            {editable ? <Input placeholder="请输入该环境的用途或阶段说明" /> : <ReadonlyField value={watchedComment} />}
                        </FormItem>
                    </div>
                </section>

                <section className={style.namespaceFormSection}>
                    <div className={style.namespaceFormSectionTitle}>命名空间标签</div>
                    <LabelInput
                        form={form}
                        label="命名空间标签"
                        name="namespace_labels"
                        editable={editable}
                        disabled={!editable}
                        hideLabel
                        editorId="namespaceTagRows"
                        emptyId="namespaceTagsEmpty"
                        countId="namespaceTagCount"
                    />
                </section>

                {editable && (
                    <FormItem className={style.namespaceFormFooter}>
                        <Space>
                            <Button type="submit" theme="primary">
                                提交
                            </Button>
                            <Button theme="default" onClick={resetForm}>
                                重置
                            </Button>
                        </Space>
                    </FormItem>
                )}
            </div>
        </Form>
    )

    return (
        <div>
            <Drawer
                size="min(720px, 94vw)"
                header={op === 'create' ? "创建命名空间" : editable ? "编辑命名空间" : "命名空间详情"}
                footer={op === 'view' && !editable ? (
                    <Space>
                        <Button theme="primary" onClick={() => setEditable(true)}>
                            编辑
                        </Button>
                        <Button theme="default" onClick={closeDrawer}>
                            关闭
                        </Button>
                    </Space>
                ) : false}
                visible={visible}
                showOverlay={false}
                closeOnEscKeydown
                destroyOnClose
                onClose={closeDrawer}
            >
                {namespaceForm}
            </Drawer>
        </div>
    );
}

export default React.memo(NamespaceEditor);

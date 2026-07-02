import React from 'react';
import { Drawer, Form, Input, Select, Space, Button } from "tdesign-react";
import type { FormProps } from 'tdesign-react';
import LabelInput from 'components/LabelInput';
import { useAppDispatch, useAppSelector } from 'modules/store';
import { NamespaceView } from 'services/namespace';
import { ServiceView, describeAllServices } from 'services/service';
import { openErrNotification, openInfoNotification } from 'utils/notifition';
import { saveServices, selectService, updateServices } from 'modules/discovery/service';
import { Op } from 'services/types';
import { listAllNamespaces, selectNamespace } from 'modules/namespace';
import style from './index.module.less';

const { FormItem } = Form;
const SERVICE_NAME_REG = /^[0-9A-Za-z._-]+$/;

interface IServiceEditorProps {
    op: Op;
    visible: boolean;
    closeDrawer: () => void;
}

interface ServiceLabelRow {
    key: string;
    value: string;
}

const metadataToRows = (metadata?: Record<string, string>): ServiceLabelRow[] => (
    Object.entries(metadata || {}).map(([key, value]) => ({ key, value }))
);

const ServiceEditor: React.FC<IServiceEditorProps> = ({ visible, op, closeDrawer }) => {
    const [form] = Form.useForm();
    const dispatch = useAppDispatch();
    const [editable, setEditable] = React.useState(op !== 'view');
    const [allServices, setAllServices] = React.useState<ServiceView[]>([]);
    const [nameError, setNameError] = React.useState('');

    const namespaceState = useAppSelector(selectNamespace);
    const { datas: namespaceDatas } = namespaceState;

    const serviceState = useAppSelector(selectService);
    const { editSvc } = serviceState;

    const watchedName = (Form.useWatch('name', form) as string | undefined) || '';
    const watchedNamespace = (Form.useWatch('namespace', form) as string | undefined) || '';

    const initialValues = React.useCallback(() => ({
        name: editSvc?.name || undefined,
        namespace: editSvc?.namespace || undefined,
        comment: editSvc?.comment || '',
        department: editSvc?.department || '',
        business: editSvc?.business || '',
        service_labels: metadataToRows(editSvc?.metadata),
    }), [editSvc?.business, editSvc?.comment, editSvc?.department, editSvc?.name, editSvc?.namespace]);

    const validateServiceName = React.useCallback((name: string, namespace: string) => {
        const trimmedName = name.trim();
        if (!trimmedName) {
            return '请输入服务名称';
        }
        if (!SERVICE_NAME_REG.test(trimmedName)) {
            return '只允许数字、英文字母、.、-、_';
        }
        if (trimmedName.length > 128) {
            return '长度不超过128个字符';
        }
        if (op === 'create' && namespace) {
            const duplicated = allServices.some((item) => item.namespace === namespace && item.name === trimmedName);
            if (duplicated) {
                return '该命名空间下服务名已存在';
            }
        }
        return '';
    }, [allServices, op]);

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

    React.useEffect(() => {
        if (visible) {
            dispatch(listAllNamespaces())
                .then(res => {
                    if (res.meta.requestStatus === 'rejected') {
                        openErrNotification('获取命名空间列表失败', res?.payload as string);
                    }
                });
            describeAllServices()
                .then(setAllServices)
                .catch((error) => {
                    openErrNotification('获取服务列表失败', (error as Error).message);
                });
        }
    }, [visible]);

    const validateLabels = () => {
        const seenKeys = new Map<string, number>();
        const metadata: Record<string, string> = {};
        const labelRows = (form.getFieldValue('service_labels') as ServiceLabelRow[] | undefined) || [];
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
            const existedIndex = seenKeys.get(key);
            if (existedIndex !== undefined) {
                const message = `标签键 ${key} 重复`;
                firstError = firstError || message;
                return;
            }
            seenKeys.set(key, index);
            metadata[key] = value;
        });

        return {
            metadata,
            firstError,
        };
    };

    const onSubmit: FormProps['onSubmit'] = async (e) => {
        if (e.validateResult !== true) {
            const nextNameError = validateServiceName(watchedName, watchedNamespace);
            setNameError(nextNameError);
            return;
        }

        const nextNameError = validateServiceName(watchedName, watchedNamespace);
        if (nextNameError) {
            setNameError(nextNameError);
            openErrNotification('校验失败', nextNameError);
            return;
        }

        const labels = validateLabels();
        if (labels.firstError) {
            openErrNotification('校验失败', labels.firstError);
            return;
        }

        const newData = {
            id: editSvc?.id || '',
            name: watchedName.trim(),
            namespace: watchedNamespace,
            comment: form.getFieldValue('comment') as string,
            department: form.getFieldValue('department') as string,
            business: form.getFieldValue('business') as string,
            metadata: labels.metadata,
            ports: '',
            owners: '',
        };

        const result = op === 'create'
            ? await dispatch(saveServices({ param: { ...newData } }))
            : await dispatch(updateServices({ param: { ...newData } }));

        if (result.meta.requestStatus !== 'fulfilled') {
            openErrNotification('请求错误', result?.payload as string);
        } else {
            openInfoNotification('请求成功', op === 'create' ? '服务已创建' : '修改服务成功');
            closeDrawer();
        }
    };

    const serviceView = (
        <div className={style.serviceView}>
            <div className={style.viewSection}>
                <div className={style.viewGrid}>
                    <div className={style.viewField}>
                        <span className={style.viewLabel}>服务名</span>
                        <span className={style.viewValue}>{editSvc?.name || '-'}</span>
                    </div>
                    <div className={style.viewField}>
                        <span className={style.viewLabel}>命名空间</span>
                        <span className={style.viewValue}>{editSvc?.namespace || '-'}</span>
                    </div>
                    <div className={style.viewField}>
                        <span className={style.viewLabel}>部门</span>
                        <span className={style.viewValue}>{editSvc?.department || '-'}</span>
                    </div>
                    <div className={style.viewField}>
                        <span className={style.viewLabel}>业务</span>
                        <span className={style.viewValue}>{editSvc?.business || '-'}</span>
                    </div>
                    <div className={style.viewFieldFull}>
                        <span className={style.viewLabel}>描述</span>
                        <span className={style.viewValue}>{editSvc?.comment || '-'}</span>
                    </div>
                    <div className={style.viewFieldFull}>
                        <span className={style.viewLabel}>服务标签</span>
                        <div className={style.labelList}>
                            {editSvc?.metadata && Object.keys(editSvc.metadata).length > 0 ? (
                                Object.entries(editSvc.metadata).map(([key, value]) => (
                                    <span key={key}>{key}: {value}</span>
                                ))
                            ) : '-'}
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );

    const serviceForm = (
        <Form
            form={form}
            className={style.serviceForm}
            labelWidth={104}
            labelAlign="right"
            onSubmit={onSubmit}
        >
            <div id="drawer" className={style.serviceDrawerContent}>
                <section className={style.serviceFormSection}>
                    <div className={style.serviceFormSectionTitle}>基础信息</div>
                    <div id="inNs">
                        <FormItem
                            label="命名空间"
                            name="namespace"
                            rules={[{ required: true, message: '请选择命名空间' }]}
                        >
                            <Select
                                filterable
                                placeholder="请选择"
                                options={namespaceDatas.map((item: NamespaceView) => ({
                                    label: item.name,
                                    value: item.name,
                                }))}
                                onChange={(value) => {
                                    const nextError = validateServiceName(watchedName, String(value || ''));
                                    setNameError(watchedName ? nextError : '');
                                }}
                            />
                        </FormItem>
                    </div>
                    <div id="inName">
                        <FormItem
                            label="名称"
                            name="name"
                            rules={[
                                { required: true, message: '请输入服务名称' },
                                { pattern: SERVICE_NAME_REG, message: '只允许数字、英文字母、.、-、_' },
                                { max: 128, message: '长度不超过128个字符' },
                            ]}
                        >
                            <Input
                                allowInputOverMax
                                placeholder="允许数字、英文字母、.、-、_，限制128个字符"
                                onChange={(value) => {
                                    setNameError(validateServiceName(String(value), watchedNamespace));
                                }}
                            />
                        </FormItem>
                        <div className={style.nameMeta}>
                            <span>命名后不可修改，请谨慎填写</span>
                            <span
                                id="nameCount"
                                className={watchedName.length > 128 ? style.nameCountError : undefined}
                            >
                                {watchedName.length}/128
                            </span>
                        </div>
                        <div id="errName" className={style.inlineError}>
                            {nameError}
                        </div>
                    </div>
                    <div id="inDesc">
                        <FormItem
                            label="描述"
                            name="comment"
                            rules={[{ max: 1024, message: '长度不超过1024个字符' }]}
                        >
                            <Input placeholder="请输入服务描述" />
                        </FormItem>
                    </div>
                </section>

                <section className={style.serviceFormSection}>
                    <div className={style.serviceFormSectionTitle}>归属信息</div>
                    <div id="inDept">
                        <FormItem
                            label="部门"
                            name="department"
                            rules={[{ max: 255, message: '长度不超过255个字符' }]}
                        >
                            <Input placeholder="请输入部门" />
                        </FormItem>
                    </div>
                    <div id="inBiz">
                        <FormItem
                            label="业务"
                            name="business"
                            rules={[{ max: 255, message: '长度不超过255个字符' }]}
                        >
                            <Input placeholder="请输入业务" />
                        </FormItem>
                    </div>
                </section>

                <section className={style.serviceFormSection}>
                    <div className={style.serviceFormSectionTitle}>服务标签</div>
                    <LabelInput
                        form={form}
                        label="服务标签"
                        name="service_labels"
                        editable
                        hideLabel
                        editorId="tagRows"
                        emptyId="tagsEmpty"
                        countId="tagCount"
                    />
                </section>

                <FormItem className={style.serviceFormFooter}>
                    <Space>
                        <Button type="submit" theme="primary">
                            提交
                        </Button>
                        <Button theme="default" onClick={resetForm}>
                            重置
                        </Button>
                    </Space>
                </FormItem>
            </div>
        </Form>
    );

    return (
        <div>
            <Drawer
                size={op === 'view' ? '680px' : 'min(720px, 94vw)'}
                header={op === 'create' ? "创建服务" : editable ? "编辑服务" : "服务详情"}
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
                showOverlay
                closeOnOverlayClick
                closeOnEscKeydown
                destroyOnClose
                onClose={closeDrawer}
            >
                {editable ? serviceForm : serviceView}
            </Drawer>
        </div>
    );
};

export default React.memo(ServiceEditor);

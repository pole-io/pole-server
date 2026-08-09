import React, { } from 'react';
import { Steps, Drawer, Form, Input, Space, Row, Col, Switch, Select, FormProps, Button, Radio, RadioGroup } from 'components/Fluent';
import { FormItem } from 'components/Fluent'

import { useAppDispatch, useAppSelector } from 'modules/store';;
import CodeEditor from 'components/CodeEditor';
import LabelInput from 'components/LabelInput';
import { openErrNotification, openInfoNotification } from 'utils/notifition';
import { listConfigFileCryptoAlgos, saveConfigFiles, selectConfigFile } from 'modules/configuration/file';
import { resolveFileFormat } from 'utils/path';
import { Label, Op } from 'services/types';
import { ConfigType } from 'services/config_files';
import {
    ConfigFileTemplate,
    NamespaceTemplateValueRelease,
    describeNamespaceTemplateValueReleases,
    describeConfigTemplates,
} from 'services/config_templates';

interface IFileCreatorProps {
    op: Op;
    namespace: string;
    group: string;
    visible: boolean;
    closeDrawer: () => void;
}

const { StepItem } = Steps;

interface ConfigFileMetaValues {
    name: string;
    comment: string;
    tags: Label[];
    encrypted: boolean;
    encryptAlgo: string;
    configType: ConfigType;
    templateId: string;
    templateReleaseId: string;
}

const emptyMetaValues = (): ConfigFileMetaValues => ({
    name: '',
    comment: '',
    tags: [],
    encrypted: false,
    encryptAlgo: '',
    configType: 'CONFIG_FILE',
    templateId: '',
    templateReleaseId: '',
});

const FileCreator: React.FC<IFileCreatorProps> = ({ op, namespace, group, visible, closeDrawer }) => {
    const dispatch = useAppDispatch();
    const [form] = Form.useForm();

    const fileState = useAppSelector(selectConfigFile);
    const { cryptoAlgos } = fileState;

    const [activeStep, setActiveStep] = React.useState<number>(1);
    const [metaValues, setMetaValues] = React.useState<ConfigFileMetaValues>(emptyMetaValues);
    const [templates, setTemplates] = React.useState<ConfigFileTemplate[]>([]);
    const [environmentReleases, setEnvironmentReleases] = React.useState<NamespaceTemplateValueRelease[]>([]);

    const refreshTemplates = React.useCallback(() => {
        describeConfigTemplates()
            .then(({ templates: items }) => setTemplates(items))
            .catch(() => setTemplates([]));
    }, []);
    const refreshEnvironmentReleases = React.useCallback((templateId: string) => {
        if (!templateId) {
            setEnvironmentReleases([]);
            return;
        }
        describeNamespaceTemplateValueReleases(namespace, templateId)
            .then(({ releases }) => {
                setEnvironmentReleases(releases);
                const active = releases.find(item => item.active && item.releaseType === 'TEMPLATE_VALUE_RELEASE_NORMAL');
                setMetaValues(current => current.templateId === templateId ? {
                    ...current,
                    templateReleaseId: active?.templateReleaseId || '',
                } : current);
            })
            .catch(() => setEnvironmentReleases([]));
    }, [namespace]);

    React.useEffect(() => {
        if (visible) {
            const initialMetaValues = emptyMetaValues();
            setActiveStep(1);
            setMetaValues(initialMetaValues);
            form.setFieldsValue({
                namespace: namespace,
                group: group,
                ...initialMetaValues,
                content: '',
            });
            dispatch(listConfigFileCryptoAlgos())
                .then(res => {
                    if (res.meta.requestStatus === 'rejected') {
                        openErrNotification('获取加密算法列表失败', res?.payload as string);
                    }
                })
            refreshTemplates();
        }
    }, [visible, namespace, group, form, dispatch, refreshTemplates]);

    React.useEffect(() => {
        if (!visible) return;
        const refreshWhenReturning = () => {
            if (document.visibilityState !== 'visible') return;
            refreshTemplates();
            refreshEnvironmentReleases(metaValues.templateId);
        };
        window.addEventListener('focus', refreshWhenReturning);
        document.addEventListener('visibilitychange', refreshWhenReturning);
        return () => {
            window.removeEventListener('focus', refreshWhenReturning);
            document.removeEventListener('visibilitychange', refreshWhenReturning);
        };
    }, [metaValues.templateId, refreshEnvironmentReleases, refreshTemplates, visible]);

    React.useEffect(() => {
        refreshEnvironmentReleases(metaValues.templateId);
    }, [metaValues.templateId, refreshEnvironmentReleases]);

    React.useEffect(() => {
        if (visible && activeStep === 1) {
            form.setFieldsValue({
                namespace: namespace,
                group: group,
                ...metaValues,
            });
        }
    }, [activeStep, visible, namespace, group, metaValues, form]);

    const collectMetaValues = () => {
        const name = form.getFieldValue('name') as string;
        const comment = form.getFieldValue('comment') as string;
        const tags = form.getFieldValue('tags') as Label[];
        const encrypted = form.getFieldValue('encrypted') as boolean;
        const encryptAlgo = form.getFieldValue('encryptAlgo') as string;
        const configType = (form.getFieldValue('configType') || 'CONFIG_FILE') as ConfigType;
        const nextMetaValues = {
            name: name || '',
            comment: comment || '',
            tags: tags || [],
            encrypted: Boolean(encrypted),
            encryptAlgo: encryptAlgo || '',
            configType,
            templateId: metaValues.templateId,
            templateReleaseId: metaValues.templateReleaseId,
        };
        setMetaValues(nextMetaValues);
        return nextMetaValues;
    };

    const changeStep = (value: number) => {
        if (activeStep === 1 && value === 2) {
            collectMetaValues();
        }
        setActiveStep(value);
    };

    const onSubmit: FormProps['onSubmit'] = async (e) => {
        if (e.validateResult !== true) {
            return;
        }
        if (metaValues.configType === 'CONFIG_TEMPLATE' && (!metaValues.templateId || !metaValues.templateReleaseId)) {
            openErrNotification('无法创建', '所选模板必须先在当前环境发布一个配置版本');
            changeStep(1);
            return;
        }

        const newData = {
            namespace: namespace,
            group: group,
            name: metaValues.name,
            comment: metaValues.comment,
            format: resolveFileFormat(metaValues.name),
            content: form.getFieldValue('content') as string,
            tags: metaValues.tags,
            encrypted: metaValues.encrypted,
            encryptAlgo: metaValues.encrypted ? metaValues.encryptAlgo : '',
            configType: metaValues.configType,
            templateBinding: metaValues.configType === 'CONFIG_TEMPLATE' ? {
                templateId: metaValues.templateId,
                templateReleaseId: metaValues.templateReleaseId,
            } : undefined,
        }

        const result = await dispatch(saveConfigFiles({ param: { ...newData } }));
        if (result.meta.requestStatus !== 'fulfilled') {
            openErrNotification('请求错误', result?.payload as string);
        } else {
            openInfoNotification('请求成功', '创建配置文件成功');
            closeDrawer();
        }
    }

    const createForm = (
        <>
            <Form
                form={form}
                layout="vertical"
                // labelWidth={120}
                // labelAlign={'left'}
                onSubmit={onSubmit}
            >
                {activeStep === 1 && (
                    <>
                        <FormItem label="命名空间" name="namespace">
                            <Input disabled={true} />
                        </FormItem>
                        <FormItem label="配置组" name="group">
                            <Input disabled={true} />
                        </FormItem>
                        <FormItem label="内容来源" name="configType">
                            <RadioGroup
                                value={metaValues.configType}
                                onChange={(configType: ConfigType) => {
                                    form.setFieldsValue({ configType });
                                    setMetaValues(current => ({
                                        ...current,
                                        configType,
                                        templateId: configType === 'CONFIG_TEMPLATE' ? current.templateId : '',
                                        templateReleaseId: configType === 'CONFIG_TEMPLATE' ? current.templateReleaseId : '',
                                    }));
                                }}
                            >
                                <Radio value="CONFIG_FILE">直接文本</Radio>
                                <Radio value="CONFIG_TEMPLATE">配置模板</Radio>
                            </RadioGroup>
                            <div>
                                {metaValues.configType === 'CONFIG_TEMPLATE'
                                    ? '绑定逻辑模板，并使用当前环境生效的模板与 Value 组合版本。'
                                    : '直接编辑并发布 YAML、JSON、TOML 或文本内容。'}
                            </div>
                        </FormItem>
                        <FormItem>
                            {({ getFieldValue }) => {
                                return (
                                    <FormItem label="文件名称" name="name">
                                        <Input suffix={`文件格式: ${resolveFileFormat(getFieldValue('name') as string)}`} />
                                    </FormItem>
                                )
                            }}
                        </FormItem>
                        <FormItem label="文件描述" name="comment">
                            <Input />
                        </FormItem>
                        {metaValues.configType === 'CONFIG_TEMPLATE' && (
                            <>
                                <FormItem label="配置模板">
                                    <Select
                                        value={metaValues.templateId}
                                        placeholder="选择模板"
                                        options={templates.map(item => ({ label: item.name, value: String(item.id) }))}
                                        onChange={(templateId: string) => setMetaValues(current => ({
                                            ...current,
                                            templateId,
                                            templateReleaseId: '',
                                        }))}
                                    />
                                </FormItem>
                                <FormItem label="当前环境配置版本">
                                    <div>
                                        {environmentReleases.find(item => item.active && item.releaseType === 'TEMPLATE_VALUE_RELEASE_NORMAL')
                                            ? `v${environmentReleases.find(item => item.active && item.releaseType === 'TEMPLATE_VALUE_RELEASE_NORMAL')?.version} · 模板与 Value 已绑定`
                                            : '当前环境尚无生效的全量配置版本'}
                                    </div>
                                </FormItem>
                                <Button
                                    variant="text"
                                    onClick={() => {
                                        collectMetaValues();
                                        window.open(
                                            `/configuration/group/templates?group=${encodeURIComponent(group)}`
                                            + `&templateId=${encodeURIComponent(metaValues.templateId)}`
                                            + `&namespace=${encodeURIComponent(namespace)}&tab=values`
                                            + `&returnTo=${encodeURIComponent(window.location.pathname + window.location.search)}`,
                                            '_blank',
                                            'noopener,noreferrer',
                                        );
                                    }}
                                >
                                    在新标签页管理模板与当前 Namespace Value
                                </Button>
                            </>
                        )}
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
                                            <Select options={cryptoAlgos.map(alg => ({ label: alg, value: alg }))} />
                                        </FormItem>
                                    );
                                }
                                return <></>;

                            }}
                        </FormItem>
                        <LabelInput editable={true} label="文件标签" name="tags" />
                    </>
                )}
                {activeStep === 2 && metaValues.configType === 'CONFIG_FILE' && (
                    <>
                        <Space style={{ marginTop: 20, width: '100%' }}>
                            <FormItem name={'content'} style={{ width: '100%' }}>
                                <CodeEditor
                                    allowFullScreen={true}
                                    readonly={false}
                                    language={resolveFileFormat(metaValues.name)}
                                />
                            </FormItem>
                        </Space>
                    </>
                )}
                {activeStep === 2 && metaValues.configType === 'CONFIG_TEMPLATE' && (
                    <div style={{ marginTop: 20 }}>
                        配置文件将绑定逻辑模板，并使用 <strong>{namespace}</strong> 当前命中的完整环境配置版本；
                        SDK 将根据其中绑定的模板快照与 Value 快照完成渲染。
                    </div>
                )}
            </Form>
        </>
    )
    return (
        <Drawer
            header={op === 'edit' ? "编辑" : "创建"}
            size='960px'
            visible={visible}
            onClose={() => {
                changeStep(1);
                setMetaValues(emptyMetaValues());
                closeDrawer();
            }}
            footer={
                activeStep === 1 ? (
                    <>
                        <Space>
                            <Button theme="default" onClick={() => {
                                changeStep(2);
                            }}>
                                下一步
                            </Button>
                        </Space>
                    </>
                )
                    :
                    activeStep === 2 ? (
                        <>
                            <Space>
                                <Button theme="default" onClick={() => {
                                    changeStep(1);
                                }}>
                                    上一步
                                </Button>
                                <Button theme='primary' type='submit' onClick={() => {
                                    form.submit({ showErrorMessage: true });
                                }}>
                                    提交
                                </Button>
                            </Space>
                        </>
                    )
                        :
                        <>
                        </>
            }
        >
            <Row>
                <Col span={3}>
                    <Steps layout="vertical" current={activeStep} onChange={(value: number) => {
                        changeStep(value as number);
                    }}>
                        <StepItem value={1} title="基本信息">
                        </StepItem>
                        <StepItem
                            value={2}
                            title={metaValues.configType === 'CONFIG_TEMPLATE' ? '模板确认' : '文本内容'}
                            style={{ width: '80' }}
                        >
                        </StepItem>
                    </Steps>
                </Col>
                <Col span={9}>
                    {createForm}
                </Col>
            </Row>
        </Drawer>
    )
}

export default React.memo(FileCreator)

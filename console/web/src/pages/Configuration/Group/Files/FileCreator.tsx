import React, { } from 'react';
import { Steps, Drawer, Form, Input, Space, Row, Col, Switch, Select, FormProps, Button } from 'components/Fluent';
import { FormItem } from 'components/Fluent'

import { useAppDispatch, useAppSelector } from 'modules/store';;
import CodeEditor from 'components/CodeEditor';
import LabelInput from 'components/LabelInput';
import { openErrNotification, openInfoNotification } from 'utils/notifition';
import { listConfigFileCryptoAlgos, saveConfigFiles, selectConfigFile } from 'modules/configuration/file';
import { resolveFileFormat } from 'utils/path';
import { Label, Op } from 'services/types';

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
}

const emptyMetaValues = (): ConfigFileMetaValues => ({
    name: '',
    comment: '',
    tags: [],
    encrypted: false,
    encryptAlgo: '',
});

const FileCreator: React.FC<IFileCreatorProps> = ({ op, namespace, group, visible, closeDrawer }) => {
    const dispatch = useAppDispatch();
    const [form] = Form.useForm();

    const fileState = useAppSelector(selectConfigFile);
    const { cryptoAlgos } = fileState;

    const [activeStep, setActiveStep] = React.useState<number>(1);
    const [metaValues, setMetaValues] = React.useState<ConfigFileMetaValues>(emptyMetaValues);

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
        }
    }, [visible, namespace, group, form, dispatch]);

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
        const nextMetaValues = {
            name: name || '',
            comment: comment || '',
            tags: tags || [],
            encrypted: Boolean(encrypted),
            encryptAlgo: encryptAlgo || '',
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
                {activeStep === 2 && (
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
                    <Steps layout="vertical" current={activeStep} onChange={(value) => {
                        changeStep(value as number);
                    }}>
                        <StepItem value={1} title="元信息">
                        </StepItem>
                        <StepItem value={2} title="文件内容" style={{ width: '80' }}>
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

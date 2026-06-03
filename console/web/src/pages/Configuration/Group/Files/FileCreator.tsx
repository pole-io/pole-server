import React, { } from 'react';
import { Steps, Drawer, Form, Input, Space, Row, Col, Switch, Select, FormProps, Button } from "tdesign-react";
import FormItem from 'tdesign-react/es/form/FormItem'

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

const FileCreator: React.FC<IFileCreatorProps> = ({ op, namespace, group, visible, closeDrawer }) => {
    const dispatch = useAppDispatch();
    const [form] = Form.useForm();

    const fileState = useAppSelector(selectConfigFile);
    const { editFile, cryptoAlgos } = fileState;

    const [activeStep, setActiveStep] = React.useState<number>(1);

    React.useEffect(() => {
        if (visible) {
            dispatch(listConfigFileCryptoAlgos())
                .then(res => {
                    if (res.meta.requestStatus === 'rejected') {
                        openErrNotification('获取加密算法列表失败', res?.payload as string);
                    }
                })
        }
    }, [visible]);

    const onSubmit: FormProps['onSubmit'] = async (e) => {
        if (e.validateResult !== true) {
            return;
        }

        const newData = {
            namespace: namespace,
            group: group,
            name: form.getFieldValue('name') as string,
            comment: form.getFieldValue('comment') as string,
            format: resolveFileFormat(form.getFieldValue('name') as string),
            content: form.getFieldValue('content') as string,
            tags: form.getFieldValue('tags') as Label[] || [],
            encrypted: form.getFieldValue('encrypted') as boolean,
            encryptAlgo: form.getFieldValue('encryptAlgo') as string,
        }

        console.log('newData', newData);

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
                            <FormItem>
                                {({ getFieldValue }) => {
                                    return (
                                        <FormItem name={'content'} style={{ width: '100%' }}>
                                            <CodeEditor
                                                allowFullScreen={true}
                                                readonly={false}
                                                language={resolveFileFormat(getFieldValue('name') as string)}
                                            />
                                        </FormItem>
                                    )
                                }}
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
            style={{ width: '100%' }}
            visible={visible}
            onClose={() => {
                setActiveStep(1);
                closeDrawer();
            }}
            footer={
                activeStep === 1 ? (
                    <>
                        <Space>
                            <Button theme="default" onClick={() => {
                                setActiveStep(2);
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
                                    setActiveStep(1);
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
                        setActiveStep(value as number);
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
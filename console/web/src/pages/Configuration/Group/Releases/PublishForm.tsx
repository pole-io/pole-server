import React, { useState } from 'react';
import { Button, CustomValidator, Drawer, Form, FormProps, Input, Radio, RadioGroup, Space, Steps } from 'tdesign-react';

import CodeDiffEditor from 'components/CodeDiffEditor';
import { openErrNotification, openInfoNotification } from 'utils/notifition';
import { useAppDispatch, useAppSelector } from 'modules/store';
import { listConfigFileReleases, listOneConfigFileRelease, publishConfigFiles, selectFileRelease } from 'modules/configuration/release';
import ClientLabelInput from 'components/ClientLabelInput';
import { MatcheLabel } from 'services/types';
import { listOneConfigFile, selectConfigFile } from 'modules/configuration/file';

const { StepItem } = Steps;
const { FormItem } = Form;

export interface IPublishFormProps {
    namespace: string;
    group: string;
    filename: string;
    visible: boolean;
    close: () => void;
}

const PublishForm: React.FC<IPublishFormProps> = ({ namespace, group, filename, visible, close }) => {
    const [form] = Form.useForm();
    const dispatch = useAppDispatch();

    const fileState = useAppSelector(selectConfigFile)
    const { editFile, viewFile } = fileState;

    const releaseState = useAppSelector(selectFileRelease);
    const { versions = [], total = 0, loading, viewFileRelease } = releaseState;

    // 合并编辑相关状态
    const [activeStep, setActiveStep] = React.useState(1);

    React.useEffect(() => {
        if (visible) {
            // 重置表单
            setActiveStep(1);
            handleFetch();
        }
    }, [namespace, group, filename]);

    const versionValidator: CustomValidator = (val) => {
        const releaseName = form.getFieldValue("name") as string;
        const exist = versions.map(item => item.name).includes(releaseName);
        if (exist) {
            return {
                result: false,
                type: 'warning',
                message: `当前配置版本已存在，如果继续发布则会重新发布 ${releaseName} 版本对应的内容`,
            };
        }
        return { result: true, message: '' };
    };

    const onSubmit: FormProps['onSubmit'] = async (e) => {
        if (e.validateResult !== true) {
            return;
        }
        // 提交发布
        const pubData = {
            namespace: namespace,
            group: group,
            fileName: filename,
            name: form.getFieldValue("name") as string,
            releaseDescription: form.getFieldValue("comment") as string,
            releaseType: form.getFieldValue("releaseType") as string,
            betaLabels: form.getFieldValue("betaLabels") as MatcheLabel[],
        }

        const ret = await dispatch(publishConfigFiles({ param: pubData }));
        if (ret.meta.requestStatus === 'fulfilled') {
            setActiveStep(1);
            openInfoNotification('请求成功', '发布配置成功');
            close();
        } else {
            openErrNotification('请求失败', ret.payload as string);
            return;
        }
    }

    // 提交发布, 获取当前配置的最新数据以及当前处于使用中的配置发布信息
    const handleFetch = () => {
        dispatch(listConfigFileReleases({
            param: {
                namespace: namespace,
                group: group,
                file_name: filename,
            }
        })).then((res) => {
            if (res.meta.requestStatus === 'rejected') {
                openErrNotification('获取配置发布版本列表失败', res.payload as string);
            }
        })

        dispatch(listOneConfigFile({
            param: {
                namespace: namespace,
                group: group,
                name: filename,
            }
        })).then((res) => {
            if (res.meta.requestStatus === 'rejected') {
                openErrNotification('获取配置文件信息失败', res.payload as string);
            }
        })

        dispatch(listOneConfigFileRelease({
            param: {
                namespace: namespace,
                group: group,
                file_name: filename,
            }
        })).then((res) => {
            if (res.meta.requestStatus === 'rejected') {
                openErrNotification('获取配置文件发布信息失败', res.payload as string);
            }
        })
    }

    const renderForm = (
        <>
            <FormItem
                label="版本名称"
                name="name"
                rules={[
                    { required: true, message: '版本名称不能为空' },
                    { max: 64, message: '长度不超过64个字符' },
                    { validator: versionValidator },
                ]}
            >
                <Input />
            </FormItem>
            <FormItem
                label="版本描述"
                name="comment"
                rules={[
                    { max: 255, message: '长度不超过255个字符' }
                ]}
            >
                <Input />
            </FormItem>
            <FormItem label='发布类型' name='releaseType' initialData={'normal'}>
                <RadioGroup>
                    <Radio value="normal">全量发布</Radio>
                    <Radio value="gray">灰度发布</Radio>
                </RadioGroup>
            </FormItem>
            <FormItem shouldUpdate={(prev, next) => {
                const enableChange = prev.releaseType !== next.releaseType;
                return enableChange;
            }}>
                {({ getFieldValue }) => {
                    if (getFieldValue('releaseType') === 'gray') {
                        return (
                            <ClientLabelInput form={form} label='客户端标签' name='betaLabels' disabled={false} />
                        );
                    }
                    return <></>
                }}
            </FormItem>
            <FormItem>
                <Button theme='primary' type='submit' style={{ marginTop: 20 }}>
                    提交
                </Button>
            </FormItem>
        </>
    )

    const renderFooter = (
        <>
            {activeStep === 1 ? (
                <>
                    <Space>
                        <Button theme="default" onClick={() => {
                            setActiveStep(2);
                        }}>
                            下一步
                        </Button>
                    </Space>
                </>
            ) : activeStep === 2 ? (
                <>
                    <Space>
                        <Button theme="default" onClick={() => {
                            setActiveStep(1);
                        }}>
                            上一步
                        </Button>
                    </Space>
                </>
            ) : null}
        </>
    )

    return (
        <>
            <Drawer
                header={'配置发布'}
                size='960px'
                style={{ width: '100%' }}
                visible={visible}
                placement="right"
                onClose={() => {
                    close();
                }}
                footer={renderFooter}
            >
                <Steps current={activeStep}>
                    <StepItem value={1} title="版本对比" />
                    <StepItem value={2} title="发布信息" />
                </Steps>
                <Form
                    form={form}
                    labelWidth={100}
                    style={{ marginTop: 20 }}
                    onSubmit={onSubmit}
                >
                    {activeStep === 1 && (
                        <CodeDiffEditor
                            key={`${namespace}-${group}-${filename}`}
                            namespace={namespace}
                            group={group}
                            filename={filename}
                            curValue={viewFileRelease?.content}
                            nextValue={viewFile?.content}
                        />
                    )}
                    {activeStep === 2 && (
                        <>
                            {renderForm}
                        </>
                    )}
                </Form>
            </Drawer>
        </>
    )
}

export default React.memo(PublishForm);
import React, { useState } from 'react';
import { Button, CustomValidator, Drawer, Form, FormProps, Input, InputNumber, Radio, RadioGroup, Space, Steps } from 'components/Fluent';

import CodeDiffEditor from 'components/CodeDiffEditor';
import { openErrNotification, openInfoNotification } from 'utils/notifition';
import { useAppDispatch, useAppSelector } from 'modules/store';
import { listConfigFileReleases, listOneConfigFileRelease, publishConfigFiles, selectFileRelease } from 'modules/configuration/release';
import { listOneConfigFile, selectConfigFile } from 'modules/configuration/file';
import style from '../Files/index.module.less';
import GrayRuleEditor, { defaultGrayRuleRow, grayRowsToBetaLabels } from './GrayRuleEditor';

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
    const [grayRows, setGrayRows] = React.useState([defaultGrayRuleRow()]);

    React.useEffect(() => {
        if (visible) {
            // 重置表单
            setActiveStep(1);
            setGrayRows([defaultGrayRuleRow()]);
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
        const releaseType = form.getFieldValue("releaseType") as string;
        const betaLabels = grayRowsToBetaLabels(grayRows);
        if (releaseType === 'gray' && betaLabels.length === 0) {
            openErrNotification('请求失败', '灰度发布至少需要一条有效灰度规则');
            return;
        }
        // 提交发布
        const pubData = {
            namespace: namespace,
            group: group,
            fileName: filename,
            name: form.getFieldValue("name") as string,
            releaseDescription: form.getFieldValue("comment") as string,
            releaseType,
            grayPriority: form.getFieldValue("grayPriority") as number,
            betaLabels: releaseType === 'gray' ? betaLabels : [],
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
            <section className={style.drawerSection}>
                <div className={style.drawerSectionTitle}>版本信息</div>
                <div className={style.drawerSectionBody}>
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
                        label="发布说明"
                        name="comment"
                        rules={[
                            { max: 255, message: '长度不超过255个字符' }
                        ]}
                    >
                        <Input />
                    </FormItem>
                </div>
            </section>
            <section className={style.drawerSection}>
                <div className={style.drawerSectionTitle}>发布范围</div>
                <div className={style.drawerSectionBody}>
                    <FormItem label='发布类型' name='releaseType' initialData={'normal'}>
                        <RadioGroup>
                            <Radio value="normal">全量发布</Radio>
                            <Radio value="gray">灰度发布</Radio>
                        </RadioGroup>
                    </FormItem>
                    <FormItem shouldUpdate={(prev, next) => prev.releaseType !== next.releaseType}>
                        {({ getFieldValue }) => {
                            if (getFieldValue('releaseType') === 'gray') {
                                return (
                                    <>
                                        <FormItem label="灰度优先级" name="grayPriority" initialData={100}>
                                            <InputNumber theme="normal" min={1} max={9999} />
                                        </FormItem>
                                        <GrayRuleEditor rows={grayRows} editable={true} onChange={setGrayRows} />
                                    </>
                                );
                            }
                            return (
                                <div className={style.fieldValue}>
                                    全量发布会替换当前全量基线，不结束正在生效的灰度版本。
                                </div>
                            )
                        }}
                    </FormItem>
                </div>
            </section>
        </>
    )

    const renderFooter = (
        <Space>
            <Button theme="default" onClick={close}>取消</Button>
            {activeStep === 2 && (
                <Button theme="default" onClick={() => setActiveStep(1)}>
                    上一步
                </Button>
            )}
            {activeStep === 1 && (
                <Button theme="primary" onClick={() => setActiveStep(2)}>
                    下一步
                </Button>
            )}
            {activeStep === 2 && (
                <Button theme="primary" onClick={() => form.submit()}>
                    确认发布
                </Button>
            )}
        </Space>
    )

    return (
        <>
            <Drawer
                header={'发布配置'}
                size='960px'
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
                        <section className={style.drawerSection}>
                            <div className={style.drawerSectionTitle}>版本对比</div>
                            <div className={style.drawerSectionBody}>
                                <CodeDiffEditor
                                    key={`${namespace}-${group}-${filename}`}
                                    namespace={namespace}
                                    group={group}
                                    filename={filename}
                                    curValue={viewFileRelease?.content}
                                    nextValue={viewFile?.content}
                                />
                            </div>
                        </section>
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

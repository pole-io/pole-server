import React from 'react';
import { Button, Drawer, Form, FormProps, Input, Radio, RadioGroup, Space, Steps } from 'tdesign-react';

import { openErrNotification, openInfoNotification } from 'utils/notifition';
import { useAppDispatch } from 'modules/store';
import ClientLabelInput from 'components/ClientLabelInput';
import { MatcheLabel, RuleRelease } from 'services/types';
import { PolicySourceType } from 'services/auth_policy';
import { releaseCustomRoutes } from 'modules/governance/route';
import { releaseRateLimitRule } from 'modules/governance/ratelimit';
import { releaseLaneGroups } from 'modules/governance/lane_group';
import { releaseLosslessRule } from 'modules/governance/lossless';
import { releaseCircuitBreaker } from 'modules/governance/circuitbreaker';
import { releaseFaultDetect } from 'modules/governance/faultdetect';

const { StepItem } = Steps;
const { FormItem } = Form;

export interface IPublishFormProps {
    ruleId: string;
    ruleName: string;
    resource: PolicySourceType;
    visible: boolean;
    close: () => void;
}

const PublishForm: React.FC<IPublishFormProps> = (props) => {
    const [form] = Form.useForm();
    const dispatch = useAppDispatch();

    const onSubmit: FormProps['onSubmit'] = async (e) => {
        console.log(e);
        if (e.validateResult !== true) {
            return;
        }
        const releaseType = form.getFieldValue("releaseType") as string;
        // 提交发布
        const pubData: RuleRelease = {
            rule_id: props.ruleId,
            rule_name: props.ruleName,
            release_name: form.getFieldValue("name") as string,
            description: form.getFieldValue("comment") as string,
            release_type: releaseType === 'normal' ? 'normal' as const : 'gray' as const,
            client_label: form.getFieldValue("betaLabels") as MatcheLabel[],
            resource: props.resource,
        }

        let ret;
        switch (props.resource) {
            case PolicySourceType.RouteRules:
                ret = await dispatch(releaseCustomRoutes({ param: [pubData] }));
                break;
            case PolicySourceType.RateLimitRules:
                ret = await dispatch(releaseRateLimitRule({ param: [pubData] }));
                break;
            case PolicySourceType.LaneRules:
                ret = await dispatch(releaseLaneGroups({ param: [pubData] }));
                break;
            case PolicySourceType.CircuitBreakerRules:
                ret = await dispatch(releaseCircuitBreaker({ param: [pubData] }));
                break;
            case PolicySourceType.FaultDetectRules:
                ret = await dispatch(releaseFaultDetect({ param: [pubData] }));
                break;
            case PolicySourceType.LossLessRules:
                ret = await dispatch(releaseLosslessRule({ param: [pubData] }));
                break;
            default:
                openErrNotification('获取规则发布版本记录失败', '不支持的规则类型');
        }

        if (ret?.meta.requestStatus === 'fulfilled') {
            openInfoNotification('请求成功', `发布${props.ruleName}规则成功`);
            props.close();
        } else {
            openErrNotification('请求失败', `发布${props.ruleName}规则失败: ${ret?.payload as string}`);
        }
    }

    return (
        <>
            <Drawer
                header={'规则发布'}
                size='960px'
                style={{ width: '100%' }}
                visible={props.visible}
                placement="right"
                onClose={() => {
                    props.close();
                }}
                footer={
                    <Space>
                        <Button theme='primary' onClick={() => form.submit()} style={{ marginTop: 20 }}>
                            发布
                        </Button>
                        <Button
                            theme='default'
                            style={{ marginLeft: 10, marginTop: 20 }}
                            onClick={() => {
                                props.close();
                            }}
                        >
                            取消
                        </Button>
                    </Space>
                }
            >
                <Form
                    form={form}
                    labelWidth={100}
                    style={{ marginTop: 20 }}
                    onSubmit={onSubmit}
                >
                    <FormItem label="规则ID" name="ruleId" initialData={props.ruleId}>
                        <Input readonly />
                    </FormItem>
                    <FormItem label="规则名称" name="ruleName" initialData={props.ruleName}>
                        <Input readonly />
                    </FormItem>
                    <FormItem
                        label="版本名称"
                        name="name"
                        rules={[
                            { required: true, message: '版本名称不能为空' },
                            { max: 64, message: '长度不超过64个字符' },
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
                </Form>
            </Drawer>
        </>
    )
}

export default React.memo(PublishForm);
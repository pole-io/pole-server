import React, { useState, useRef } from 'react';
import classnames from 'classnames';
import { Form, MessagePlugin, Input, Checkbox, Button, FormInstanceFunctions, SubmitContext } from 'tdesign-react';
import { LockOnIcon, UserIcon, MailIcon, BrowseOffIcon, BrowseIcon } from 'tdesign-icons-react';
import useCountdown from '../components/hooks/useCountDown';

import Style from './index.module.less';

const { FormItem } = Form;

export type ERegisterType = 'phone' | 'email';

export default function Register() {
  const [registerType, changeRegisterType] = useState('phone');
  const [showPsw, toggleShowPsw] = useState(false);
  const { countdown, setupCountdown } = useCountdown(60);
  const formRef = useRef<FormInstanceFunctions>();

  const onSubmit = (e: SubmitContext) => {
    if (e.validateResult === true) {
      const { checked } = formRef.current?.getFieldsValue?.(['checked']) as { checked: boolean };
      if (!checked) {
        MessagePlugin.error('请同意 TDesign 服务协议和 TDesign 隐私声明');
        return;
      }
      MessagePlugin.success('注册成功');
    }
  };

  const switchType = (val: ERegisterType) => {
    formRef.current?.reset?.();
    changeRegisterType(val);
  };

  return (
    <div>
      <Form
        ref={formRef}
        className={classnames(Style.itemContainer, `register-${registerType}`)}
        labelWidth={0}
        onSubmit={onSubmit}
      >
        <FormItem name='name' rules={[{ required: true, message: '账户名必填', type: 'error' }]}>
          <Input maxlength={11} size='large' placeholder='请输入账户名' prefixIcon={<UserIcon />} />
        </FormItem>

        <FormItem name='password' rules={[{ required: true, message: '密码必填', type: 'error' }]}>
          <Input
            size='large'
            type={showPsw ? 'text' : 'password'}
            clearable
            placeholder='请输入登录密码'
            prefixIcon={<LockOnIcon />}
            suffixIcon={
              showPsw ? (
                <BrowseIcon onClick={() => toggleShowPsw((current) => !current)} />
              ) : (
                <BrowseOffIcon onClick={() => toggleShowPsw((current) => !current)} />
              )
            }
          />
        </FormItem>
        <FormItem>
          <Button block size='large' type='submit'>
            注册
          </Button>
        </FormItem>
        <div className={Style.switchContainer}>
          <span className={Style.switchTip} onClick={() => switchType(registerType === 'phone' ? 'email' : 'phone')}>
            {registerType === 'phone' ? '使用邮箱注册' : '使用手机号注册'}
          </span>
        </div>
      </Form>
    </div>
  );
}

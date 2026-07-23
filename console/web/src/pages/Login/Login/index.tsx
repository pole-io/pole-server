import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Form, MessagePlugin, Input, Button, SubmitContext } from 'components/Fluent';
import { LockOnIcon, UserIcon, BrowseOffIcon, BrowseIcon } from 'components/Fluent/icons';
import classnames from 'classnames';
import { useAppDispatch } from 'modules/store';
import { login } from 'modules/user/login';

import Style from './index.module.less';
import { openErrNotification } from 'utils/notifition';

const { FormItem } = Form;

export type ELoginType = 'password' | 'phone' | 'qrcode';

export default function Login() {
  const [loginType] = useState<ELoginType>('password');
  const [showPsw, toggleShowPsw] = useState(false);
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const navigate = useNavigate();
  const dispatch = useAppDispatch();

  const onSubmit = async (e: SubmitContext) => {
    if (e.validateResult !== true) {
      return;
    }
    setSubmitting(true);
    try {
      const result = await dispatch(login({ username, password }));
      if (result.meta.requestStatus !== 'fulfilled') {
        openErrNotification('请求错误', result?.payload as string);
      } else {
        MessagePlugin.success('登录成功');
        navigate('/namespace');
      }
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div>
      <Form
        className={classnames(Style.itemContainer, `login-${loginType}`)}
        labelWidth={0}
        onSubmit={onSubmit}
      >
        {loginType === 'password' && (
          <>
            <FormItem name='account' rules={[{ required: true, message: '账号必填', type: 'error' }]}>
              <Input
                size='large'
                aria-label='账号'
                aria-required='true'
                autoComplete='username'
                placeholder='请输入账号'
                prefixIcon={<UserIcon />}
                value={username}
                onChange={(value) => setUsername(value)}></Input>
            </FormItem>
            <FormItem name='password' rules={[{ required: true, message: '密码必填', type: 'error' }]}>
              <Input
                size='large'
                aria-label='登录密码'
                aria-required='true'
                autoComplete='current-password'
                type={showPsw ? 'text' : 'password'}
                clearable
                placeholder='请输入登录密码'
                value={password}
                onChange={(value) => setPassword(value)}
                prefixIcon={<LockOnIcon />}
                suffixIcon={
                  <Button
                    type='button'
                    className={Style.passwordToggle}
                    variant='text'
                    shape='square'
                    size='small'
                    aria-label={showPsw ? '隐藏密码' : '显示密码'}
                    onClick={() => toggleShowPsw((current) => !current)}
                  >
                    {showPsw ? <BrowseIcon /> : <BrowseOffIcon />}
                  </Button>
                }
              />
            </FormItem>
          </>
        )}

        <FormItem className={Style.btnContainer}>
          <Button block size='large' theme='primary' type='submit' loading={submitting}>
            登录
          </Button>
        </FormItem>
      </Form>
    </div>
  );
}

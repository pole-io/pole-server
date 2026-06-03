import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Form, MessagePlugin, Input, Button, SubmitContext } from 'tdesign-react';
import { LockOnIcon, UserIcon, BrowseOffIcon, BrowseIcon } from 'tdesign-icons-react';
import classnames from 'classnames';
import { useAppDispatch, useAppSelector } from 'modules/store';
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
  const navigate = useNavigate();
  const dispatch = useAppDispatch();

  const onSubmit = async (e: SubmitContext) => {
    if (e.validateResult !== true) {
      return
    }
    const result = await dispatch(login({ username, password }));
    console.log(result);
    if (result.meta.requestStatus !== 'fulfilled') {
      openErrNotification('请求错误', result?.payload as string);
    } else {
      MessagePlugin.success('登录成功');
      navigate('/namespace');
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
                placeholder='请输入账号'
                prefixIcon={<UserIcon />}
                value={username}
                onChange={(value) => setUsername(value)}></Input>
            </FormItem>
            <FormItem name='password' rules={[{ required: true, message: '密码必填', type: 'error' }]}>
              <Input
                size='large'
                type={showPsw ? 'text' : 'password'}
                clearable
                placeholder='请输入登录密码'
                value={password}
                onChange={(value) => setPassword(value)}
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
          </>
        )}

        <FormItem className={Style.btnContainer}>
          <Button block size='large' type='submit'>
            登录
          </Button>
        </FormItem>
      </Form>
    </div>
  );
}

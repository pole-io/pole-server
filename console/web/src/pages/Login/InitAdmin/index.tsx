import React, { useState, useEffect, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import classNames from 'classnames';
import { Form, MessagePlugin, Input, Button, SubmitContext, Loading } from 'tdesign-react';
import { LockOnIcon, UserIcon, BrowseOffIcon, BrowseIcon } from 'tdesign-icons-react';
import { useAppSelector } from 'modules/store';
import { selectGlobal } from 'modules/global';
import { checkExistAdminUser, initAdminUser } from 'services/login';
import { openErrNotification } from 'utils/notifition';
import LoginHeader from '../components/Header';

import Style from '../index.module.less';
import FormStyle from './index.module.less';

const { FormItem } = Form;

export default function InitAdmin() {
  const [showPsw, toggleShowPsw] = useState(false);
  const [loading, setLoading] = useState(false);
  const [checking, setChecking] = useState(true);
  const navigate = useNavigate();
  const globalState = useAppSelector(selectGlobal);
  const { theme } = globalState;
  const cancelledRef = useRef(false);

  useEffect(() => {
    cancelledRef.current = false;
    const ac = new AbortController();

    checkExistAdminUser({ signal: ac.signal })
      .then((res) => {
        if (cancelledRef.current) return;
        if (res && typeof res === 'object' && 'id' in res) {
          navigate('/login', { replace: true });
        }
      })
      .catch(() => {})
      .finally(() => {
        if (!cancelledRef.current) {
          setChecking(false);
        }
      });

    return () => {
      cancelledRef.current = true;
      ac.abort();
    };
  }, [navigate]);

  const onSubmit = async (e: SubmitContext) => {
    if (e.validateResult !== true) return;
    const fields = e.fields as { name?: string; password?: string; confirmPassword?: string };
    const name = fields?.name ?? '';
    const password = fields?.password ?? '';
    const confirmPassword = fields?.confirmPassword ?? '';
    if (password !== confirmPassword) {
      openErrNotification('校验失败', '两次输入的密码不一致');
      return;
    }
    setLoading(true);
    try {
      await initAdminUser({ name, password });
      MessagePlugin.success('管理员账户创建成功，请使用该账户登录');
      navigate('/login', { replace: true });
    } catch (err) {
      openErrNotification('创建失败', (err as Error)?.message ?? '未知错误');
    } finally {
      setLoading(false);
    }
  };

  if (checking) {
    return (
      <div
        className={classNames(Style.loginWrapper, {
          [Style.light]: theme === 'light',
          [Style.dark]: theme !== 'light',
        })}
      >
        <LoginHeader />
        <div className={Style.loginContainer} style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', minHeight: 200 }}>
          <Loading text="检查中..." />
        </div>
      </div>
    );
  }

  return (
    <div
      className={classNames(Style.loginWrapper, {
        [Style.light]: theme === 'light',
        [Style.dark]: theme !== 'light',
      })}
    >
      <LoginHeader />
      <div className={Style.loginContainer}>
        <div className={Style.titleContainer}>
          <h1 className={Style.title}>初始化管理员账户</h1>
          <div className={Style.subTitle}>
            <p className={classNames(Style.tip, Style.registerTip)}>检测到当前无管理员账户，请创建首个管理员账号</p>
          </div>
        </div>
        <Form
          className={classNames(FormStyle.itemContainer, 'init-admin')}
          labelWidth={0}
          onSubmit={onSubmit}
        >
          <FormItem name="name" rules={[{ required: true, message: '管理员账号必填', type: 'error' }]}>
            <Input size="large" placeholder="请输入管理员账号" prefixIcon={<UserIcon />} />
          </FormItem>
          <FormItem name="password" rules={[{ required: true, message: '密码必填', type: 'error' }]}>
            <Input
              size="large"
              type={showPsw ? 'text' : 'password'}
              clearable
              placeholder="请输入登录密码"
              prefixIcon={<LockOnIcon />}
              suffixIcon={
                showPsw ? (
                  <BrowseIcon onClick={() => toggleShowPsw(false)} />
                ) : (
                  <BrowseOffIcon onClick={() => toggleShowPsw(true)} />
                )
              }
            />
          </FormItem>
          <FormItem
            name="confirmPassword"
            rules={[{ required: true, message: '请再次输入密码', type: 'error' }]}
          >
            <Input
              size="large"
              type="password"
              placeholder="请再次输入密码"
              prefixIcon={<LockOnIcon />}
            />
          </FormItem>
          <FormItem className={FormStyle.btnContainer}>
            <Button block size="large" type="submit" loading={loading}>
              创建并前往登录
            </Button>
          </FormItem>
        </Form>
      </div>
    </div>
  );
}

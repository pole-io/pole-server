import React, { useState, useEffect, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import classNames from 'classnames';
import { Form, MessagePlugin, Input, Button, SubmitContext, Loading } from 'components/Fluent';
import { LockOnIcon, UserIcon, BrowseOffIcon, BrowseIcon } from 'components/Fluent/icons';
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
      .catch(() => {
        if (!cancelledRef.current) {
          MessagePlugin.error('无法检查管理员账户，请确认服务可用后重试');
          navigate('/login', { replace: true });
        }
      })
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
        <main className={Style.loadingState}>
          <Loading text="检查中..." />
        </main>
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
      <main className={Style.loginMain}>
        <section className={Style.introPanel} aria-labelledby="init-admin-product-title">
          <div className={Style.introContent}>
            <p className={Style.eyebrow}>
              <span aria-hidden="true" />
              POLE CONTROL PLANE
            </p>
            <h1 id="init-admin-product-title">开始使用服务治理控制台</h1>
            <p className={Style.introDescription}>创建首个管理员后，即可配置服务、治理规则与 Agent 工作台。</p>
          </div>
        </section>

        <section className={Style.authSection} aria-labelledby="init-admin-form-title">
          <div className={Style.authCard}>
            <div className={Style.authHeading}>
              <p>首次设置</p>
              <h2 id="init-admin-form-title">初始化管理员账户</h2>
              <span>此账户将拥有控制台的管理权限。</span>
            </div>
            <Form
              className={classNames(FormStyle.itemContainer, 'init-admin')}
              labelWidth={0}
              onSubmit={onSubmit}
            >
              <FormItem name="name" rules={[{ required: true, message: '管理员账号必填', type: 'error' }]}>
                <Input
                  size="large"
                  aria-label="管理员账号"
                  aria-required="true"
                  autoComplete="username"
                  placeholder="请输入管理员账号"
                  prefixIcon={<UserIcon />}
                />
              </FormItem>
              <FormItem name="password" rules={[{ required: true, message: '密码必填', type: 'error' }]}>
                <Input
                  size="large"
                  aria-label="管理员密码"
                  aria-required="true"
                  autoComplete="new-password"
                  type={showPsw ? 'text' : 'password'}
                  clearable
                  placeholder="请输入登录密码"
                  prefixIcon={<LockOnIcon />}
                  suffixIcon={
                    <Button
                      type="button"
                      className={FormStyle.passwordToggle}
                      variant="text"
                      shape="square"
                      size="small"
                      aria-label={showPsw ? '隐藏密码' : '显示密码'}
                      onClick={() => toggleShowPsw((current) => !current)}
                    >
                      {showPsw ? <BrowseIcon /> : <BrowseOffIcon />}
                    </Button>
                  }
                />
              </FormItem>
              <FormItem name="confirmPassword" rules={[{ required: true, message: '请再次输入密码', type: 'error' }]}>
                <Input
                  size="large"
                  aria-label="确认管理员密码"
                  aria-required="true"
                  autoComplete="new-password"
                  type="password"
                  placeholder="请再次输入密码"
                  prefixIcon={<LockOnIcon />}
                />
              </FormItem>
              <FormItem className={FormStyle.btnContainer}>
                <Button block size="large" theme="primary" type="submit" loading={loading}>
                  创建并前往登录
                </Button>
              </FormItem>
            </Form>
          </div>
          <p className={Style.authFootnote}>已有管理员时，系统会自动返回登录页。</p>
        </section>
      </main>
    </div>
  );
}

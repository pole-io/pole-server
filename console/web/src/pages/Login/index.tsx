import React, { memo, useState, useEffect, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import classNames from 'classnames';
import { Loading, MessagePlugin } from 'components/Fluent';
import Login from './Login';
import LoginHeader from './components/Header';
import { useAppSelector } from 'modules/store';
import { selectGlobal } from 'modules/global';
import { checkExistAdminUser } from 'services/login';

import Style from './index.module.less';

export default memo(() => {
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
        if (!res || typeof res !== 'object' || !('id' in res)) {
          navigate('/init-admin', { replace: true });
          return;
        }
      })
      .catch(() => {
        if (!cancelledRef.current) {
          MessagePlugin.error('无法检查管理员账户，请稍后重试');
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

  if (checking) {
    return (
      <div className={classNames(Style.loginWrapper, { [Style.light]: theme === 'light', [Style.dark]: theme !== 'light' })}>
        <LoginHeader />
        <main className={Style.loadingState}>
          <Loading text="检查管理员账户..." />
        </main>
      </div>
    );
  }

  return (
    <div
      className={classNames(Style.loginWrapper, { [Style.light]: theme === 'light', [Style.dark]: theme !== 'light' })}
    >
      <LoginHeader />
      <main className={Style.loginMain}>
        <section className={Style.introPanel} aria-labelledby="login-product-title">
          <div className={Style.introContent}>
            <p className={Style.eyebrow}>
              <span aria-hidden="true" />
              POLE CONTROL PLANE
            </p>
            <h1 id="login-product-title">服务治理控制台</h1>
            <p className={Style.introDescription}>
              统一管理服务发现、配置发布与治理策略，并通过 Agent 协助完成资源操作。
            </p>
          </div>
        </section>

        <section className={Style.authSection} aria-labelledby="login-form-title">
          <div className={Style.authCard}>
            <div className={Style.authHeading}>
              <p>欢迎回来</p>
              <h2 id="login-form-title">登录控制台</h2>
              <span>使用管理员或已创建的账号继续。</span>
            </div>
            <Login />
          </div>
          <p className={Style.authFootnote}>账号由管理员在控制台的认证管理中创建。</p>
        </section>
      </main>
    </div>
  );
});

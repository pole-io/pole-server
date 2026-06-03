import React, { memo, useState, useEffect, useRef } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import classNames from 'classnames';
import { Loading, MessagePlugin } from 'tdesign-react';
import Login from './Login';
import Register from './Register';
import LoginHeader from './components/Header';
import { useAppSelector } from 'modules/store';
import { selectGlobal } from 'modules/global';
import { checkExistAdminUser } from 'services/login';

import Style from './index.module.less';

export default memo(() => {
  const [type, setType] = useState('login');
  const [checking, setChecking] = useState(true);
  const navigate = useNavigate();
  const location = useLocation();
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
        if ((location.state as { from?: unknown })?.from) {
          MessagePlugin.warning('您当前未登录，请先登录');
        }
      })
      .catch(() => {
        if (!cancelledRef.current) {
          navigate('/init-admin', { replace: true });
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
  }, [navigate, location.state]);

  const handleSwitchLoginType = () => {
    setType(type === 'register' ? 'login' : 'register');
  };

  if (checking) {
    return (
      <div className={classNames(Style.loginWrapper, { [Style.light]: theme === 'light', [Style.dark]: theme !== 'light' })}>
        <LoginHeader />
        <div className={Style.loginContainer} style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', minHeight: 200 }}>
          <Loading text="检查管理员账户..." />
        </div>
      </div>
    );
  }

  return (
    <div
      className={classNames(Style.loginWrapper, { [Style.light]: theme === 'light', [Style.dark]: theme !== 'light' })}
    >
      <LoginHeader />
      <div className={Style.loginContainer}>
        <div className={Style.titleContainer}>
          <h1 className={Style.title}>AI Native 的服务治理平台</h1>
          <div className={Style.subTitle}>
            <p className={classNames(Style.tip, Style.registerTip)}>
              {type === 'register' ? '已有账号?' : '默认帐户: pole/pole123, 如果没有账号'}
            </p>
            <p className={classNames(Style.tip, Style.loginTip)} onClick={handleSwitchLoginType}>
              {type === 'register' ? '登录' : '注册新账号'}
            </p>
          </div>
        </div>
        {type === 'login' ? <Login /> : <Register />}
      </div>
    </div>
  );
});

import React from 'react';
import { FluentProvider } from '@fluentui/react-components';
import { useAppSelector } from 'modules/store';
import { selectGlobal } from 'modules/global';
import { ETheme } from 'types/index.d';
import { latticeDarkTheme, latticeLightTheme } from './theme';
import { FluentToastHost } from './toast';

interface FluentAppProviderProps {
  children: React.ReactNode;
}

const FluentAppProvider: React.FC<FluentAppProviderProps> = ({ children }) => {
  const { theme } = useAppSelector(selectGlobal);
  const baseTheme = theme === ETheme.dark ? latticeDarkTheme : latticeLightTheme;
  const fluentTheme = {
    ...baseTheme,
    colorBrandBackground: 'var(--app-brand)',
    colorBrandBackgroundHover: 'var(--app-brand-7)',
    colorBrandBackgroundPressed: 'var(--app-brand-9)',
    colorBrandForeground1: 'var(--app-brand)',
    colorBrandForeground2: 'var(--app-brand-7)',
    colorCompoundBrandForeground1: 'var(--app-brand)',
    colorCompoundBrandForeground1Hover: 'var(--app-brand-7)',
    colorCompoundBrandForeground1Pressed: 'var(--app-brand-9)',
    colorCompoundBrandStroke: 'var(--app-brand)',
    colorCompoundBrandStrokeHover: 'var(--app-brand-7)',
    colorCompoundBrandStrokePressed: 'var(--app-brand-9)',
  };
  return (
    <div className="fluent-provider-shell">
      <FluentProvider theme={fluentTheme}>
        <div className="fluent-app-provider">
          {children}
          <FluentToastHost />
        </div>
      </FluentProvider>
    </div>
  );
};

export default FluentAppProvider;

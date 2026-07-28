import React from 'react';
import { Button, Popconfirm, Tooltip } from 'components/Fluent';
import style from './index.module.less';
import {
  BrowseIcon,
  CopyIcon,
  DeleteIcon,
  Edit1Icon,
  ListIcon,
  LockOnIcon,
  RefreshIcon,
  RollbackIcon,
  SendIcon,
  UserVisibleIcon,
} from 'components/Fluent/icons';

export type OperationAction =
  | 'viewEdit'
  | 'view'
  | 'edit'
  | 'authorize'
  | 'delete'
  | 'publish'
  | 'rollback'
  | 'copy'
  | 'tools'
  | 'token'
  | 'refresh';

const operationLabelMap: Record<OperationAction, string> = {
  viewEdit: '查看',
  view: '查看',
  edit: '编辑',
  authorize: '授权',
  delete: '删除',
  publish: '发布',
  rollback: '回滚',
  copy: '复制',
  tools: '查看工具',
  token: '查看 Token',
  refresh: '刷新',
};

const operationIconMap: Record<OperationAction, React.ReactNode> = {
  viewEdit: <BrowseIcon />,
  view: <BrowseIcon />,
  edit: <Edit1Icon />,
  authorize: <LockOnIcon />,
  delete: <DeleteIcon />,
  publish: <SendIcon />,
  rollback: <RollbackIcon />,
  copy: <CopyIcon />,
  tools: <ListIcon />,
  token: <UserVisibleIcon />,
  refresh: <RefreshIcon />,
};

type ButtonShape = React.ComponentProps<typeof Button>['shape'];
type ButtonVariant = React.ComponentProps<typeof Button>['variant'];
type ButtonTheme = React.ComponentProps<typeof Button>['theme'];
type ButtonSize = React.ComponentProps<typeof Button>['size'];
type TooltipPlacement = React.ComponentProps<typeof Tooltip>['placement'];

const joinClassNames = (...classNames: Array<string | undefined>) => classNames.filter(Boolean).join(' ');

export interface OperationButtonGroupProps {
  children: React.ReactNode;
  className?: string;
}

export const OperationButtonGroup: React.FC<OperationButtonGroupProps> = ({ children, className }) => (
  <div className={joinClassNames(style.operationGroup, className)}>{children}</div>
);

export interface OperationButtonProps {
  action: OperationAction;
  label?: string;
  disabledLabel?: string;
  disabled?: boolean;
  onClick?: () => void;
  shape?: ButtonShape;
  variant?: ButtonVariant;
  theme?: ButtonTheme;
  size?: ButtonSize;
  placement?: TooltipPlacement;
  className?: string;
}

const getLabel = (action: OperationAction, label?: string, disabled?: boolean, disabledLabel?: string) => {
  if (disabled && disabledLabel) {
    return disabledLabel;
  }
  return label || operationLabelMap[action];
};

export const OperationButton: React.FC<OperationButtonProps> = ({
  action,
  label,
  disabledLabel,
  disabled,
  onClick,
  shape = 'square',
  variant = 'text',
  theme,
  size = 'small',
  placement = 'top',
  className,
}) => {
  const resolvedLabel = getLabel(action, label, disabled, disabledLabel);
  return (
    <Tooltip content={resolvedLabel} placement={placement}>
      <span className={style.operationItem}>
        <Button
          aria-label={resolvedLabel}
          title={resolvedLabel}
          shape={shape}
          variant={variant}
          theme={theme}
          size={size}
          className={joinClassNames(style.operationButton, className)}
          disabled={disabled}
          onClick={onClick}
        >
          {operationIconMap[action]}
        </Button>
      </span>
    </Tooltip>
  );
};

export interface ConfirmOperationButtonProps extends Omit<OperationButtonProps, 'onClick'> {
  confirmContent: React.ReactNode;
  onConfirm: () => void;
}

export const ConfirmOperationButton: React.FC<ConfirmOperationButtonProps> = ({
  action,
  label,
  disabledLabel,
  disabled,
  confirmContent,
  onConfirm,
  shape = 'square',
  variant = 'text',
  theme,
  size = 'small',
  placement = 'top',
  className,
}) => {
  const resolvedLabel = getLabel(action, label, disabled, disabledLabel);
  return (
    <Tooltip content={resolvedLabel} placement={placement}>
      <span className={style.operationItem}>
        <Popconfirm
          content={confirmContent}
          destroyOnClose
          placement="top"
          showArrow
          theme="default"
          onConfirm={disabled ? undefined : onConfirm}
        >
          <Button
            aria-label={resolvedLabel}
            title={resolvedLabel}
            shape={shape}
            variant={variant}
            theme={theme}
            size={size}
            className={joinClassNames(style.operationButton, className)}
            disabled={disabled}
          >
            {operationIconMap[action]}
          </Button>
        </Popconfirm>
      </span>
    </Tooltip>
  );
};

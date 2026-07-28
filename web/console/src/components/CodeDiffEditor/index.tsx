import React from 'react';
import { DiffEditor } from '@monaco-editor/react'
import { LoadingIcon } from 'components/Fluent/icons';
import { useAppSelector } from 'modules/store';
import { selectGlobal } from 'modules/global';
import Style from './index.module.less';
import classNames from 'classnames';

export interface ICodeDiffEditorProps {
    namespace: string;
    group: string;
    filename: string;
    readonly?: boolean;
    allowFullScreen?: boolean;
    curValue?: string;
    nextValue?: string;
    language?: string;
    theme?: string;
    height?: string | number;
    onChange?: (value: string | undefined, event: any) => void;
    onMount?: (editor: any, monaco: any) => void;
}

export function toHighlightLanguage(format?: string) {
    if (!format) {
        return "text"
    }
    if (format === "properties") {
        return 'ini'
    }
    if (format === 'yml') {
        return 'yaml'
    }
    return format
}

const CodeDiffEditor: React.FC<ICodeDiffEditorProps> = props => {
    const { theme: appTheme } = useAppSelector(selectGlobal);
    const modelPathPrefix = React.useMemo(() => {
        const fileKey = `${props.namespace}/${props.group}/${props.filename}` || 'config-file';
        return `inmemory://config-diff/${encodeURIComponent(fileKey)}`;
    }, [props.namespace, props.group, props.filename]);

    return (
        <section
            className={classNames({
                [Style.monacoSection]: true,
            })}
            style={{
                height: props.height || 'clamp(280px, 48dvh, 560px)',
                minHeight: 240,
                width: "100%",
                position: 'relative'
            }}
        >
            <div style={{ display: 'flex', justifyContent: 'space-between', padding: '0 8px 4px 8px', fontSize: 12, color: 'var(--app-text-secondary)' }}>
                <span>{'当前版本'}</span>
                <span>{'目标版本'}</span>
            </div>
            <DiffEditor
                loading={<LoadingIcon />}
                theme={props.theme || (appTheme === 'dark' ? 'vs-dark' : 'vs')}
                options={{
                    // 控制是否只读
                    readOnly: true,
                    // 控制是否显示行号
                    lineNumbers: 'on',
                    // 控制布局自适应
                    automaticLayout: true,
                    //
                }}
                original={props.curValue}
                modified={props.nextValue}
                originalModelPath={`${modelPathPrefix}/original.${toHighlightLanguage(props.language)}`}
                modifiedModelPath={`${modelPathPrefix}/modified.${toHighlightLanguage(props.language)}`}
                keepCurrentOriginalModel={true}
                keepCurrentModifiedModel={true}
                originalLanguage={toHighlightLanguage(props.language)}
                modifiedLanguage={toHighlightLanguage(props.language)}
            />
        </section>
    )
}

// 使用React.memo优化组件渲染
export default React.memo(CodeDiffEditor);

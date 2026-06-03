import React from 'react';
import { Editor as MonacoEditor } from '@monaco-editor/react'
import type { OnChange, OnMount } from '@monaco-editor/react'
import { Icon, LoadingIcon } from 'tdesign-icons-react';
import { editor } from 'monaco-editor';
import Style from './index.module.less';
import classNames from 'classnames';

export interface ICodeEditorProps {
    readonly?: boolean;
    allowFullScreen?: boolean;
    value?: string;
    language?: string;
    theme?: string;
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

const CodeEditor: React.FC<ICodeEditorProps> = props => {
    // 默认不开启全屏
    const [isFullScreen, setFullScreen] = React.useState(false)
    const editorRef = React.useRef<editor.IStandaloneCodeEditor>(null!);
    const [rect, setRect] = React.useState({ width: 0, height: 0 })

    const handleEditorChange: OnChange = (value, event) => {
        if (props.onChange) {
            props.onChange(value, event);
        }
    }

    const handleEditorDidMount: OnMount = (editor: editor.IStandaloneCodeEditor, monaco) => {
        editorRef.current = editor;
    }

    const handleFullScreen = () => {
        setFullScreen(!isFullScreen)
        if (editorRef.current) {
            if (isFullScreen) {
                editorRef.current.layout({
                    height: rect.height,
                    width: rect.width,
                })
            } else {
                const originRect: DOMRect = editorRef.current.getContainerDomNode().getBoundingClientRect();
                if (originRect) {
                    setRect({ width: originRect.width, height: originRect.height });
                    setFullScreen(true);
                    setTimeout(() => {
                        editorRef.current?.layout({
                            height: document.body.clientHeight,
                            width: document.body.clientWidth,
                        });
                    }, 0);
                }
            }
        }
    }

    return (
        <section
            className={classNames({
                [Style.monacoSection]: true,
                [Style.monacoSectionFull]: isFullScreen,
            })}
            style={
                isFullScreen
                    ? {
                        position: 'fixed',
                        top: 0,
                        left: 0,
                        right: 0,
                        bottom: 0,
                        width: '100vw',
                        height: '100vh',
                        zIndex: 2000,
                    }
                    : {
                        height: 'calc(100vh - 390px)',
                        width: "100%",
                        position: 'relative'
                    }
            }
        >
            {props.allowFullScreen && (
                <Icon
                    name={!isFullScreen ? 'fullscreen-1' : 'fullscreen-exit-1'}
                    size={'30'}
                    style={{
                        position: 'absolute',
                        right: 100,
                        top: 0,
                        zIndex: 999,
                    }}
                    onClick={handleFullScreen}
                />
            )}
            <MonacoEditor
                loading={<LoadingIcon />}
                theme={'vs'}
                language={toHighlightLanguage(props.language)}
                value={props.value}
                onMount={handleEditorDidMount}
                onChange={handleEditorChange}
                options={{
                    // 控制是否只读
                    readOnly: props.readonly,
                    // 控制是否显示行号
                    lineNumbers: 'on',
                    //
                    automaticLayout: true,
                }}
            />
        </section>
    )
}

export default React.memo(CodeEditor);
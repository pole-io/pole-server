export const VisibilityMode_Single: string = 'single'
export const VisibilityMode_All: string = 'all'
export const VisibilityMode_Specified: string = 'specified'

export const VisibilityModeMap = {
    [VisibilityMode_Single]: '仅当前命名空间可见',
    [VisibilityMode_All]: '全部命名空间可见（包括新增）',
    [VisibilityMode_Specified]: '指定命名空间',
}

export const CheckVisibilityMode = (exportTo: string[] = [], namespace: string) => {
    if (!exportTo?.length) {
        return VisibilityMode_Single
    }
    return exportTo?.includes('*')
        ? VisibilityMode_All
        : (exportTo.length === 1 && exportTo?.[0] === namespace) || exportTo.length === 0
            ? VisibilityMode_Single
            : VisibilityMode_Specified
}
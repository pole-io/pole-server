import { apiRequest, getApiRequest, putApiRequest, RequestError } from 'utils/request';
import { BaseURL, MatcheLabel } from './types';
import type { ConfigFile } from './config_files';

export const TemplateEngine = {
  name: 'pole-mustache',
  version: 'v1',
} as const;

export type TemplateParameterType =
  | 'TEMPLATE_PARAMETER_STRING'
  | 'TEMPLATE_PARAMETER_BOOLEAN'
  | 'TEMPLATE_PARAMETER_INTEGER'
  | 'TEMPLATE_PARAMETER_DECIMAL';

export interface ConfigTemplateParameterSchema {
  name: string;
  type: TemplateParameterType;
  required: boolean;
  defaultValue?: ConfigTemplateValue;
  sensitive?: boolean;
  description?: string;
}

export type ConfigTemplateValue =
  | { stringValue: string }
  | { booleanValue: boolean }
  | { integerValue: number | string }
  | { decimalValue: string };

export interface ConfigTemplateBinding {
  templateId: string | number;
  templateReleaseId: string;
  bindingReleaseId?: string;
}

export interface ConfigFileTemplate {
  id: string | number;
  name: string;
  content: string;
  comment?: string;
  format: string;
  engine: typeof TemplateEngine;
  parameterSchema: ConfigTemplateParameterSchema[];
  labels?: Record<string, string>;
  revision?: string;
  ctime?: string;
  mtime?: string;
	 draftVersion?: number;
	 initializedFrom?: string;
}

export async function describeConfigTemplateLabels(templateId: string | number) {
  const res = await getApiRequest<{
    labels?: Record<string, string>;
    value?: { labels?: Record<string, string> };
  }>({
    action: `${BaseURL.CONFIG_TEMPLATE}/labels`,
    data: { template_id: templateId },
  });
  return res.labels ?? res.value?.labels ?? {};
}

export async function saveConfigTemplateLabels(templateId: string | number, labels: Record<string, string>) {
  return putApiRequest({
    action: `${BaseURL.CONFIG_TEMPLATE}/labels`,
    data: { template_id: Number(templateId), labels },
  });
}

export interface ConfigTemplateRelease extends ConfigFileTemplate {
  id: string;
  templateId: string | number;
  version: string | number;
  contentSha256?: string;
  createBy?: string;
}

export interface NamespaceTemplateValues {
  id?: string;
  namespace: string;
  templateId: string | number;
  values: Record<string, ConfigTemplateValue>;
  revision?: string;
  modifyBy?: string;
}

export type TemplateValueReleaseType = 'TEMPLATE_VALUE_RELEASE_NORMAL' | 'TEMPLATE_VALUE_RELEASE_GRAY';

export interface NamespaceTemplateValueRelease extends NamespaceTemplateValues {
  id: string;
  valuesId?: string;
  templateReleaseId: string;
  releaseType: TemplateValueReleaseType;
  betaLabels?: MatcheLabel[];
  priority?: number;
  active?: boolean;
  version?: string | number;
  comment?: string;
  createBy?: string;
}

export interface RenderDiagnostic {
  severity: 'DIAGNOSTIC_INFO' | 'DIAGNOSTIC_WARNING' | 'DIAGNOSTIC_ERROR';
  code: string;
  message: string;
  parameter?: string;
}

export interface RenderPreview {
  code: number;
  info: string;
  valid: boolean;
  renderedContent: string;
  format: string;
  renderedSha256: string;
  templateReleaseId?: string;
  valueReleaseId?: string;
  engine?: typeof TemplateEngine;
  diagnostics: RenderDiagnostic[];
}

type ApiTemplateParameterSchema = Omit<ConfigTemplateParameterSchema, 'defaultValue'> & {
  default_value?: ApiTemplateValue;
  defaultValue?: ApiTemplateValue;
};
type ApiTemplateValue = {
  string_value?: string;
  stringValue?: string;
  boolean_value?: boolean;
  booleanValue?: boolean;
  integer_value?: number | string;
  integerValue?: number | string;
  decimal_value?: string;
  decimalValue?: string;
};
type ApiTemplate = Omit<ConfigFileTemplate, 'parameterSchema'> & {
  parameter_schema?: ApiTemplateParameterSchema[];
  parameterSchema?: ApiTemplateParameterSchema[];
};

const normalizeValue = (value?: ApiTemplateValue): ConfigTemplateValue => {
  if (!value) return { stringValue: '' };
  if (value.string_value !== undefined || value.stringValue !== undefined) {
    return { stringValue: value.string_value ?? value.stringValue ?? '' };
  }
  if (value.boolean_value !== undefined || value.booleanValue !== undefined) {
    return { booleanValue: value.boolean_value ?? value.booleanValue ?? false };
  }
  if (value.integer_value !== undefined || value.integerValue !== undefined) {
    return { integerValue: value.integer_value ?? value.integerValue ?? 0 };
  }
  return { decimalValue: value.decimal_value ?? value.decimalValue ?? '0' };
};

const toApiValue = (value: ConfigTemplateValue): ApiTemplateValue => {
  if ('stringValue' in value) return { string_value: value.stringValue };
  if ('booleanValue' in value) return { boolean_value: value.booleanValue };
  if ('integerValue' in value) return { integer_value: value.integerValue };
  return { decimal_value: value.decimalValue };
};

const normalizeSchema = (schema: ApiTemplateParameterSchema): ConfigTemplateParameterSchema => ({
  ...schema,
  description: schema.description ?? (schema as any).comment ?? '',
  defaultValue:
    schema.default_value || schema.defaultValue
      ? normalizeValue(schema.default_value ?? schema.defaultValue)
      : undefined,
});

const persistedSchemaType: Record<TemplateParameterType, number> = {
  TEMPLATE_PARAMETER_STRING: 1,
  TEMPLATE_PARAMETER_BOOLEAN: 2,
  TEMPLATE_PARAMETER_INTEGER: 3,
  TEMPLATE_PARAMETER_DECIMAL: 4,
};

const toPersistedValue = (value: ConfigTemplateValue) => {
  if ('stringValue' in value) return { type: 'string', value: value.stringValue };
  if ('booleanValue' in value) return { type: 'boolean', value: String(value.booleanValue) };
  if ('integerValue' in value) return { type: 'integer', value: String(value.integerValue) };
  return { type: 'decimal', value: value.decimalValue };
};

const toPersistedSchema = (schema: ConfigTemplateParameterSchema) => ({
  name: schema.name,
  type: persistedSchemaType[schema.type],
  required: schema.required,
  sensitive: Boolean(schema.sensitive),
  comment: schema.description || '',
  default: schema.defaultValue ? toPersistedValue(schema.defaultValue) : undefined,
});

const toApiSchema = (schema: ConfigTemplateParameterSchema) => ({
  name: schema.name,
  type: schema.type,
  required: schema.required,
  sensitive: Boolean(schema.sensitive),
  description: schema.description || '',
  default_value: schema.defaultValue ? toApiValue(schema.defaultValue) : undefined,
});

const normalizeTemplate = (template: ApiTemplate): ConfigFileTemplate => ({
  ...template,
  engine: template.engine || TemplateEngine,
  parameterSchema: (template.parameter_schema ?? template.parameterSchema ?? []).map(normalizeSchema),
});

const toApiTemplate = (template: ConfigFileTemplate) => ({
  id: template.id,
  name: template.name,
  content: template.content,
  comment: template.comment || '',
  format: template.format || 'text',
  engine: template.engine || TemplateEngine,
  parameter_schema: template.parameterSchema.map(toApiSchema),
  revision: template.revision || '',
});

export async function describeNamespaceConfigTemplateDraft(
  namespace: string,
  templateId: string | number,
): Promise<ConfigFileTemplate | undefined> {
  const response = await getApiRequest<any>({
    action: `${BaseURL.CONFIG_TEMPLATE}/environment-draft`,
    data: { namespace, template_id: templateId },
  });
  const raw = response?.value ?? response;
  if (!raw?.template_id && !raw?.templateId) return undefined;
  let schema: ApiTemplateParameterSchema[] = [];
  try {
    schema = JSON.parse(raw.parameter_schema ?? raw.parameterSchema ?? '[]');
  } catch {
    schema = [];
  }
  return normalizeTemplate({
    id: raw.template_id ?? raw.templateId,
    name: raw.name || '',
    comment: raw.comment || '',
    content: raw.content || '',
    format: raw.format || 'text',
    engine: {
      name: raw.engine || TemplateEngine.name,
      version: raw.engine_version ?? raw.engineVersion ?? TemplateEngine.version,
    },
    parameter_schema: schema,
    revision: raw.revision || '',
    draftVersion: Number(raw.draft_version ?? raw.draftVersion ?? 0),
    initializedFrom: raw.initialized_from ?? raw.initializedFrom ?? '',
  } as ApiTemplate);
}

export async function saveNamespaceConfigTemplateDraft(
  namespace: string,
  template: ConfigFileTemplate,
) {
  const response = await putApiRequest<any>({
    action: `${BaseURL.CONFIG_TEMPLATE}/environment-draft`,
    data: {
      namespace,
      templateID: Number(template.id),
      content: template.content,
      format: template.format || 'text',
      parameterSchema: JSON.stringify(template.parameterSchema.map(toPersistedSchema)),
      engine: template.engine?.name || TemplateEngine.name,
      engineVersion: template.engine?.version || TemplateEngine.version,
      revision: template.revision || '',
      draftVersion: template.draftVersion || 0,
      initializedFrom: template.initializedFrom || '',
    },
  });
  return response?.value ?? response;
}

const normalizeBinding = (binding: any): ConfigTemplateBinding => ({
  templateId: binding.template_id ?? binding.templateId,
  templateReleaseId: binding.template_release_id ?? binding.templateReleaseId ?? '',
  bindingReleaseId: binding.binding_release_id ?? binding.bindingReleaseId,
});

const normalizeRelease = (release: any): ConfigTemplateRelease => ({
  ...normalizeTemplate(release),
  id: release.id,
  templateId: release.template_id ?? release.templateId,
  version: release.version,
  contentSha256: release.content_sha256 ?? release.contentSha256,
  createBy: release.create_by ?? release.createBy,
});

const normalizeValues = (values: any): NamespaceTemplateValues => ({
  id: values.id,
  namespace: values.namespace,
  templateId: values.template_id ?? values.templateId,
  values: Object.fromEntries(
    Object.entries(values.values || {}).map(([key, value]) => [key, normalizeValue(value as ApiTemplateValue)])
  ),
  revision: values.revision,
  modifyBy: values.modify_by ?? values.modifyBy,
});

const normalizeValueRelease = (release: any): NamespaceTemplateValueRelease => ({
  ...normalizeValues(release),
  id: release.id,
  valuesId: release.values_id ?? release.valuesId,
  templateReleaseId: release.template_release_id ?? release.templateReleaseId,
  releaseType: release.release_type ?? release.releaseType,
  betaLabels: release.beta_labels ?? release.betaLabels ?? [],
  priority: release.priority,
  active: release.active,
  version: release.version,
  comment: release.comment,
  createBy: release.create_by ?? release.createBy,
});

export async function describeConfigTemplates() {
  const res = await getApiRequest<{ data?: ApiTemplate[]; amount?: number }>({
    action: BaseURL.CONFIG_TEMPLATE,
  });
  const templates = (res.data ?? []).map(normalizeTemplate);
  return { templates, amount: res.amount ?? templates.length };
}

export async function createConfigTemplate(template: ConfigFileTemplate) {
  return apiRequest({
    action: BaseURL.CONFIG_TEMPLATE,
    data: [toApiTemplate(template)],
  });
}

export async function updateConfigTemplate(template: ConfigFileTemplate) {
  return putApiRequest({
    action: BaseURL.CONFIG_TEMPLATE,
    data: [toApiTemplate(template)],
  });
}

export async function describeConfigTemplateReleases(templateId: string | number) {
  const res = await getApiRequest<{ data?: any[]; amount?: number }>({
    action: `${BaseURL.CONFIG_TEMPLATE}/releases`,
    data: { template_id: templateId },
  });
  const releases = (res.data ?? []).map(normalizeRelease);
  return { releases, amount: res.amount ?? releases.length };
}

export async function publishConfigTemplateRelease(template: ConfigFileTemplate) {
  const payload = {
    ...toApiTemplate(template),
    id: '',
    template_id: template.id,
    version: 0,
  };
  return apiRequest<ConfigTemplateRelease>({
    action: `${BaseURL.CONFIG_TEMPLATE}/releases`,
    data: payload,
  });
}

export async function describeNamespaceTemplateValues(namespace: string, templateId: string | number) {
  try {
    const res = await getApiRequest<any>({
      action: `${BaseURL.CONFIG_TEMPLATE}/values`,
      data: { namespace, template_id: templateId },
    });
    return res?.namespace ? normalizeValues(res) : undefined;
  } catch (error) {
    // 尚未创建草稿是合法空态，不应阻塞该 Namespace 首次录入 Value。
    if (error instanceof RequestError && error.code === 404202) return undefined;
    throw error;
  }
}

export async function saveNamespaceTemplateValues(values: NamespaceTemplateValues) {
  return putApiRequest({
    action: `${BaseURL.CONFIG_TEMPLATE}/values`,
    data: {
      id: values.id || '',
      namespace: values.namespace,
      template_id: values.templateId,
      values: Object.fromEntries(Object.entries(values.values).map(([key, value]) => [key, toApiValue(value)])),
      revision: values.revision || '',
    },
  });
}

export async function describeNamespaceTemplateValueReleases(namespace: string, templateId: string | number) {
  const res = await getApiRequest<{ data?: any[]; amount?: number }>({
    action: `${BaseURL.CONFIG_TEMPLATE}/environment-releases`,
    data: { namespace, template_id: templateId },
  });
  const releases = (res.data ?? []).map(normalizeValueRelease);
  return { releases, amount: res.amount ?? releases.length };
}

export async function publishEnvironmentConfigRelease(release: NamespaceTemplateValueRelease) {
  return apiRequest({
    action: `${BaseURL.CONFIG_TEMPLATE}/environment-releases`,
    data: {
      id: '',
      values_id: release.valuesId || '',
      namespace: release.namespace,
      template_id: release.templateId,
      values: Object.fromEntries(Object.entries(release.values).map(([key, value]) => [key, toApiValue(value)])),
      release_type: release.releaseType,
      beta_labels: release.betaLabels || [],
      priority: release.priority || 0,
      active: release.active ?? true,
      version: 0,
      comment: release.comment || '',
    },
  });
}

export async function previewConfigTemplate(
  template: ConfigFileTemplate,
  values: Record<string, ConfigTemplateValue>
): Promise<RenderPreview> {
  const res = await apiRequest<any>({
    action: `${BaseURL.CONFIG_TEMPLATE}/preview`,
    data: {
      input: {
        content: template.content,
        format: template.format,
        engine: template.engine || TemplateEngine,
        parameter_schema: template.parameterSchema.map(toApiSchema),
        values: Object.fromEntries(Object.entries(values).map(([key, value]) => [key, toApiValue(value)])),
      },
    },
  });
  return {
    code: res.code ?? 0,
    info: res.info ?? '',
    valid: Boolean(res.valid),
    renderedContent: res.rendered_content ?? res.renderedContent ?? '',
    format: res.format ?? template.format,
    renderedSha256: res.rendered_sha256 ?? res.renderedSha256 ?? '',
    templateReleaseId: res.template_release_id ?? res.templateReleaseId,
    valueReleaseId: res.value_release_id ?? res.valueReleaseId,
    engine: res.engine,
    diagnostics: res.diagnostics ?? [],
  };
}

export async function bindConfigFileTemplate(file: ConfigFile, binding: ConfigTemplateBinding) {
  return apiRequest({
    action: `${BaseURL.CONFIG_TEMPLATE}/bindings`,
    data: {
      name: file.name,
      namespace: file.namespace,
      group: file.group,
      content: file.content || '',
      comment: file.comment || '',
      format: file.format || 'text',
      labels: Object.fromEntries((file.tags || []).map((item) => [item.key, item.value])),
      encrypted: Boolean(file.encrypted),
      encrypt_algo: file.encryptAlgo || '',
      config_type: 'CONFIG_TEMPLATE',
      template_binding: {
        template_id: binding.templateId,
        template_release_id: binding.templateReleaseId,
        binding_release_id: binding.bindingReleaseId || '',
      },
    },
  });
}

export async function describeConfigTemplateBindings(namespace: string, group: string, fileName: string) {
  const res = await getApiRequest<{ data?: any[]; amount?: number }>({
    action: `${BaseURL.CONFIG_TEMPLATE}/bindings`,
    data: { namespace, group, file_name: fileName },
  });
  const bindings = (res.data ?? []).map(normalizeBinding);
  return { bindings, amount: res.amount ?? bindings.length };
}

import fs from 'node:fs';
import path from 'node:path';
import process from 'node:process';

const root = process.cwd();

const read = (file) => fs.readFileSync(path.join(root, file), 'utf8');

const checks = [
  {
    file: 'src/components/Fluent/DateRangePicker.tsx',
    expectations: [
      ['uses the Fluent Calendar compat component', /from '@fluentui\/react-calendar-compat'/],
      ['uses one popover-backed range trigger', /<Popover[\s\S]*fluent-date-range__trigger[\s\S]*<PopoverSurface/],
      ['supports a two-step start and end date selection', /type SelectionPhase = 'start' \| 'end'[\s\S]*setPhase\('end'\)/],
      ['marks both range endpoints in the calendar', /customDayCellRef[\s\S]*is-range-start[\s\S]*is-range-end/],
      ['keeps second precision for datetime values', /withSeconds[\s\S]*step=\{withSeconds \? 1 : undefined\}/],
      ['provides linked hour minute and second columns', /fluent-date-range__time-columns[\s\S]*part: 'hour'[\s\S]*part: 'minute'[\s\S]*part: 'second'/],
      ['normalizes values to the requested format', /formatDateValue\(date, format\)/],
      ['normalizes preset Date objects before emitting', /applyPreset[\s\S]*parseDateValue\(preset\[0\]\)[\s\S]*emit\(\[start, end\], 'preset'\)/],
      ['rejects reversed ranges', /draftEnd < draftStart[\s\S]*结束时间不能早于开始时间/],
      ['recalculates natural-day boundaries for reverse selection', /applyEndpointDefaultTime\(nextEnd, 'start'[\s\S]*applyEndpointDefaultTime\(start, 'end'/],
      ['keeps explicit confirmation semantics', /disabled=\{!completeRange\} onClick=\{confirm\}/],
      ['supports a clear action', /trigger: 'clear'/],
      ['localizes the calendar for Chinese UI', /CHINESE_CALENDAR_STRINGS[\s\S]*selectedDateFormatString: '已选择 \{0\}'/],
    ],
    forbidden: [
      ['must not restore TDesign', /tdesign/],
      ['must not render a standalone preset dropdown', /placeholder="快捷选择"|fluent-date-range__presets"[^>]*Dropdown/],
    ],
  },
  {
    file: 'src/components/Fluent/legacyWidgets.tsx',
    expectations: [
      ['reuses the shared range value contract for TimeRangePicker', /import type \{ DateRangePickerProps, DateRangeValue \} from '.\/DateRangePicker'/],
    ],
    forbidden: [
      ['must not keep the old pair of datetime-local inputs', /export const DateRangePicker[\s\S]*datetime-local/],
      ['must not keep the old standalone preset select', /fluent-date-range__presets[\s\S]*FluentSelect/],
    ],
  },
  {
    file: 'src/styles/fluent.less',
    expectations: [
      ['renders the range as one control', /\.fluent-date-range__trigger[\s\S]*width: 100%/],
      ['provides a calendar and time panel layout', /\.fluent-date-range__panel[\s\S]*grid-template-columns[\s\S]*&\.with-time/],
      ['keeps the picker usable on narrow screens', /@media \(max-width: 620px\)[\s\S]*\.fluent-date-range__panel\.with-time/],
    ],
    forbidden: [
      ['must not keep wrapping three independent controls', /\.fluent-date-range[\s\S]{[^}]*flex-wrap:\s*wrap/],
    ],
  },
  {
    file: 'src/pages/Metrics/ServerOperation/index.tsx',
    expectations: [
      ['keeps second-precision operation audit filters', /<DateRangePicker[\s\S]*format="YYYY-MM-DD HH:mm:ss"[\s\S]*presets=\{presets\}/],
    ],
    forbidden: [],
  },
  {
    file: 'src/pages/Metrics/ServerEvent/index.tsx',
    expectations: [
      ['keeps second-precision event filters', /<DateRangePicker[\s\S]*format="YYYY-MM-DD HH:mm:ss"[\s\S]*presets=\{presets\}/],
    ],
    forbidden: [],
  },
];

const errors = [];
for (const check of checks) {
  const source = read(check.file);
  for (const [label, pattern] of check.expectations) {
    if (!pattern.test(source)) errors.push(`${check.file}: 缺少 ${label}`);
  }
  for (const [label, pattern] of check.forbidden) {
    if (pattern.test(source)) errors.push(`${check.file}: 仍包含 ${label}`);
  }
}

const packageJson = JSON.parse(read('package.json'));
if (packageJson.dependencies?.['@fluentui/react-calendar-compat'] !== '^0.4.4') {
  errors.push('package.json: Fluent Calendar compat 依赖未固定为 ^0.4.4');
}
if (packageJson.devDependencies?.typescript !== '5.1.6') {
  errors.push('package.json: TypeScript 未固定为当前 ESLint 工具链支持的 5.1.6');
}

if (errors.length > 0) {
  console.error(errors.map((error) => `- ${error}`).join('\n'));
  process.exit(1);
}

console.log('DateRangePicker 专项契约检查通过');

import React from 'react';
import {
  Button,
  Input,
  Popover,
  PopoverSurface,
  PopoverTrigger,
} from '@fluentui/react-components';
import {
  Calendar,
  DateRangeType,
  DayOfWeek,
  defaultCalendarStrings,
} from '@fluentui/react-calendar-compat';
import { CalendarLtr20Regular, Dismiss16Regular } from '@fluentui/react-icons';

type AnyRecord = Record<string, unknown>;
type RangeEndpoint = string | Date;
type SelectionPhase = 'start' | 'end';
type TimePart = 'hour' | 'minute' | 'second';

export type DateRangeValue = RangeEndpoint[];

export interface DateRangePickerProps extends Omit<React.HTMLAttributes<HTMLDivElement>, 'defaultValue' | 'onChange'> {
  value?: DateRangeValue;
  defaultValue?: DateRangeValue;
  onChange?: (value: DateRangeValue, context?: AnyRecord) => void;
  placeholder?: [string, string];
  format?: string;
  valueType?: string;
  mode?: 'date';
  presets?: Record<string, DateRangeValue>;
  clearable?: boolean;
  allowInput?: boolean;
  disabled?: boolean;
}

const CHINESE_CALENDAR_STRINGS = {
  ...defaultCalendarStrings,
  months: ['一月', '二月', '三月', '四月', '五月', '六月', '七月', '八月', '九月', '十月', '十一月', '十二月'],
  shortMonths: ['1 月', '2 月', '3 月', '4 月', '5 月', '6 月', '7 月', '8 月', '9 月', '10 月', '11 月', '12 月'],
  days: ['星期日', '星期一', '星期二', '星期三', '星期四', '星期五', '星期六'],
  shortDays: ['日', '一', '二', '三', '四', '五', '六'],
  goToToday: '今天',
  prevMonthAriaLabel: '上个月',
  nextMonthAriaLabel: '下个月',
  prevYearAriaLabel: '上一年',
  nextYearAriaLabel: '下一年',
  prevYearRangeAriaLabel: '上一个年份范围',
  nextYearRangeAriaLabel: '下一个年份范围',
  monthPickerHeaderAriaLabel: '{0}，选择年份',
  yearPickerHeaderAriaLabel: '{0}，选择月份',
  selectedDateFormatString: '已选择 {0}',
  todayDateFormatString: '今天 {0}',
  dayMarkedAriaLabel: '范围内日期',
};

const pad = (value: number) => String(value).padStart(2, '0');
const HOURS = Array.from({ length: 24 }, (_, index) => index);
const MINUTES_AND_SECONDS = Array.from({ length: 60 }, (_, index) => index);

const parseDateValue = (value?: RangeEndpoint): Date | undefined => {
  if (!value) return undefined;
  if (value instanceof Date) {
    return Number.isNaN(value.getTime()) ? undefined : new Date(value.getTime());
  }

  const match = String(value).trim().match(
    /^(\d{4})-(\d{2})-(\d{2})(?:[T\s](\d{2}):(\d{2})(?::(\d{2}))?)?$/,
  );
  if (match) {
    const [, year, month, day, hour = '0', minute = '0', second = '0'] = match;
    const date = new Date(
      Number(year),
      Number(month) - 1,
      Number(day),
      Number(hour),
      Number(minute),
      Number(second),
    );
    return Number.isNaN(date.getTime()) ? undefined : date;
  }

  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? undefined : date;
};

const formatDateValue = (date: Date, format = 'YYYY-MM-DD') => {
  const datePart = `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}`;
  if (!format.includes('HH')) return datePart;

  const timePart = `${pad(date.getHours())}:${pad(date.getMinutes())}`;
  return format.includes('ss') ? `${datePart} ${timePart}:${pad(date.getSeconds())}` : `${datePart} ${timePart}`;
};

const toInputValue = (value: RangeEndpoint | undefined, withTime: boolean, withSeconds: boolean) => {
  const date = parseDateValue(value);
  if (!date) return '';

  const datePart = formatDateValue(date);
  if (!withTime) return datePart;

  const timePart = `${pad(date.getHours())}:${pad(date.getMinutes())}`;
  return `${datePart}T${timePart}${withSeconds ? `:${pad(date.getSeconds())}` : ''}`;
};

const normalizeValue = (
  value: DateRangeValue | undefined,
  format: string,
): [string, string] => [
  value?.[0] && parseDateValue(value[0]) ? formatDateValue(parseDateValue(value[0])!, format) : '',
  value?.[1] && parseDateValue(value[1]) ? formatDateValue(parseDateValue(value[1])!, format) : '',
];

const mergeCalendarDate = (
  current: RangeEndpoint | undefined,
  picked: Date,
  endpoint: SelectionPhase,
  withTime: boolean,
) => {
  const next = parseDateValue(current) || new Date(picked);
  next.setFullYear(picked.getFullYear(), picked.getMonth(), picked.getDate());

  if (!current && withTime) {
    next.setHours(endpoint === 'start' ? 0 : 23, endpoint === 'start' ? 0 : 59, endpoint === 'start' ? 0 : 59, 0);
  } else if (!withTime) {
    next.setHours(0, 0, 0, 0);
  }
  return next;
};

const applyEndpointDefaultTime = (
  date: Date,
  endpoint: SelectionPhase,
  withTime: boolean,
) => {
  const next = new Date(date);
  next.setHours(
    withTime && endpoint === 'end' ? 23 : 0,
    withTime && endpoint === 'end' ? 59 : 0,
    withTime && endpoint === 'end' ? 59 : 0,
    0,
  );
  return next;
};

const inclusiveCalendarDays = (start?: Date, end?: Date) => {
  if (!start || !end || end < start) return 1;
  const startDay = new Date(start.getFullYear(), start.getMonth(), start.getDate()).getTime();
  const endDay = new Date(end.getFullYear(), end.getMonth(), end.getDate()).getTime();
  return Math.floor((endDay - startDay) / 86400000) + 1;
};

const isSameCalendarDate = (left: Date, right?: Date) => Boolean(
  right
  && left.getFullYear() === right.getFullYear()
  && left.getMonth() === right.getMonth()
  && left.getDate() === right.getDate(),
);

const isDateValueType = (valueType?: string) => valueType?.toLowerCase() === 'date';

export const DateRangePicker: React.FC<DateRangePickerProps> = ({
  value,
  defaultValue,
  onChange,
  placeholder = ['开始日期', '结束日期'],
  format = 'YYYY-MM-DD',
  presets,
  clearable,
  disabled,
  allowInput = true,
  valueType,
  mode: _mode,
  className,
  style,
  ...props
}) => {
  const withTime = format.includes('HH');
  const withSeconds = format.includes('ss');
  const [inner, setInner] = React.useState<DateRangeValue>(defaultValue || []);
  const selected = value ?? inner;
  const normalizedSelected = React.useMemo(
    () => normalizeValue(selected, format),
    [format, selected],
  );
  const [draft, setDraft] = React.useState<[string, string]>(normalizedSelected);
  const [open, setOpen] = React.useState(false);
  const [phase, setPhase] = React.useState<SelectionPhase>('start');

  React.useEffect(() => {
    if (!open) setDraft(normalizedSelected);
  }, [normalizedSelected, open]);

  const draftStart = parseDateValue(draft[0]);
  const draftEnd = parseDateValue(draft[1]);
  const activeTime = phase === 'start' ? draftStart : draftEnd;
  const completeRange = Boolean(draftStart && draftEnd && draftStart <= draftEnd);
  const timeColumnsRef = React.useRef<HTMLDivElement>(null);

  React.useLayoutEffect(() => {
    if (!open || !withTime || !activeTime) return;
    const selectedButtons = timeColumnsRef.current?.querySelectorAll<HTMLElement>('.is-selected') || [];
    selectedButtons.forEach((button) => {
      const column = button.parentElement;
      if (!column) return;
      const buttonTop = button.getBoundingClientRect().top - column.getBoundingClientRect().top + column.scrollTop;
      column.scrollTop = buttonTop - (column.clientHeight / 2) + (button.clientHeight / 2);
    });
  }, [activeTime, open, phase, withTime]);

  const emit = (nextDates: [Date, Date], trigger: 'confirm' | 'preset') => {
    const next: DateRangeValue = isDateValueType(valueType)
      ? nextDates
      : nextDates.map((date) => formatDateValue(date, format));
    if (value === undefined) setInner(next);
    onChange?.(next, { dayjsValue: nextDates, trigger });
  };

  const clear = (event?: React.MouseEvent) => {
    event?.stopPropagation();
    if (value === undefined) setInner([]);
    setDraft(['', '']);
    setOpen(false);
    onChange?.([], { dayjsValue: [], trigger: 'clear' });
  };

  const selectCalendarDate = (picked: Date) => {
    if (phase === 'start') {
      const nextStart = mergeCalendarDate(draft[0], picked, 'start', withTime);
      setDraft([formatDateValue(nextStart, format), '']);
      setPhase('end');
      return;
    }

    const nextEnd = mergeCalendarDate(draft[1], picked, 'end', withTime);
    const start = parseDateValue(draft[0]);
    if (start && nextEnd < start) {
      setDraft([
        formatDateValue(applyEndpointDefaultTime(nextEnd, 'start', withTime), format),
        formatDateValue(applyEndpointDefaultTime(start, 'end', withTime), format),
      ]);
    } else {
      setDraft([draft[0], formatDateValue(nextEnd, format)]);
    }
    setPhase('start');
  };

  const updateDraftInput = (index: 0 | 1, nextValue: string) => {
    const next = [...draft] as [string, string];
    const parsed = parseDateValue(nextValue);
    next[index] = parsed ? formatDateValue(parsed, format) : '';
    setDraft(next);
    setPhase(index === 0 ? 'start' : 'end');
  };

  const updateTimePart = (part: TimePart, nextValue: number) => {
    if (!activeTime) return;
    const next = new Date(activeTime);
    if (part === 'hour') next.setHours(nextValue);
    if (part === 'minute') next.setMinutes(nextValue);
    if (part === 'second') next.setSeconds(nextValue);

    const nextDraft = [...draft] as [string, string];
    nextDraft[phase === 'start' ? 0 : 1] = formatDateValue(next, format);
    setDraft(nextDraft);
  };

  const applyPreset = (preset: DateRangeValue) => {
    const start = parseDateValue(preset[0]);
    const end = parseDateValue(preset[1]);
    if (!start || !end || end < start) return;
    setDraft([formatDateValue(start, format), formatDateValue(end, format)]);
    emit([start, end], 'preset');
    setOpen(false);
  };

  const confirm = () => {
    if (!draftStart || !draftEnd || draftEnd < draftStart) return;
    emit([draftStart, draftEnd], 'confirm');
    setOpen(false);
  };

  const hasValue = normalizedSelected.some(Boolean);
  const displayStart = normalizedSelected[0] || placeholder[0];
  const displayEnd = normalizedSelected[1] || placeholder[1];

  return (
    <div
      {...props}
      className={`fluent-date-range ${hasValue ? 'has-value' : ''} ${className || ''}`.trim()}
      style={style}
    >
      <Popover
        open={open}
        onOpenChange={(_event, data) => {
          setOpen(data.open);
          if (data.open) {
            setDraft(normalizedSelected);
            setPhase('start');
          }
        }}
        positioning={{ position: 'below', align: 'start' }}
        trapFocus
      >
        <PopoverTrigger disableButtonEnhancement>
          <Button
            type="button"
            appearance="outline"
            className="fluent-date-range__trigger"
            aria-label={`${displayStart} 至 ${displayEnd}`}
            aria-expanded={open}
            disabled={disabled}
          >
            <span className={`fluent-date-range__value ${normalizedSelected[0] ? '' : 'is-placeholder'}`.trim()}>
              {displayStart}
            </span>
            <span className="fluent-date-range__separator">—</span>
            <span className={`fluent-date-range__value ${normalizedSelected[1] ? '' : 'is-placeholder'}`.trim()}>
              {displayEnd}
            </span>
            <CalendarLtr20Regular className="fluent-date-range__calendar-icon" aria-hidden />
          </Button>
        </PopoverTrigger>

        <PopoverSurface className="fluent-date-range__popover" aria-label="选择日期范围">
          <div className={`fluent-date-range__panel ${withTime ? 'with-time' : ''}`.trim()}>
            <section
              className={`fluent-date-range__calendar ${draftStart && draftEnd ? 'has-range' : ''}`.trim()}
              aria-label={phase === 'start' ? '选择开始日期' : '选择结束日期'}
            >
              <div className="fluent-date-range__phase">
                {phase === 'start' ? '选择开始日期' : '选择结束日期'}
              </div>
              <Calendar
                key={`${draft[0]}|${draft[1]}`}
                value={draftStart || draftEnd || new Date()}
                onSelectDate={selectCalendarDate}
                dateRangeType={DateRangeType.Day}
                firstDayOfWeek={DayOfWeek.Monday}
                strings={CHINESE_CALENDAR_STRINGS}
                isMonthPickerVisible={false}
                showMonthPickerAsOverlay
                showGoToToday={false}
                calendarDayProps={{
                  daysToSelectInDayView: inclusiveCalendarDays(draftStart, draftEnd),
                  customDayCellRef: (element, date) => {
                    if (!element) return;
                    element.classList.toggle('is-range-start', isSameCalendarDate(date, draftStart));
                    element.classList.toggle('is-range-end', isSameCalendarDate(date, draftEnd));
                  },
                }}
              />
            </section>

            <section className="fluent-date-range__inputs" aria-label={withTime ? '日期和时间' : '日期'}>
              <div className="fluent-date-range__manual-inputs">
                <label>
                  <span>{placeholder[0]}</span>
                  <Input
                    aria-label={placeholder[0]}
                    type={withTime ? 'datetime-local' : 'date'}
                    step={withSeconds ? 1 : undefined}
                    value={toInputValue(draft[0], withTime, withSeconds)}
                    readOnly={!allowInput}
                    onFocus={() => setPhase('start')}
                    onChange={(_event, data) => updateDraftInput(0, data.value)}
                  />
                </label>
                <label>
                  <span>{placeholder[1]}</span>
                  <Input
                    aria-label={placeholder[1]}
                    type={withTime ? 'datetime-local' : 'date'}
                    step={withSeconds ? 1 : undefined}
                    value={toInputValue(draft[1], withTime, withSeconds)}
                    readOnly={!allowInput}
                    onFocus={() => setPhase('end')}
                    onChange={(_event, data) => updateDraftInput(1, data.value)}
                  />
                </label>
              </div>

              <div className="fluent-date-range__endpoint-tabs" aria-label="选择要编辑的时间">
                {(['start', 'end'] as SelectionPhase[]).map((endpoint, index) => {
                  const endpointDate = endpoint === 'start' ? draftStart : draftEnd;
                  return (
                    <Button
                      key={endpoint}
                      appearance="subtle"
                      className={phase === endpoint ? 'is-active' : ''}
                      disabled={!endpointDate}
                      aria-pressed={phase === endpoint}
                      onClick={() => setPhase(endpoint)}
                    >
                      <span>{index === 0 ? '开始' : '结束'}</span>
                      <strong>{endpointDate ? formatDateValue(endpointDate, 'YYYY-MM-DD HH:mm:ss').slice(11) : '--:--:--'}</strong>
                    </Button>
                  );
                })}
              </div>

              <div className="fluent-date-range__time-heading">
                <span>{phase === 'start' ? '开始时间' : '结束时间'}</span>
                <strong>{activeTime ? formatDateValue(activeTime, 'YYYY-MM-DD HH:mm:ss').slice(11) : '--:--:--'}</strong>
              </div>

              <div className="fluent-date-range__time-columns" ref={timeColumnsRef}>
                {[
                  { label: '时', part: 'hour' as const, values: HOURS, selected: activeTime?.getHours() },
                  { label: '分', part: 'minute' as const, values: MINUTES_AND_SECONDS, selected: activeTime?.getMinutes() },
                  ...(withSeconds
                    ? [{ label: '秒', part: 'second' as const, values: MINUTES_AND_SECONDS, selected: activeTime?.getSeconds() }]
                    : []),
                ].map((column) => (
                  <div
                    key={column.part}
                    className="fluent-date-range__time-column"
                    role="listbox"
                    aria-label={column.label}
                  >
                    {column.values.map((item) => (
                      <Button
                        key={item}
                        appearance="subtle"
                        className={item === column.selected ? 'is-selected' : ''}
                        disabled={!activeTime}
                        role="option"
                        aria-selected={item === column.selected}
                        onClick={() => updateTimePart(column.part, item)}
                      >
                        {pad(item)}
                      </Button>
                    ))}
                  </div>
                ))}
              </div>

              {draftStart && draftEnd && draftEnd < draftStart && (
                <span className="fluent-date-range__error" role="alert">结束时间不能早于开始时间</span>
              )}
            </section>
          </div>

          <footer className="fluent-date-range__footer">
            <div className="fluent-date-range__presets" aria-label="快捷时间范围">
              {Object.entries(presets || {}).map(([label, preset]) => (
                <Button key={label} appearance="subtle" onClick={() => applyPreset(preset)}>
                  {label}
                </Button>
              ))}
            </div>
            <Button appearance="primary" disabled={!completeRange} onClick={confirm}>
              确定
            </Button>
          </footer>
        </PopoverSurface>
      </Popover>

      {clearable && hasValue && !disabled && (
        <Button
          className="fluent-date-range__clear"
          appearance="subtle"
          size="small"
          icon={<Dismiss16Regular />}
          aria-label="清空日期范围"
          title="清空日期范围"
          onClick={clear}
        />
      )}
    </div>
  );
};

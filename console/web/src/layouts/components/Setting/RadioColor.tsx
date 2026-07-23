import React, { useState } from 'react';
import {
  Button,
  ColorSwatch,
  Field,
  Popover,
  PopoverSurface,
  PopoverTrigger,
  SwatchPicker,
} from '@fluentui/react-components';
import { defaultColor } from 'configs/color';
import { ColorPickerPanel } from 'components/Fluent';
import Style from './RadioColor.module.less';

interface IProps {
  value: string;
  onChange: (color: string) => void;
}

const RadioColor = (props: IProps) => {
  const [isColorPickerOpen, setIsColorPickerOpen] = useState(false);
  const normalizedValue = props.value.toLowerCase();
  const isCustomColor = !defaultColor.includes(normalizedValue);

  return (
    <Field label='主题色'>
      <div className={Style.colorControls}>
        <SwatchPicker
          layout='row'
          shape='circular'
          size='medium'
          selectedValue={normalizedValue}
          onSelectionChange={(_event, data) => props.onChange(data.selectedValue)}
        >
          {defaultColor.map((color) => (
            <ColorSwatch key={color} color={color} value={color} aria-label={`切换主题色为 ${color}`} />
          ))}
        </SwatchPicker>

        <Popover
          open={isColorPickerOpen}
          onOpenChange={(_event, data) => setIsColorPickerOpen(data.open)}
          positioning='below-end'
        >
          <PopoverTrigger disableButtonEnhancement>
            <Button appearance={isCustomColor ? 'primary' : 'secondary'} aria-pressed={isCustomColor}>
              {isCustomColor ? `自定义颜色 ${props.value.toUpperCase()}` : '自定义颜色'}
            </Button>
          </PopoverTrigger>
          <PopoverSurface className={Style.colorPickerSurface}>
            <ColorPickerPanel
              value={props.value}
              onChange={props.onChange}
              colorModes={['monochrome']}
              format='HEX'
              swatchColors={[]}
            />
          </PopoverSurface>
        </Popover>
      </div>
    </Field>
  );
};

export default React.memo(RadioColor);

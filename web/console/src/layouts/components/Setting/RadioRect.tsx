import React, { memo } from 'react';
import { Field, Radio, RadioGroup } from '@fluentui/react-components';

interface IOption {
  value: string;
  name: string;
}

interface IProps {
  value: string;
  onChange: (value: string) => void;
  options: IOption[];
}

export default memo((props: IProps) => {
  return (
    <Field label='主题模式'>
      <RadioGroup
        layout='horizontal'
        value={props.value}
        onChange={(_event, data) => props.onChange(data.value)}
      >
        {props.options.map((item) => (
          <Radio key={item.value} value={item.value} label={item.name} />
        ))}
      </RadioGroup>
    </Field>
  );
});

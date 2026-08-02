import React from 'react';
import { Select } from '@radix-ui/themes';
import './MobileSelector.css';
import { TabViews } from '../../organisms/Statistics/NetworkStats/NetworkStats';

interface MobileSelectorProps {
  options: { label: string; value: TabViews }[];
  setMobileActiveTab: (value: TabViews) => void;
  defaultValue: TabViews;
}

const MobileSelector: React.FC<MobileSelectorProps> = ({ defaultValue, options, setMobileActiveTab }) => {
  return (
    <Select.Root defaultValue={defaultValue} onValueChange={setMobileActiveTab}>
      <Select.Trigger className="mobile-symbol-select-trigger" variant="ghost" />
      <Select.Content className="mobile-selector-content" position="popper">
        {options.map((opt) => (
          <Select.Item key={opt.value} className="mobile-selector-item" value={opt.value}>
            {opt.label}
          </Select.Item>
        ))}
      </Select.Content>
    </Select.Root>
  );
};

export default MobileSelector;

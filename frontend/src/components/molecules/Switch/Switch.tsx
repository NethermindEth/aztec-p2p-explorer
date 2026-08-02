import { Switch } from '@radix-ui/themes';
import './Switch.css';

const SwitchToggle = ({ is3D, toggleView }: { is3D: boolean; toggleView: () => void }) => {
  return (
    <div className="switch-container">
      <Switch checked={is3D} onCheckedChange={toggleView} color="yellow" size={'1'} />
      3D Map
    </div>
  );
};

export default SwitchToggle;

import * as Tabs from '@radix-ui/react-tabs';
import { useAtom } from 'jotai';
import { selectedNodeAtom, selectedRowAtom } from '../../../../hooks/atoms';
import { tabViewOptions, TabViews } from './NetworkStats';

const RenderTabTitles = () => {
  const [, setSelectedNode] = useAtom(selectedNodeAtom);
  const [, setSelectedRow] = useAtom(selectedRowAtom);

  return (
    <>
      {tabViewOptions.map((tab) => (
        <Tabs.Trigger
          key={tab.value}
          value={tab.value}
          className={'network-stats__tabs-trigger'}
          onClick={() => {
            if (tab.value === TabViews.NodeList) {
              setSelectedNode(null);
              setSelectedRow(null);
            }
          }}
        >
          {tab.label}
        </Tabs.Trigger>
      ))}
    </>
  );
};

export default RenderTabTitles;

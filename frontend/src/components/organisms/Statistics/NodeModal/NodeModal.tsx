import { Button, Dialog } from '@radix-ui/themes';
import DetailsTable from '../../../molecules/PanelNodeDetails/DetailsTable/DetailsTable';
import { useAtom } from 'jotai';
import { nodeModalAtom } from '../../../../hooks/atoms';
import '../../../organisms/PanelNodeDetails/PanelNodeDetails.css';
import '../../../molecules/PanelNodeDetails/NodeStatsBlock/NodeStatsBlock.css';
import './NodeModal.css';
import { Cross2Icon } from '@radix-ui/react-icons';
import { usePeerId } from '../../../../api/peers';
import { useSearchParams } from 'react-router-dom';
import { useEffect, useState } from 'react';

const NodeModal = () => {
  const [modalData, setModalAtom] = useAtom(nodeModalAtom);
  const { data: node } = usePeerId(modalData.nodeId);
  const [searchParams, setSearchParams] = useSearchParams();
  const id = searchParams.get('id');
  const [open, setOpen] = useState(Boolean(id));

  const closeModal = () => {
    const next = new URLSearchParams(searchParams);
    next.delete('id');
    setSearchParams(next, { replace: true });
    setModalAtom({ nodeId: '', node: null, isOpen: false });
  };

  useEffect(() => {
    if (id) {
      setOpen(true);
    } else {
      setOpen(false);
    }
  }, [id]);

  return (
    <Dialog.Root open={open}>
      <Dialog.Content className="panel-node-details-container modal-container">
        <div className="node-stats-block modal-block">
          <h3 className="modal-title">Node Stats</h3>
          <Button variant="outline" color="gray" className="modal-btn" onClick={closeModal}>
            <Cross2Icon />
          </Button>
        </div>
        <DetailsTable node={node ?? null} />
      </Dialog.Content>
    </Dialog.Root>
  );
};

export default NodeModal;

import React from 'react';
import './RowItem.css';
import Skeleton from '../../molecules/Skeleton/Skeleton';

interface RowItemProps {
  title: string;
  content: React.ReactNode;
}

const RowItem: React.FC<RowItemProps> = ({ title, content }) => {
  return (
    <div className="row-item">
      <div className="title">{title}</div>
      <div className="content">{content ? content : <Skeleton width={'100%'} height={'18px'} />}</div>
    </div>
  );
};

export default RowItem;

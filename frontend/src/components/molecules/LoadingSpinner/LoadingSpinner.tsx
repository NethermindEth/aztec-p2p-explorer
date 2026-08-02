import './LoadingSpinner.css';

type LoadingSpinnerProps = {
  size?: 'sm' | 'md' | 'lg';
  className?: string;
  style?: React.CSSProperties;
};

const sizeMap = {
  sm: { width: '24px', height: '24px', borderWidth: '3px' },
  md: { width: '48px', height: '48px', borderWidth: '4px' },
  lg: { width: '72px', height: '72px', borderWidth: '6px' },
};

const LoadingSpinner = ({ size = 'md', className = '', style = {} }: LoadingSpinnerProps) => {
  const { width, height, borderWidth } = sizeMap[size];
  return (
    <div className={`spinner-container ${className}`} style={style}>
      <div
        className="spinner"
        style={{
          width,
          height,
          borderWidth,
          borderStyle: 'solid',
          borderColor: '#a9a9a9',
          borderTopColor: '#d4ff28',
        }}
      />
    </div>
  );
};

export default LoadingSpinner;

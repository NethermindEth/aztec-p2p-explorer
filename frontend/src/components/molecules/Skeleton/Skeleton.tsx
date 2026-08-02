import { Skeleton as RadixSkeleton, SkeletonProps as RadixSkeletonProps } from '@radix-ui/themes';

type SkeletonProps = RadixSkeletonProps & {
  radius?: string;
};
const Skeleton: React.FC<SkeletonProps> = ({ children, radius, ...rest }) => {
  return (
    <RadixSkeleton {...rest} style={{ borderRadius: radius || '4px' }}>
      {children}
    </RadixSkeleton>
  );
};

export default Skeleton;

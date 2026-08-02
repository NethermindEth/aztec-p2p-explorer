import React, { useRef, useEffect } from 'react';
import { Text } from '@react-three/drei';
import { applyTextRotation } from '../../../utils';
import * as THREE from 'three';
import { RawMarker } from '../../../utils/clusterMarkers';

const TextComponent: React.FC<{ data: RawMarker[]; count: number }> = ({ data, count }) => {
  const textRef = useRef<THREE.Object3D>(null);

  useEffect(() => {
    if (textRef.current) {
      applyTextRotation(textRef.current);
    }
  }, [textRef, data]);

  return (
    <Text
      ref={textRef}
      position={[0, 0 * Math.min(count, 3), 0.02]}
      fontSize={0.085}
      color="black"
      anchorX="center"
      anchorY="middle"
    >
      {count === 1 ? '' : count > 999 ? '999+' : count.toString()}
    </Text>
  );
};
export default TextComponent;

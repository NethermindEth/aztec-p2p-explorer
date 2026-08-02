import React, { useMemo, useRef, useEffect, useState } from 'react';
import * as THREE from 'three';
import { animated } from '@react-spring/three';
import { createArc } from '../../../utils';
import { useFrame, useThree } from '@react-three/fiber';

type LatLon = {
  lat: number;
  lon: number;
};

type AnimatedProps = {
  opacity: number;
  animatedValue: { get: () => number };
};

type AnimatedLineProps = {
  from: LatLon;
  to: LatLon;
  animatedProps: AnimatedProps;
};

const AnimatedLine: React.FC<AnimatedLineProps> = ({ from, to, animatedProps }) => {
  const ref = useRef<any>(null!);
  const [drawRange, setDrawRange] = useState(0);
  const { invalidate } = useThree();

  const arcPoints = useMemo(() => createArc(from.lat, from.lon, to.lat, to.lon, 3, 100), [from, to]);
  const positions = useMemo(() => {
    const pos = new Float32Array(arcPoints.length * 3);
    arcPoints.forEach((point, i) => {
      pos[i * 3] = point.x;
      pos[i * 3 + 1] = point.y;
      pos[i * 3 + 2] = point.z;
    });
    return pos;
  }, [arcPoints]);

  useFrame(() => {
    const numPoints = positions.length / 3;
    const drawRangeValue = Math.floor(numPoints * animatedProps.animatedValue.get());
    setDrawRange(drawRangeValue);
    invalidate(); // Request frame during animation
  });

  const geometry = useMemo(() => {
    const geo = new THREE.BufferGeometry();
    geo.setAttribute('position', new THREE.BufferAttribute(positions, 3));
    return geo;
  }, [positions]);

  useEffect(() => {
    if (ref.current) {
      ref.current.geometry.setDrawRange(0, drawRange);
    }
  }, [drawRange]);

  return (
    // @ts-ignore
    <animated.line ref={ref} geometry={geometry as any} material-opacity={animatedProps.opacity}>
      <lineBasicMaterial attach="material" color="#36ffeb" transparent />
    </animated.line>
  );
};

export default AnimatedLine;

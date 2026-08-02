import React, { useMemo } from 'react';
import * as THREE from 'three';
import { convertLatLongToVector3, applyRotation } from '../../../utils';
import TextComponent from './Text';
import clusterMarkers, { RawMarker } from '../../../utils/clusterMarkers';

type Props = {
  setHoveredMarker: React.Dispatch<React.SetStateAction<{ data: RawMarker[] | null; x: number; y: number } | null>>;
  onPointerOut: () => void;
  markersData: RawMarker[];
};

const Markers: React.FC<Props> = ({ setHoveredMarker, onPointerOut, markersData }) => {
  const clusteredMarkers = useMemo(() => {
    const validPeers = markersData.filter(
      (p) => p.latitude != null && p.longitude != null && !(p.latitude === 0 && p.longitude === 0)
    );

    const clusters = clusterMarkers(validPeers, 3.1, 1); // Adjust epsilon if needed

    return clusters.map((cluster) => {
      const position = convertLatLongToVector3(cluster.latitude, cluster.longitude);
      return { position, data: cluster.data, count: cluster.count };
    });
  }, [markersData]);

  return (
    <group>
      {clusteredMarkers.map((marker, index) => {
        const multiplier = 1.0;
        let scaleFactor = marker.count;
        if (scaleFactor > 3) {
          scaleFactor = 3;
        }

        const highCountBump = marker.count > 999 ? 1.1 : 1.0;
        const circleRadius = 0.032 * scaleFactor * multiplier * highCountBump;
        const invisibleCircleRadius = 0.035 * scaleFactor * multiplier * highCountBump;

        return (
          <group
            key={index}
            ref={(group) => group && applyRotation(group, marker.data[0].latitude, marker.data[0].longitude)}
          >
            <TextComponent data={marker.data} count={marker.count} />
            <mesh position={[0, 0, 0.01]}>
              <circleGeometry args={[circleRadius, 16]} />
              <meshBasicMaterial color="#8E8C99" side={THREE.DoubleSide} />
            </mesh>
            <mesh
              position={[0, 0, 0]}
              onPointerOver={(e) => {
                e.stopPropagation();
                setHoveredMarker({ data: marker.data, x: e.clientX, y: e.clientY });
              }}
              onPointerOut={onPointerOut}
            >
              <circleGeometry args={[invisibleCircleRadius, 16]} />
              <meshBasicMaterial visible={false} />
            </mesh>
          </group>
        );
      })}
    </group>
  );
};

export default Markers;

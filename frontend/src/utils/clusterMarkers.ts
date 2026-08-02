import * as clustering from 'density-clustering';

export type RawMarker = {
  latitude: number;
  longitude: number;
  count: number;
  continent: string;
  country: string;
  city: string;
  peer_ids: string[];
};

export type ClusteredMarker = {
  latitude: number;
  longitude: number;
  data: RawMarker[];
  count: number; // total peer_ids length across all data[]
};

const clusterMarkers = (markers: RawMarker[], epsilon: number, minPts: number): ClusteredMarker[] => {
  const dbscan = new clustering.DBSCAN();

  const coords = markers.map((m) => [m.latitude, m.longitude]);

  const clusters = dbscan.run(coords, epsilon, minPts);

  const clusteredMarkers: ClusteredMarker[] = clusters.map((indices) => {
    const clusterPoints = indices.map((i) => markers[i]);

    const latSum = clusterPoints.reduce((sum, p) => sum + p.latitude, 0);
    const lonSum = clusterPoints.reduce((sum, p) => sum + p.longitude, 0);

    const totalPeerCount = clusterPoints.reduce((sum, m) => sum + m.count, 0);

    return {
      latitude: latSum / clusterPoints.length,
      longitude: lonSum / clusterPoints.length,
      data: clusterPoints,
      count: totalPeerCount,
    };
  });

  return clusteredMarkers;
};

export default clusterMarkers;
